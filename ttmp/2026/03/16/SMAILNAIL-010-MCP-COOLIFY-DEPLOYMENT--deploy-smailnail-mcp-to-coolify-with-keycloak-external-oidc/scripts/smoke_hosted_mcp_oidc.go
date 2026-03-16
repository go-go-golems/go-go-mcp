package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

type executeResponse struct {
	Success bool            `json:"success"`
	Value   json.RawMessage `json:"value"`
	Error   any             `json:"error"`
}

func main() {
	var (
		serverURL    = flag.String("server-url", envOrDefault("SMAILNAIL_MCP_SERVER_URL", "https://smailnail.mcp.scapegoat.dev/mcp"), "Hosted MCP endpoint")
		tokenURL     = flag.String("token-url", envOrDefault("KEYCLOAK_TOKEN_URL", "https://auth.scapegoat.dev/realms/smailnail/protocol/openid-connect/token"), "Keycloak token endpoint")
		clientID     = flag.String("client-id", envOrDefault("SMAILNAIL_MCP_SMOKE_CLIENT_ID", "smailnail-mcp-smoke"), "Keycloak confidential client ID")
		clientSecret = flag.String("client-secret", os.Getenv("SMAILNAIL_MCP_SMOKE_CLIENT_SECRET"), "Keycloak confidential client secret")
		imapServer   = flag.String("imap-server", envOrDefault("SMAILNAIL_IMAP_HOST", "89.167.52.236"), "Hosted IMAP server")
		imapPort     = flag.Int("imap-port", envOrDefaultInt("SMAILNAIL_IMAP_PORT", 993), "Hosted IMAP port")
		imapUser     = flag.String("imap-username", envOrDefault("SMAILNAIL_IMAP_USERNAME", "a"), "Hosted IMAP username")
		imapPassword = flag.String("imap-password", envOrDefault("SMAILNAIL_IMAP_PASSWORD", "pass"), "Hosted IMAP password")
		imapMailbox  = flag.String("imap-mailbox", envOrDefault("SMAILNAIL_IMAP_MAILBOX", "INBOX"), "Mailbox to select")
		imapInsecure = flag.Bool("imap-insecure", envOrDefaultBool("SMAILNAIL_IMAP_INSECURE", true), "Skip TLS verification for the hosted IMAP fixture")
		timeout      = flag.Duration("timeout", 30*time.Second, "Overall timeout")
	)
	flag.Parse()

	if *clientSecret == "" {
		fatalf("client secret is required; pass --client-secret or set SMAILNAIL_MCP_SMOKE_CLIENT_SECRET")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	token, err := fetchAccessToken(ctx, *tokenURL, *clientID, *clientSecret)
	if err != nil {
		fatalf("fetch access token: %v", err)
	}

	client, err := mcpclient.NewStreamableHttpClient(
		*serverURL,
		transport.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		}),
		transport.WithHTTPBasicClient(&http.Client{Timeout: *timeout}),
	)
	if err != nil {
		fatalf("create MCP client: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	if err := client.Start(ctx); err != nil {
		fatalf("start MCP client: %v", err)
	}

	if _, err := client.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "smailnail-ticket-smoke",
				Version: "dev",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	}); err != nil {
		fatalf("initialize MCP client: %v", err)
	}

	tools, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		fatalf("list tools: %v", err)
	}
	if !hasTool(tools, "executeIMAPJS") {
		fatalf("executeIMAPJS not advertised by hosted server")
	}

	code := buildSmokeCode(*imapServer, *imapPort, *imapUser, *imapPassword, *imapMailbox, *imapInsecure)
	result, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "executeIMAPJS",
			Arguments: map[string]any{
				"code": code,
			},
		},
	})
	if err != nil {
		fatalf("call executeIMAPJS: %v", err)
	}
	if result.IsError {
		fatalf("tool returned error result: %s", firstText(result.Content))
	}
	if len(result.Content) == 0 {
		fatalf("tool returned no content")
	}

	text, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		fatalf("expected first content item to be text, got %T", result.Content[0])
	}

	var decoded executeResponse
	if err := json.Unmarshal([]byte(text.Text), &decoded); err != nil {
		fatalf("decode tool response: %v", err)
	}
	if !decoded.Success {
		fatalf("tool reported failure: %v", decoded.Error)
	}

	output := map[string]any{
		"tokenEndpoint": *tokenURL,
		"serverURL":     *serverURL,
		"toolCount":     len(tools.Tools),
		"value":         json.RawMessage(decoded.Value),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		fatalf("encode output: %v", err)
	}
}

func fetchAccessToken(ctx context.Context, tokenURL, clientID, clientSecret string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded tokenResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", err
	}
	if decoded.AccessToken == "" {
		return "", fmt.Errorf("token endpoint returned empty access token")
	}
	return decoded.AccessToken, nil
}

func buildSmokeCode(server string, port int, username, password, mailbox string, insecure bool) string {
	return fmt.Sprintf(`
const smailnail = require("smailnail");
const svc = smailnail.newService();
const session = svc.connect({
  server: %q,
  port: %d,
  username: %q,
  password: %q,
  mailbox: %q,
  insecure: %t
});
const result = { mailbox: session.mailbox };
session.close();
result;
`, server, port, username, password, mailbox, insecure)
}

func hasTool(tools *mcp.ListToolsResult, name string) bool {
	if tools == nil {
		return false
	}
	for _, tool := range tools.Tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var parsed int
		if _, err := fmt.Sscanf(value, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return fallback
}

func envOrDefaultBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	return fallback
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func firstText(content []mcp.Content) string {
	if len(content) == 0 {
		return "(no content)"
	}
	if text, ok := mcp.AsTextContent(content[0]); ok {
		return text.Text
	}
	return fmt.Sprintf("%T", content[0])
}
