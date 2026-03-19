# Tasks

## TODO

- [x] Capture the current smailnail production logs for Claude MCP login attempts
- [x] Patch the MCP bearer challenge to advertise only `resource_metadata`
- [x] Suppress `/readyz` noise from hosted request debug logs
- [x] Redeploy the hosted MCP app on Coolify and verify the new challenge is live
- [x] Inspect Keycloak logs and confirm whether Claude reaches dynamic client registration
- [x] Inspect Keycloak realm configuration and client registration policies with `kcadm.sh`
- [x] Write an intern-facing design and implementation guide explaining the full auth stack
- [x] Record the debugging chronology in the ticket diary
- [x] Decide whether to fix Claude login via anonymous DCR policy changes or a pre-provisioned client strategy
- [x] Apply the chosen Keycloak-side remediation and validate Claude end-to-end
