package doc

import (
	"embed"
	"io/fs"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed *
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}

// GetDocsFS returns the embedded filesystem containing all documentation
func GetDocsFS() embed.FS {
	return docFS
}

// GetJSWebServerDocsFS returns a sub-filesystem containing just the js-web-server docs
func GetJSWebServerDocsFS() (fs.FS, error) {
	return fs.Sub(docFS, "docs")
}
