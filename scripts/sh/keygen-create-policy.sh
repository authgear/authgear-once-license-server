#!/bin/sh

set -eux

data=$(jq \
    --null-input \
    --arg POLICY_NAME "$POLICY_NAME" \
    --arg POLICY_DURATION_SECONDS "$POLICY_DURATION_SECONDS" \
    --arg KEYGEN_PRODUCT_ID "$KEYGEN_PRODUCT_ID" \
'
{
    "data": {
        "type": "policy",
        "attributes": {
            "name": $POLICY_NAME,
            "duration": ($POLICY_DURATION_SECONDS | tonumber),
            "strict": true,
            "requireFingerprintScope": true,
            "expirationBasis": "FROM_FIRST_ACTIVATION",
            "authenticationStrategy": "LICENSE"
        },
        "relationships": {
            "product": {
                "data": {
                    "type": "product",
                    "id": $KEYGEN_PRODUCT_ID
                }
            }
        }
    }
}
')

curl -v -H "X-Forwarded-Proto: https" -H "Content-Type: application/json" -H "Authorization: Bearer $KEYGEN_ADMIN_TOKEN" --location "$ENDPOINT/v1/policies" --data "$data"
