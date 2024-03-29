#!/bin/sh 

VERSION=$(git describe --tags --exact-match)
REPO=$(basename $(pwd))
ARCHS="linux/386 linux/amd64 linux/arm darwin/amd64"

set -e

if [ -z "${VERSION}" ]; then
	echo "No tag present, stopping build now."
	exit 0
fi

if [ -z "${GITHUB_TOKEN}" ]; then
	echo "Please set \$GITHUB_TOKEN environment variable"
	exit 1
fi

set -x

go install github.com/aktau/github-release@latest
go install github.com/mitchellh/gox@latest

github-release release --user Jimdo --repo ${REPO} --tag ${VERSION} --name ${VERSION} || true

gox -ldflags="-X main.version=${VERSION}" -osarch="${ARCHS}"
for file in ${REPO}_*; do
  github-release upload --user Jimdo --repo ${REPO} --tag ${VERSION} --name ${file} --file ${file}
done
