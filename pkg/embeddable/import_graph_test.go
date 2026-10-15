package embeddable

import (
	"os/exec"
	"strings"
	"testing"
)

// TestImportGraphExcludesEmbeddedAuthorizationServer keeps the embedded
// development OIDC server, and the fosite/SQLite/gRPC dependencies it brings,
// out of every program that imports this package. Programs that want
// embedded_dev link it explicitly through package embeddeddev.
func TestImportGraphExcludesEmbeddedAuthorizationServer(t *testing.T) {
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go tool not available")
	}
	out, err := exec.Command(goTool, "list", "-deps", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps: %v\n%s", err, out)
	}
	for _, forbidden := range []string{
		"github.com/go-go-golems/go-go-mcp/pkg/auth/oidc",
		"github.com/ory/fosite",
		"github.com/mattn/go-sqlite3",
		"google.golang.org/grpc",
	} {
		for _, dep := range strings.Fields(string(out)) {
			if dep == forbidden || strings.HasPrefix(dep, forbidden+"/") {
				t.Errorf("pkg/embeddable depends on %s", dep)
				break
			}
		}
	}
}
