#!/usr/bin/env bash
set -euo pipefail

WORKSPACE_ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp"
SMAILNAIL_ROOT="$WORKSPACE_ROOT/smailnail"
DOCKER_ROOT="/home/manuel/code/others/docker-test-dovecot"

STAMP="$(date +%s)"
ARCHIVE_MAILBOX="FacadeArchive${STAMP}"
ACTION_SUBJECT="Facade Action Validation ${STAMP}"
MAILGEN_SUBJECT="Facade Mailgen Validation ${STAMP}"

tmpdir="$(mktemp -d /tmp/smailnail-docker-validation-XXXXXX)"
trap 'rm -rf "$tmpdir"' EXIT

action_rule="$tmpdir/action-rule.yaml"
mailgen_config="$tmpdir/mailgen-display-name.yaml"

cat >"$action_rule" <<EOF
name: "Facade action validation"
description: "Validate UID-based actions against the Docker IMAP fixture"
search:
  subject_contains: "${ACTION_SUBJECT}"
output:
  format: json
  fields:
    - uid
    - subject
    - flags
actions:
  flags:
    add: ["seen", "flagged"]
  copy_to: "${ARCHIVE_MAILBOX}"
EOF

cat >"$mailgen_config" <<EOF
variables:
  noop: "noop"
templates:
  basic:
    subject: "${MAILGEN_SUBJECT}"
    from: "Facade Sender <sender@example.com>"
    to: "User A <a@testcot>"
    body: |
      This message validates display-name address serialization.
rules:
  validate:
    template: basic
    variations:
      - marker: "{{ .variables.noop }}"
generate:
  - rule: validate
    count: 1
EOF

echo "==> Starting Docker IMAP fixture"
cd "$DOCKER_ROOT"
docker compose up -d --build

echo "==> Waiting for IMAP port"
for _ in $(seq 1 30); do
  if nc -z 127.0.0.1 993 >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "==> CLI help smoke"
cd "$SMAILNAIL_ROOT"
go run ./cmd/smailnail --help >/dev/null
go run ./cmd/mailgen --help >/dev/null
go run ./cmd/imap-tests --help >/dev/null

echo "==> Creating target mailbox"
go run ./cmd/imap-tests create-mailbox \
  --server localhost \
  --username a \
  --password pass \
  --mailbox INBOX \
  --new-mailbox "$ARCHIVE_MAILBOX" \
  --insecure \
  --output json >/dev/null

echo "==> Storing unique source message for action validation"
go run ./cmd/imap-tests store-text-message \
  --server localhost \
  --username a \
  --password pass \
  --mailbox INBOX \
  --from "Action Seeder <seed@example.com>" \
  --to "User A <a@testcot>" \
  --subject "$ACTION_SUBJECT" \
  --body "This message should be flagged and copied by smailnail actions." \
  --insecure \
  --output json >/dev/null

echo "==> Running mail-rules action validation"
go run ./cmd/smailnail mail-rules \
  --rule "$action_rule" \
  --server localhost \
  --username a \
  --password pass \
  --mailbox INBOX \
  --insecure \
  --output json >/dev/null

echo "==> Verifying flags were applied to the intended message"
inbox_fetch="$(go run ./cmd/smailnail fetch-mail \
  --server localhost \
  --username a \
  --password pass \
  --mailbox INBOX \
  --subject-contains "$ACTION_SUBJECT" \
  --insecure \
  --output json)"
printf '%s\n' "$inbox_fetch"
printf '%s\n' "$inbox_fetch" | grep -F "$ACTION_SUBJECT" >/dev/null
printf '%s\n' "$inbox_fetch" | grep -F 'Flagged' >/dev/null
printf '%s\n' "$inbox_fetch" | grep -F 'Seen' >/dev/null

echo "==> Verifying copied message exists in archive mailbox"
archive_fetch="$(go run ./cmd/smailnail fetch-mail \
  --server localhost \
  --username a \
  --password pass \
  --mailbox "$ARCHIVE_MAILBOX" \
  --subject-contains "$ACTION_SUBJECT" \
  --insecure \
  --output json)"
printf '%s\n' "$archive_fetch"
printf '%s\n' "$archive_fetch" | grep -F "$ACTION_SUBJECT" >/dev/null

echo "==> Storing generated mail with display-name addresses"
go run ./cmd/mailgen generate \
  --configs "$mailgen_config" \
  --store-imap \
  --server localhost \
  --username a \
  --password pass \
  --mailbox INBOX \
  --insecure \
  --output json >/dev/null

echo "==> Verifying generated mail round-trips display name"
mailgen_fetch="$(go run ./cmd/smailnail fetch-mail \
  --server localhost \
  --username a \
  --password pass \
  --mailbox INBOX \
  --subject-contains "$MAILGEN_SUBJECT" \
  --insecure \
  --output json)"
printf '%s\n' "$mailgen_fetch"
printf '%s\n' "$mailgen_fetch" | grep -F "$MAILGEN_SUBJECT" >/dev/null
printf '%s\n' "$mailgen_fetch" | grep -F 'Facade Sender' >/dev/null
printf '%s\n' "$mailgen_fetch" | grep -F 'sender@example.com' >/dev/null

echo "Validation completed successfully."
