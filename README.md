## How to setup

1. `make .env`.
2. `docker compose up postgres redis`.
3. Wait until the services are up and running.
4. `make setup-keygen`.
5. `make keygen-create-admin-token`. Update AUTHGEAR_ONCE_KEYGEN_ADMIN_TOKEN in .env with the output.
6. `make keygen-create-product`. Update AUTHGEAR_ONCE_KEYGEN_PRODUCT_ID in .env with the output.
7. `make keygen-create-policy`. Update AUTHGEAR_ONCE_KEYGEN_POLICY_ID in .env with the output.
