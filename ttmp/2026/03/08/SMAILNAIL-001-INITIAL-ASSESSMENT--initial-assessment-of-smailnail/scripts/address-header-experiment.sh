#!/usr/bin/env bash
set -euo pipefail

WORKSPACE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../../.." && pwd)"
cd "$WORKSPACE_ROOT/smailnail"

tmp_go="$(mktemp /tmp/smailnail-address-header-XXXXXX.go)"
trap 'rm -f "$tmp_go"' EXIT

cat >"$tmp_go" <<'EOF'
package main

import (
	"fmt"

	"github.com/emersion/go-message/mail"
)

func main() {
	h := mail.Header{}
	h.SetAddressList("From", []*mail.Address{{Address: "John Doe <john@example.com>"}})
	h.SetAddressList("To", []*mail.Address{{Address: "user@example.com"}})

	fmt.Println("from:", h.Get("From"))
	fmt.Println("to:", h.Get("To"))
}
EOF

go run "$tmp_go"
