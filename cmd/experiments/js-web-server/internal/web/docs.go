package web

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/web/templates"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/pkg/doc"
)

var docsFS fs.FS

func init() {
	var err error
	docsFS, err = doc.GetJSWebServerDocsFS()
	if err != nil {
		panic("Failed to initialize docs filesystem: " + err.Error())
	}
}

// Markdown renderer with extensions
var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Typographer,
		extension.DefinitionList,
		extension.Footnote,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

type DocInfo struct {
	Filename string
	Title    string
	Content  string
}

// getDocumentList returns a map of filename -> title for all documentation files
func getDocumentList() (map[string]string, error) {
	entries, err := fs.ReadDir(docsFS, ".")
	if err != nil {
		return nil, err
	}

	docs := make(map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			// Read the file to extract title from first line
			content, err := fs.ReadFile(docsFS, entry.Name())
			if err != nil {
				continue
			}

			title := extractTitle(string(content))
			if title == "" {
				title = strings.TrimSuffix(entry.Name(), ".md")
			}
			docs[entry.Name()] = title
		}
	}

	return docs, nil
}

// extractTitle extracts the title from the first line of markdown
func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "#"))
		}
	}
	return ""
}

// getPresetExamples returns a list of preset code examples
func getPresetExamples() []templates.PresetExample {
	return []templates.PresetExample{
		{
			ID:          "hello-world",
			Name:        "Hello World",
			Description: "Basic console output example",
			Code: `// Hello World example
console.log("Hello, JavaScript Playground!");

// Return a simple object
return {
    message: "Welcome to the JS Playground!",
    timestamp: new Date().toISOString(),
    status: "success"
};`,
		},
		{
			ID:          "express-basic",
			Name:        "Express Route",
			Description: "Create a simple Express.js route",
			Code: `// Basic Express.js route example
app.get("/hello", (req, res) => {
    res.json({
        message: "Hello from Express!",
        timestamp: new Date().toISOString(),
        path: req.path
    });
});

console.log("Route registered: GET /hello");
return "Express route created successfully!";`,
		},
		{
			ID:          "database-query",
			Name:        "Database Query",
			Description: "Example of database interaction",
			Code: `// Database query example
const users = db.query("SELECT * FROM users LIMIT 5");
console.log("Found users:", users.length);

// Create a new user
const newUser = {
    name: "Test User",
    email: "test@example.com",
    created_at: new Date().toISOString()
};

db.exec("INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)", 
        [newUser.name, newUser.email, newUser.created_at]);

console.log("User created successfully");
return { users, newUser };`,
		},
		{
			ID:          "api-endpoints",
			Name:        "RESTful API",
			Description: "Complete CRUD API example",
			Code: `// RESTful API example
const items = [];
let nextId = 1;

// GET all items
app.get("/api/items", (req, res) => {
    res.json(items);
});

// POST new item
app.post("/api/items", (req, res) => {
    const item = {
        id: nextId++,
        name: req.body.name || "Unnamed Item",
        created: new Date().toISOString()
    };
    items.push(item);
    res.status(201).json(item);
});

// GET single item
app.get("/api/items/:id", (req, res) => {
    const item = items.find(i => i.id === parseInt(req.params.id));
    if (!item) return res.status(404).json({error: "Item not found"});
    res.json(item);
});

console.log("RESTful API endpoints registered");
return "API endpoints created: GET/POST /api/items, GET /api/items/:id";`,
		},
		{
			ID:          "middleware-example",
			Name:        "Middleware",
			Description: "Express middleware example with logging",
			Code: `// Middleware example
app.use("/api/*", (req, res, next) => {
    console.log('[' + new Date().toISOString() + '] ' + req.method + ' ' + req.path);
    
    // Add CORS headers
    res.header("Access-Control-Allow-Origin", "*");
    res.header("Access-Control-Allow-Headers", "Content-Type");
    
    next();
});

// Protected route
app.get("/api/protected", (req, res) => {
    const token = req.headers.authorization;
    if (!token) {
        return res.status(401).json({error: "No authorization token"});
    }
    
    res.json({
        message: "Access granted!",
        token: token,
        timestamp: new Date().toISOString()
    });
});

console.log("Middleware and protected route registered");
return "Middleware applied to /api/* routes";`,
		},
		{
			ID:          "data-processing",
			Name:        "Data Processing",
			Description: "Advanced data manipulation example",
			Code: `// Data processing example
const sampleData = [
    {id: 1, name: "Alice", age: 30, department: "Engineering"},
    {id: 2, name: "Bob", age: 25, department: "Design"},
    {id: 3, name: "Charlie", age: 35, department: "Engineering"},
    {id: 4, name: "Diana", age: 28, department: "Marketing"}
];

// Filter and transform data
const engineers = sampleData
    .filter(person => person.department === "Engineering")
    .map(person => ({
        ...person,
        seniorityLevel: person.age > 30 ? "Senior" : "Junior"
    }));

// Group by department
const byDepartment = sampleData.reduce((acc, person) => {
    acc[person.department] = acc[person.department] || [];
    acc[person.department].push(person);
    return acc;
}, {});

// Calculate statistics
const avgAge = sampleData.reduce((sum, p) => sum + p.age, 0) / sampleData.length;

console.log("Engineers:", engineers);
console.log("By Department:", byDepartment);
console.log("Average Age:", avgAge);

return {engineers, byDepartment, avgAge};`,
		},
	}
}

// DocsHandler handles documentation page requests
func DocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docs, err := getDocumentList()
		if err != nil {
			http.Error(w, "Failed to load documentation", http.StatusInternalServerError)
			return
		}

		selectedDoc := r.URL.Query().Get("doc")
		var content string

		if selectedDoc != "" {
			// Read and render the selected document
			docContent, err := fs.ReadFile(docsFS, selectedDoc)
			if err != nil {
				http.Error(w, "Document not found", http.StatusNotFound)
				return
			}

			// Convert markdown to HTML
			var buf strings.Builder
			if err := md.Convert(docContent, &buf); err != nil {
				http.Error(w, "Failed to render markdown", http.StatusInternalServerError)
				return
			}
			content = buf.String()
		}

		presets := getPresetExamples()
		component := templates.DocsPageWithPresets(docs, selectedDoc, content, presets)

		w.Header().Set("Content-Type", "text/html")
		err = component.Render(context.Background(), w)
		if err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}
	}
}

// PresetHandler returns preset examples as JSON
func PresetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		presetID := r.URL.Query().Get("id")
		presets := getPresetExamples()

		for _, preset := range presets {
			if preset.ID == presetID {
				w.Header().Set("Content-Type", "application/json")

				// Use proper JSON marshaling
				data, err := json.Marshal(preset)
				if err != nil {
					http.Error(w, "Failed to encode preset", http.StatusInternalServerError)
					return
				}

				_, _ = w.Write(data)
				return
			}
		}

		http.Error(w, "Preset not found", http.StatusNotFound)
	}
}
