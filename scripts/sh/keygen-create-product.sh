#!/bin/sh

# distributionStrategy=LICENSED
# https://keygen.sh/docs/api/products/#products-object-attrs-distributionStrategy
# LICENSED is the default value.
# It means the product required license to access.

set -eux

data=$(jq \
    --null-input \
    --arg PRODUCT_NAME "$PRODUCT_NAME" \
'
{
    "data": {
        "type": "product",
        "attributes": {
            "name": $PRODUCT_NAME,
            "distributionStrategy": "LICENSED"
        }
    }
}
')

curl -v -H "X-Forwarded-Proto: https" -H "Content-Type: application/json" -H "Authorization: Bearer $KEYGEN_ADMIN_TOKEN" --location "$ENDPOINT/v1/products" --data "$data"
