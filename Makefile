VERSION = $(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
OSARCH=$(shell go env GOHOSTOS)-$(shell go env GOHOSTARCH)

NINJASCEPSERVER=\
	ninjascepserver-darwin-amd64 \
	ninjascepserver-darwin-arm64 \
	ninjascepserver-linux-amd64

my: ninjascepserver-$(OSARCH)

docker: ninjascepserver-linux-amd64

$(NINJASCEPSERVER):
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<

%-$(VERSION).zip: %.exe
	rm -f $@
	zip $@ $<

%-$(VERSION).zip: %
	rm -f $@
	zip $@ $<

clean:
	rm -f ninjascepserver-*

release: $(foreach bin,$(NINJASCEPSERVER),$(subst .exe,,$(bin))-$(VERSION).zip)

test:
	go test -v -cover -race ./...

.PHONY: my docker $(NINJASCEPSERVER) clean release test
