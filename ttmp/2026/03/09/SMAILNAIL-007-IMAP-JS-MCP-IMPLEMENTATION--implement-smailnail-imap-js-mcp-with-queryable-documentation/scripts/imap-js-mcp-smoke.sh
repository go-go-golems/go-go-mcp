#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail"

cd "$REPO_ROOT"
./scripts/imap-js-mcp-smoke.sh
