#!/usr/bin/env bash
set -x

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

APIURL=${APIURL:-https://localhost:8443/api}
USERNAME=${USERNAME:-u$(date +%s)}
EMAIL=${EMAIL:-$USERNAME@example.com}
PASSWORD=${PASSWORD:-pA55w0Rd!}
NAME=${NAME:-$USERNAME}

npx newman run $SCRIPTDIR/Conduit.postman_collection.json \
    --delay-request 500 \
    --insecure \
    --ssl-client-cert ~/certs/localCA.pem \
    --global-var "APIURL=$APIURL" \
    --global-var "USERNAME=$USERNAME" \
    --global-var "EMAIL=$EMAIL" \
    --global-var "PASSWORD=$PASSWORD" \
    --global-var "NAME=$NAME"