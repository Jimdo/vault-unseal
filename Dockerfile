FROM golang:alpine

ADD . /go/src/github.com/Jimdo/vault-unseal
WORKDIR /go/src/github.com/Jimdo/vault-unseal

RUN set -ex \
 && apk add --update git ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && apk del --purge git

ENTRYPOINT ["/go/bin/vault-unseal"]
