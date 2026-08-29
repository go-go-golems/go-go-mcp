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

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type bearerRoundTripper struct {
	token string
	next  http.RoundTripper
}

func (r bearerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header.Set("Authorization", "Bearer "+r.token)
	return r.next.RoundTrip(cloned)
}

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
		serverURL = flag.String("server-url", envOrDefault("SMAILNAIL_SERVER_URL", "http://127.0.0.1:8081/mcp"), "Merged MCP endpoint")
		tokenURL  = flag.String("token-url", envOrDefault("SMAILNAIL_TOKEN_URL", "http://127.0.0.1:18080/realms/smailnail-dev/protocol/openid-connect/token"), "OIDC token endpoint")
		clientID  = flag.String("client-id", envOrDefault("SMAILNAIL_CLIENT_ID", "smailnail-mcp"), "OIDC client ID")
		username  = flag.String("username", envOrDefault("SMAILNAIL_USERNAME", "alice"), "OIDC username")
		password  = flag.String("password", envOrDefault("SMAILNAIL_PASSWORD", "secret"), "OIDC password")
		accountID = flag.String("account-id", os.Getenv("SMAILNAIL_ACCOUNT_ID"), "Stored IMAP account id to use")
		timeout   = flag.Duration("timeout", 30*time.Second, "Overall timeout")
	)
	flag.Parse()

	if *accountID == "" {
		fatalf("account-id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	token, err := fetchPasswordGrantToken(ctx, *tokenURL, *clientID, *username, *password)
	if err != nil {
		fatalf("fetch token: %v", err)
	}

	httpClient := &http.Client{
		Timeout: *timeout,
		Transport: bearerRoundTripper{
			token: token,
			next:  http.DefaultTransport,
		},
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "smailnail-merged-smoke", Version: "dev"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             *serverURL,
		HTTPClient:           httpClient,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		fatalf("connect and initialize client: %v", err)
	}
	defer func() { _ = session.Close() }()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "executeIMAPJS",
		Arguments: map[string]any{
			"code": buildAccountCode(*accountID),
		},
	})
	if err != nil {
		fatalf("call executeIMAPJS: %v", err)
	}
	if result.IsError {
		fatalf("tool returned error: %s", firstText(result.Content))
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		fatalf("expected text result, got %T", result.Content[0])
	}

	var decoded executeResponse
	if err := json.Unmarshal([]byte(text.Text), &decoded); err != nil {
		fatalf("decode tool response: %v", err)
	}
	if !decoded.Success {
		fatalf("tool reported failure: %v", decoded.Error)
	}

	output := map[string]any{
		"serverURL": *serverURL,
		"accountID": *accountID,
		"value":     json.RawMessage(decoded.Value),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		fatalf("encode output: %v", err)
	}
}

func fetchPasswordGrantToken(ctx context.Context, tokenURL, clientID, username, password string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", clientID)
	form.Set("username", username)
	form.Set("password", password)

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
		return "", fmt.Errorf("access token missing")
	}
	return decoded.AccessToken, nil
}

func buildAccountCode(accountID string) string {
	return fmt.Sprintf(`
const smailnail = require("smailnail");
const svc = smailnail.newService();
const session = svc.connect({ accountId: %q });
const result = { mailbox: session.mailbox };
session.close();
result;
`, accountID)
}

func firstText(content []mcp.Content) string {
	if len(content) == 0 {
		return "(no content)"
	}
	if text, ok := content[0].(*mcp.TextContent); ok {
		return text.Text
	}
	return fmt.Sprintf("%T", content[0])
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
