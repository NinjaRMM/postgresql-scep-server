package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/micromdm/scep/v2/depot"
	"math/big"
)

type SQLDepot struct {
	database  *sql.DB
	context context.Context
	certificate *x509.Certificate
	rsaKey *rsa.PrivateKey
}

func SqlDepot(conn string) (*SQLDepot, error) {
	database, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	err = database.Ping()
	if err != nil {
		return nil, err
	}
	return &SQLDepot{
		database:  database,
		context: context.Background(),
	}, nil
}

func (d *SQLDepot) loadCA(pass []byte) (*x509.Certificate, *rsa.PrivateKey, error) {
	var pemCert, pemKey []byte
	err := d.database.QueryRowContext(
		d.context, `
SELECT
    certificate_pem, key_pem
FROM
    certificates INNER JOIN ca_keys
        ON certificates.id = ca_keys.certificate_id
WHERE
    certificates.id = 1;`,
	).Scan(&pemCert, &pemKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	block, _ := pem.Decode(pemCert)
	if block.Type != "CERTIFICATE" {
		return nil, nil, errors.New("PEM block not a certificate")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	block, _ = pem.Decode(pemKey)
	if !x509.IsEncryptedPEMBlock(block) {
		return nil, nil, errors.New("PEM block not encrypted")
	}
	keyBytes, err := x509.DecryptPEMBlock(block, pass)
	if err != nil {
		return nil, nil, err
	}
	rsaKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
	if err != nil {
		return nil, nil, err
	}
	return certificate, rsaKey, nil
}

func (d *SQLDepot) createCA(pass []byte, years int, cn, org, country string) (*x509.Certificate, *rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	caCert := depot.NewCACert(
		depot.WithYears(years),
		depot.WithOrganization(org),
		depot.WithCommonName(cn),
		depot.WithCountry(country),
	)
	certificateBytes, err := caCert.SelfSign(rand.Reader, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, nil, err
	}
	certificate, err := x509.ParseCertificate(certificateBytes)
	if err != nil {
		return nil, nil, err
	}
	err = d.Put(certificate.Subject.CommonName, certificate)
	if err != nil {
		return nil, nil, err
	}
	encPemBlock, err := x509.EncryptPEMBlock(
		rand.Reader,
		"RSA PRIVATE KEY",
		x509.MarshalPKCS1PrivateKey(privKey),
		pass,
		x509.PEMCipher3DES,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = d.database.ExecContext(
		d.context,
		`
INSERT INTO ca_keys
    (certificate_id, key_pem)
VALUES
    ($1, $2);`,
		1,
		pem.EncodeToMemory(encPemBlock),
	)
	if err != nil {
		return nil, nil, err
	}
	d.certificate = certificate
	d.rsaKey = privKey
	return d.certificate, d.rsaKey, nil
}

func (d *SQLDepot) CreateOrLoadCA(pass []byte, years int, cn, org, country string) (*x509.Certificate, *rsa.PrivateKey, error) {
	var err error
	d.certificate, d.rsaKey, err = d.loadCA(pass)
	if err != nil {
		return nil, nil, err
	}
	if d.certificate != nil && d.rsaKey != nil {
		return d.certificate, d.rsaKey, nil
	}
	return d.createCA(pass, years, cn, org, country)
}

func (d *SQLDepot) CA(pass []byte) ([]*x509.Certificate, *rsa.PrivateKey, error) {
	if d.certificate == nil || d.rsaKey == nil {
		return nil, nil, errors.New("CA certificate or rsaKey is empty")
	}
	return []*x509.Certificate{d.certificate}, d.rsaKey, nil
}

func (d *SQLDepot) Put(name string, certificate *x509.Certificate) error {
	if certificate.Subject.CommonName == "" {
		// this means our cn was replaced by the certificate Signature
		// which is inappropriate for a filename
		name = fmt.Sprintf("%x", sha256.Sum256(certificate.Raw))
	}
	if !certificate.SerialNumber.IsInt64() {
		return errors.New("cannot represent serial number as int64")
	}
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	}
	_, err := d.database.ExecContext(
		d.context, `
INSERT INTO certificates
    (name, not_valid_before, not_valid_after, certificate_pem)
VALUES
    ($1, $2, $3, $4);`,
		name,
		certificate.NotBefore,
		certificate.NotAfter,
		pem.EncodeToMemory(block),
	)

	return err
}

func (d *SQLDepot) Serial() (*big.Int, error) {
	var id int64
	err := d.database.QueryRow(`SELECT max(id) from certificates`).Scan(&id)
	if err != nil {
		panic(err)
	}
	return big.NewInt(id), err
}

func (d *SQLDepot) HasCN(cn string, allowTime int, cert *x509.Certificate, revokeOldCertificate bool) (bool, error) {
	var ct int
	row := d.database.QueryRowContext(d.context, `SELECT COUNT(*) FROM certificates WHERE name = $1 ;`, "ca")
	if err := row.Scan(&ct); err != nil {
		return false, err
	}
	return ct >= 1, nil
}

func (d *SQLDepot) SCEPChallenge() (string, error) {
	rsaKey := make([]byte, 24)
	_, err := rand.Read(rsaKey)
	if err != nil {
		return "", err
	}
	challenge := base64.StdEncoding.EncodeToString(rsaKey)
	_, err = d.database.ExecContext(d.context, `INSERT INTO challenges (challenge) VALUES ($1);`, challenge)
	if err != nil {
		return "", err
	}
	return challenge, nil
}

func (d *SQLDepot) HasChallenge(pw string) (bool, error) {
	result, err := d.database.ExecContext(d.context, `DELETE FROM challenges WHERE challenge = $1;`, pw)
	if err != nil {
		return false, err
	}
	rowCt, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowCt < 1 {
		return false, errors.New("challenge not found")
	}
	return true, nil
}
