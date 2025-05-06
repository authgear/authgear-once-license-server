#!/bin/sh

# duration
# https://keygen.sh/docs/api/policies/#policies-object-attrs-duration
# The default is null, which means never expire.
# The minimum value is 86400, which is 1-day.
#
# strict=true
# https://keygen.sh/docs/api/policies/#policies-object-attrs-strict
# The default is false.
# If strict=true, then a license is valid if the license is associated
# with a machine according to the machine limit.
#
# requireFingerprintScope=true
# https://keygen.sh/docs/api/policies/#policies-object-attrs-requireFingerprintScope
# The default is false.
# If requireFingerprintScope=true, then scope.fingerprint is required in /v1/licenses/actions/validate-key
# This means we need to persist the fingerprint
#
# expirationStrategy=MAINTAIN_ACCESS
# https://keygen.sh/docs/api/policies/#policies-object-attrs-expirationStrategy
# The default is RESTRICT_ACCESS.
# If expirationStrategy=MAINTAIN_ACCESS, then valid=true even if code=EXPIRED.
# This means our code will first look at valid, if it is true, then even CODE!=VALID is still considered as valid.
#
# expirationBasis=FROM_FIRST_ACTIVATION
# https://keygen.sh/docs/api/policies/#policies-object-attrs-expirationBasis
# The default is FROM_CREATION.
# If expirationBasis=FROM_FIRST_ACTIVATION, then the license is set after the first machine activation event.
# This matches our intended use case.
#
# authenticationStrategy=LICENSE
# https://keygen.sh/docs/api/policies/#policies-object-attrs-authenticationStrategy
# The default is TOKEN.
# If authenticationStrategy=LICENSE, then the license key itself can be used to authenticate.
# This matches our intended use case, as the bearer of the license key is granted access.

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
            "expirationStrategy": "MAINTAIN_ACCESS",
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
