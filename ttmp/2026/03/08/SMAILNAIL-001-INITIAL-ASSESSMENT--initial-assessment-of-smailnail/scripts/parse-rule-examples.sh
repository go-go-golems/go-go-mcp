#!/usr/bin/env bash
set -euo pipefail

WORKSPACE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../../.." && pwd)"
cd "$WORKSPACE_ROOT/smailnail"

tmp_go="$(mktemp /tmp/smailnail-parse-rules-XXXXXX.go)"
trap 'rm -f "$tmp_go"' EXIT

cat >"$tmp_go" <<'EOF'
package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/smailnail/pkg/dsl"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: parse-rules <file> [file...]")
		os.Exit(2)
	}

	failures := 0
	for _, path := range os.Args[1:] {
		rule, err := dsl.ParseRuleFile(path)
		if err != nil {
			fmt.Printf("FAIL %s: %v\n", path, err)
			failures++
			continue
		}
		fmt.Printf("OK   %s -> rule=%q format=%q fields=%d\n", path, rule.Name, rule.Output.Format, len(rule.Output.Fields))
	}

	if failures > 0 {
		os.Exit(1)
	}
}
EOF

files=(examples/smailnail/*.yaml examples/complex-search.yaml)
go run "$tmp_go" "${files[@]}"
