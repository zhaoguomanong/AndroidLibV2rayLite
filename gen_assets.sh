#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

# Set magic variables for current file & dir
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__file="${__dir}/$(basename "${BASH_SOURCE[0]}")"
__base="$(basename ${__file} .sh)"


ASSETS=${__dir}/assets
mkdir -p ${ASSETS}

gen_assets() {
    local TMPDIR=$(mktemp -d)

    trap 'echo -e "Aborted, error $? in command: $BASH_COMMAND"; rm -rf $TMPDIR; trap ERR; exit 1' ERR

    local GEOSITE=${GOPATH}/src/github.com/v2ray/domain-list-community
    if [[ -d ${GEOSITE} ]]; then
        cd ${GEOSITE} && git pull
    else
        mkdir -p ${GEOSITE}
        cd ${GEOSITE} && git clone https://github.com/v2ray/domain-list-community.git .
    fi
    go run main.go
    mv dlc.dat $ASSETS/geosite.dat

    if [[ ! -x $GOPATH/bin/geoip ]]; then
        go get -v -u github.com/v2ray/geoip
    fi

    cd $TMPDIR
    curl -L -O http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country-CSV.zip
    unzip GeoLite2-Country-CSV.zip
    mkdir geoip && find . -name '*.csv' -exec mv {} ./geoip \;
    $GOPATH/bin/geoip \
        --country=./geoip/GeoLite2-Country-Locations-en.csv \
        --ipv4=./geoip/GeoLite2-Country-Blocks-IPv4.csv \
        --ipv6=./geoip/GeoLite2-Country-Blocks-IPv6.csv

    mv ./geoip.dat $ASSETS/geoip.dat
    trap ERR
    return 0
}

gen_assets
