#!/usr/bin/env bash
set -euo pipefail

root="/home/manuel/workspaces/2026-03-08/update-imap-mcp"

show() {
  local title="$1"
  local file="$2"
  local range="$3"
  echo
  echo "== ${title} =="
  nl -ba "${file}" | sed -n "${range}"
}

show "smailnail JS module" \
  "${root}/smailnail/pkg/js/modules/smailnail/module.go" \
  '1,120p'

show "smailnail JS service" \
  "${root}/smailnail/pkg/services/smailnailjs/service.go" \
  '1,260p'

show "jesus executeJS MCP pattern" \
  "${root}/jesus/pkg/mcp/server.go" \
  '117,395p'

show "go-go-goja jsdoc CLI entrypoint" \
  "${root}/go-go-goja/cmd/goja-jsdoc/main.go" \
  '1,120p'

show "go-go-goja jsdoc data model" \
  "${root}/go-go-goja/pkg/jsdoc/model/model.go" \
  '1,120p'

show "go-go-goja jsdoc store" \
  "${root}/go-go-goja/pkg/jsdoc/model/store.go" \
  '1,120p'

show "go-go-goja jsdoc extractor" \
  "${root}/go-go-goja/pkg/jsdoc/extract/extract.go" \
  '1,360p'

show "go-go-goja glazehelp query pattern" \
  "${root}/go-go-goja/modules/glazehelp/glazehelp.go" \
  '1,180p'

show "go-go-goja jsdoc SQLite export" \
  "${root}/go-go-goja/pkg/jsdoc/exportsq/exportsq.go" \
  '1,260p'
