# Tasks

## Completed

- [x] Define the local stack scope around Dovecot, Keycloak, and Keycloak persistence.
- [x] Add a repo-local `docker-compose.local.yml` for Dovecot, Keycloak, and PostgreSQL.
- [x] Add a Keycloak realm import with initial `smailnail-web` and `smailnail-mcp` clients.
- [x] Document startup, ports, credentials, issuer URL, and shutdown in the smailnail README.
- [x] Bring the stack up locally with `docker compose -f docker-compose.local.yml up -d`.
- [x] Verify the stack with `docker compose ps`, the Keycloak OIDC discovery endpoint, and Dovecot IMAPS reachability.
- [x] Record the implementation details, verification commands, and results in the ticket docs and diary.
