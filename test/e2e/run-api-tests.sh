#!/usr/bin/env bash
set -x

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

APIURL=${APIURL:-https://localhost:8000}
USERNAME=${USERNAME:-u$(date +%s)}
EMAIL=${EMAIL:-$USERNAME@mail.com}
PASSWORD=${PASSWORD:-pA55w0Rd!}

npx newman run $SCRIPTDIR/Conduit.postman_collection.json \
    --delay-request 500 \
    --ssl-client-cert ~/certs/localCA.pem \
    --global-var "APIURL=$APIURL" \
    --global-var "USERNAME=$USERNAME" \
    --global-var "EMAIL=$EMAIL" \
    --global-var "PASSWORD=$PASSWORD"