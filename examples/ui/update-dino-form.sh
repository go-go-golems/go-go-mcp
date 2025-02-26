#!/bin/bash
set -euo pipefail

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Read the YAML file
YAML_CONTENT=$(cat "$SCRIPT_DIR/example-form.yaml")

# Run the update-ui command with the YAML content
go-go-mcp run-command "$SCRIPT_DIR/update-ui.yaml" \
  --components "$YAML_CONTENT" \
  --verbose 