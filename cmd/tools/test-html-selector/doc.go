package main

import (
	"embed"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed doc/general-topics/*.md
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
