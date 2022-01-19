module com/ninjaone/ninjascepserver

go 1.17

require (
	github.com/go-kit/kit v0.4.0
	github.com/lib/pq v1.10.4
	github.com/micromdm/scep/v2 v2.1.0
)

require (
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/groob/finalizer v0.0.0-20170707115354-4c2ed49aabda // indirect
	github.com/pkg/errors v0.8.0 // indirect
	go.mozilla.org/pkcs7 v0.0.0-20210826202110-33d05740a352 // indirect
	golang.org/x/net v0.0.0-20191009170851-d66e71096ffb // indirect
	golang.org/x/sys v0.0.0-20190412213103-97732733099d // indirect
)

replace go.mozilla.org/pkcs7 v0.0.0-20200128120323-432b2356ecb1 => github.com/omorsi/pkcs7 v0.0.0-20210217142924-a7b80a2a8568
