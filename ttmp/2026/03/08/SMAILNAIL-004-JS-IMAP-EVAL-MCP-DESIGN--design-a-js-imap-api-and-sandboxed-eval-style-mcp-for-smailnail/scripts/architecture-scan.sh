#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp"

echo "== smailnail core =="
nl -ba "$ROOT/smailnail/pkg/imap/layer.go" | sed -n '1,120p'
nl -ba "$ROOT/smailnail/pkg/dsl/types.go" | sed -n '1,260p'
nl -ba "$ROOT/smailnail/pkg/dsl/processor.go" | sed -n '1,260p'
nl -ba "$ROOT/smailnail/cmd/smailnail/commands/fetch_mail.go" | sed -n '1,260p'
nl -ba "$ROOT/smailnail/cmd/smailnail/commands/mail_rules.go" | sed -n '1,260p'
nl -ba "$ROOT/smailnail/pkg/mailgen/mailgen.go" | sed -n '1,220p'

echo "== go-go-goja native module/runtime =="
nl -ba "$ROOT/go-go-goja/modules/common.go" | sed -n '1,220p'
nl -ba "$ROOT/go-go-goja/engine/factory.go" | sed -n '1,220p'
sed -n '1,220p' "$ROOT/go-go-goja/pkg/doc/02-creating-modules.md"

echo "== jesus sandbox/eval reference =="
nl -ba "$ROOT/jesus/pkg/engine/engine.go" | sed -n '1,220p'
nl -ba "$ROOT/jesus/pkg/mcp/server.go" | sed -n '1,260p'
nl -ba "$ROOT/jesus/pkg/repl/model.go" | sed -n '1,220p'
