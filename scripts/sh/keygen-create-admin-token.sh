#!/bin/sh

set -eux

data=$(jq \
    --null-input \
    --arg TOKEN_NAME "$TOKEN_NAME" \
'
{
    "data": {
        "type": "tokens",
        "attributes": {
            "name": $TOKEN_NAME
        }
    }
}
')

curl -v -H "X-Forwarded-Proto: https" -H "Content-Type: application/json" -u "$KEYGEN_ADMIN_EMAIL:$KEYGEN_ADMIN_PASSWORD" --location "$ENDPOINT/v1/tokens" --data "$data"
