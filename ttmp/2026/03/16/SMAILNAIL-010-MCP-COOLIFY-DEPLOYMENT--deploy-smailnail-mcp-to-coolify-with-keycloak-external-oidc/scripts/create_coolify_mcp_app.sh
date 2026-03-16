#!/usr/bin/env bash
set -euo pipefail

ssh root@89.167.52.236 '
  export PATH=$PATH:/usr/local/bin:$HOME/go/bin
  coolify app create public \
    --context scapegoat \
    --server-uuid cgl105090ljoxitdf7gmvbrm \
    --project-uuid n8xkgqpbjj04m4pishy3su5e \
    --environment-name production \
    --name smailnail-imap-mcp \
    --git-repository https://github.com/wesen/smailnail \
    --git-branch task/update-imap-mcp \
    --build-pack dockerfile \
    --ports-exposes 3201 \
    --domains https://smailnail.mcp.scapegoat.dev \
    --health-check-enabled \
    --health-check-path /.well-known/oauth-protected-resource \
    --format json
'
