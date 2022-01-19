DROP TABLE IF EXISTS certificates CASCADE;
CREATE TABLE certificates (
                              id serial NOT NULL PRIMARY KEY,
                              name             VARCHAR(1024) NULL,
                              not_valid_before timestamptz NOT NULL,
                              not_valid_after  timestamptz NOT NULL,
                              certificate_pem  TEXT NOT NULL,
                              revoked          BOOLEAN NOT NULL DEFAULT FALSE,
                              created_at timestamptz NOT NULL DEFAULT current_timestamp,
                              updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

DROP TABLE IF EXISTS ca_keys CASCADE;
CREATE TABLE ca_keys (
                         id serial NOT NULL PRIMARY KEY,
                         certificate_id BIGINT NOT NULL,
                         key_pem  TEXT NOT NULL,
                         created_at timestamptz NOT NULL DEFAULT current_timestamp,
                         updated_at timestamptz NOT NULL DEFAULT current_timestamp,
                         CONSTRAINT fk_certificates FOREIGN KEY (certificate_id) REFERENCES certificates (id)
);

DROP TABLE IF EXISTS challenges CASCADE;
CREATE TABLE challenges (
                            id serial NOT NULL PRIMARY KEY,
                            challenge TEXT,
                            created_at timestamptz NOT NULL DEFAULT current_timestamp,
                            updated_at timestamptz NOT NULL DEFAULT current_timestamp
);
