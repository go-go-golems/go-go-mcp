#!/usr/bin/env bash
set -euo pipefail

SMAILNAIL_ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail"
HOST="${HOST:-89.167.52.236}"
PORT="${PORT:-993}"
USER_NAME="${USER_NAME:-a}"
PASSWORD="${PASSWORD:-pass}"
SUBJECT="${SUBJECT:-Hosted Coolify Dovecot Test}"

cd "$SMAILNAIL_ROOT"

go run ./cmd/imap-tests create-mailbox \
  --server "$HOST" \
  --port "$PORT" \
  --username "$USER_NAME" \
  --password "$PASSWORD" \
  --mailbox INBOX \
  --new-mailbox Archive \
  --insecure \
  --output json

go run ./cmd/imap-tests store-text-message \
  --server "$HOST" \
  --port "$PORT" \
  --username "$USER_NAME" \
  --password "$PASSWORD" \
  --mailbox INBOX \
  --from 'Remote Seeder <seed@example.com>' \
  --to 'User A <a@testcot>' \
  --subject "$SUBJECT" \
  --body 'Remote hosted IMAP fixture validation.' \
  --insecure \
  --output json

go run ./cmd/smailnail fetch-mail \
  --server "$HOST" \
  --port "$PORT" \
  --username "$USER_NAME" \
  --password "$PASSWORD" \
  --mailbox INBOX \
  --subject-contains "$SUBJECT" \
  --insecure \
  --output json
