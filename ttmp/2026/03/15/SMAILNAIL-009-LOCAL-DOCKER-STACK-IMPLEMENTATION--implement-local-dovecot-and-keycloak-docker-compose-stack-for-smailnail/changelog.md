# Changelog

## 2026-03-15

- Initial workspace created
- Added `smailnail/docker-compose.local.yml` with Dovecot, Keycloak, and PostgreSQL services
- Added a pre-imported `smailnail-dev` Keycloak realm with starter `smailnail-web` and `smailnail-mcp` clients
- Documented local stack startup, ports, credentials, issuer URL, and shutdown in `smailnail/README.md`
- Verified the stack locally with `docker compose ps`, Keycloak discovery, and Dovecot IMAPS reachability
