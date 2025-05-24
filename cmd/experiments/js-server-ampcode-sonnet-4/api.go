package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/jsserver"
	"github.com/spf13/cobra"
)

func loadJavaScriptFiles(server *jsserver.JSWebServer, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".js") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read %s: %v", path, err)
			return nil // Continue with other files
		}

		// Extract filename without extension for name
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		// Execute the JavaScript file via HTTP API
		req := jsserver.ExecuteRequest{
			Code:    string(content),
			Persist: true,
			Name:    fmt.Sprintf("startup-%s", name),
		}

		// Make HTTP request to execute the code
		jsonBody, _ := json.Marshal(req)
		resp, err := http.Post("http://localhost:"+port+"/api/execute", "application/json", bytes.NewReader(jsonBody))
		if err != nil {
			log.Printf("Failed to execute %s: %v", path, err)
			return nil
		}
		defer resp.Body.Close()

		var response jsserver.ExecuteResponse
		json.NewDecoder(resp.Body).Decode(&response)

		if response.Success {
			log.Printf("Loaded JavaScript file: %s", path)
		} else {
			log.Printf("Failed to load %s: %s", path, response.Error)
		}

		return nil
	})
}

func executeCode(cmd *cobra.Command, args []string) error {
	var code string

	if len(args) == 1 {
		// Read from file
		content, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		code = string(content)
	} else {
		// Read from stdin
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		code = string(content)
	}

	persist, _ := cmd.Flags().GetBool("persist")
	name, _ := cmd.Flags().GetString("name")

	req := jsserver.ExecuteRequest{
		Code:    code,
		Persist: persist,
		Name:    name,
	}

	return makeAPIRequest("POST", "/api/execute", req, true)
}

func listRoutes(cmd *cobra.Command, args []string) error {
	return makeAPIRequest("GET", "/api/routes", nil, true)
}

func deleteRoute(cmd *cobra.Command, args []string) error {
	return makeAPIRequest("DELETE", "/api/routes/"+strings.TrimPrefix(args[0], "/"), nil, false)
}

func listFiles(cmd *cobra.Command, args []string) error {
	return makeAPIRequest("GET", "/api/files", nil, true)
}

func deleteFile(cmd *cobra.Command, args []string) error {
	return makeAPIRequest("DELETE", "/api/files/"+strings.TrimPrefix(args[0], "/"), nil, false)
}

func getState(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return makeAPIRequest("GET", "/api/state", nil, true)
	}
	return makeAPIRequest("GET", "/api/state/"+args[0], nil, true)
}

func setState(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	// Try to parse value as JSON, otherwise treat as string
	var jsonValue interface{}
	if err := json.Unmarshal([]byte(value), &jsonValue); err != nil {
		jsonValue = value
	}

	req := map[string]interface{}{"value": jsonValue}
	return makeAPIRequest("PUT", "/api/state/"+key, req, false)
}

func deleteState(cmd *cobra.Command, args []string) error {
	return makeAPIRequest("DELETE", "/api/state/"+args[0], nil, false)
}

func listArchive(cmd *cobra.Command, args []string) error {
	return makeAPIRequest("GET", "/api/archive", nil, true)
}

func getArchive(cmd *cobra.Command, args []string) error {
	return makeAPIRequest("GET", "/api/archive/"+args[0], nil, true)
}

func getStatus(cmd *cobra.Command, args []string) error {
	return makeAPIRequest("GET", "/status", nil, true)
}

func loadDirectory(cmd *cobra.Command, args []string) error {
	dir := args[0]

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".js") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read %s: %v", path, err)
			return nil // Continue with other files
		}

		// Extract filename without extension for name
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		req := jsserver.ExecuteRequest{
			Code:    string(content),
			Persist: true,
			Name:    name,
		}

		fmt.Printf("Loading %s...\n", path)
		if err := makeAPIRequest("POST", "/api/execute", req, false); err != nil {
			log.Printf("Failed to load %s: %v", path, err)
		} else {
			fmt.Printf("✓ Loaded %s\n", path)
		}

		return nil
	})
}

func makeAPIRequest(method, endpoint string, body interface{}, printResponse bool) error {
	url := serverURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(responseBody))
	}

	if printResponse && len(responseBody) > 0 {
		// Pretty print JSON if possible
		var jsonData interface{}
		if err := json.Unmarshal(responseBody, &jsonData); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
			fmt.Println(string(prettyJSON))
		} else {
			fmt.Println(string(responseBody))
		}
	} else if !printResponse {
		fmt.Printf("✓ %s %s\n", method, endpoint)
	}

	return nil
}
