package web

import (
	"encoding/json"
	"fmt"
	"io"
	"m365-copilot2api/internal/auth"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestConsumerImportUsesEnterpriseAdminAuthentication(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("M365_DATA_DIR", dir)
	t.Setenv("M365_ADMIN_PASSWORD", "Test-Only.Strong.123!")
	t.Setenv("M365_ADMIN_PASSWORD_FILE", filepath.Join(dir, "admin-password"))
	t.Setenv("M365_API_KEYS", filepath.Join(dir, "keys.json"))
	t.Setenv("M365_MASTER_KEY", "test-only-key")
	server, err := New()
	if err != nil {
		t.Fatal(err)
	}
	app := server.Routes()
	denied := httptest.NewRecorder()
	app.ServeHTTP(denied, httptest.NewRequest("POST", "/api/accounts/consumer", strings.NewReader("{}")))
	if denied.Code != 401 {
		t.Fatalf("anonymous import status=%d", denied.Code)
	}
	site, client := adminTestClient(t, app)
	response := postJSON(t, client, site.URL+"/api/admin/login", `{"password":"Test-Only.Strong.123!"}`)
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal("login failed")
	}
	response = postJSON(t, client, site.URL+"/api/accounts/consumer", `{"consumer_account_id":"home:sample.subject","access_token":"never-expose-this-token","cookies":[{"name":"session","value":"never-expose-this-cookie","domain":"copilot.microsoft.com"}]}`)
	raw, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatalf("import: %d %s", response.StatusCode, raw)
	}
	response, _ = client.Get(site.URL + "/api/accounts")
	raw, _ = io.ReadAll(response.Body)
	response.Body.Close()
	if strings.Contains(string(raw), "never-expose") {
		t.Fatal("credentials leaked in account listing")
	}
	var listing struct {
		Accounts []map[string]any `json:"accounts"`
	}
	json.Unmarshal(raw, &listing)
	if len(listing.Accounts) != 1 || listing.Accounts[0]["provider"] != "consumer" {
		t.Fatal("account missing from enterprise store")
	}
	for _, path := range []string{"/api/usage", "/api/admin/keys", "/api/conversations", "/api/admin/settings", "/api/admin/models"} {
		response, _ = client.Get(site.URL + path)
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("enterprise route %s: %d", path, response.StatusCode)
		}
	}
}

func TestConsumerModesDoNotPretendToBeEnterpriseModels(t *testing.T) {
	t.Setenv("M365_CONSUMER_ONLY", "1")
	models := modelCatalog()
	if len(models) == 0 {
		t.Fatal("empty catalog")
	}
	for _, model := range models {
		if !strings.HasPrefix(model["id"].(string), "copilot") {
			t.Fatal("enterprise model in consumer catalog")
		}
	}
	tone, err := reasoningTone("copilot-reasoning", "")
	if err != nil || tone != "reasoning" {
		t.Fatal("wrong personal mode")
	}
}

func TestConsumerEnterpriseProtocolPipeline(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("M365_DATA_DIR", dir)
	t.Setenv("M365_ADMIN_PASSWORD", "Test-Only.Strong.123!")
	t.Setenv("M365_ADMIN_PASSWORD_FILE", filepath.Join(dir, "admin-password"))
	t.Setenv("M365_API_KEYS", filepath.Join(dir, "keys.json"))
	t.Setenv("M365_SESSION_CACHE", filepath.Join(dir, "sessions.json"))
	t.Setenv("M365_MASTER_KEY", "test-only-key")
	t.Setenv("M365_CONSUMER_ONLY", "1")
	t.Setenv("M365_PUBLIC_IDENTITY_POLICY", "false")
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"delta":"bridge-test-answer"}`)
		fmt.Fprintln(w, `{"done":true}`)
	}))
	defer endpoint.Close()
	t.Setenv("M365_CONSUMER_TRANSPORT_URL", endpoint.URL)
	t.Setenv("M365_CONSUMER_TRANSPORT_KEY", "test")
	server, err := New()
	if err != nil {
		t.Fatal(err)
	}
	_, err = server.tokens.ImportConsumer(auth.ConsumerCredentials{AccountID: "home:test.subject", AccessToken: "test-access-token", Cookies: []auth.ConsumerCookie{{Name: "session", Value: "test-cookie"}}})
	if err != nil {
		t.Fatal(err)
	}
	_, key, err := server.apiKeys.create("pipeline-test")
	if err != nil {
		t.Fatal(err)
	}
	for _, stream := range []bool{false, true} {
		for _, path := range []string{"/v1/chat/completions", "/v1/responses", "/v1/messages"} {
			body := map[string]any{"model": "copilot", "stream": stream, "messages": []map[string]any{{"role": "user", "content": "Say a short greeting"}}, "max_tokens": 200}
			if path == "/v1/responses" {
				delete(body, "messages")
				body["input"] = "Say a short greeting"
			}
			encoded, _ := json.Marshal(body)
			request := httptest.NewRequest("POST", path, strings.NewReader(string(encoded)))
			request.Header.Set("Authorization", "Bearer "+key)
			response := httptest.NewRecorder()
			server.Routes().ServeHTTP(response, request)
			observed := response.Body.String()
			if stream {
				var assembled strings.Builder
				for _, line := range strings.Split(observed, "\n") {
					if !strings.HasPrefix(line, "data: ") {
						continue
					}
					var event map[string]any
					if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event) != nil {
						continue
					}
					if choices, ok := event["choices"].([]any); ok && len(choices) > 0 {
						choice, _ := choices[0].(map[string]any)
						delta, _ := choice["delta"].(map[string]any)
						text, _ := delta["content"].(string)
						assembled.WriteString(text)
					} else if event["type"] == "response.output_text.delta" {
						text, _ := event["delta"].(string)
						assembled.WriteString(text)
					} else if event["type"] == "content_block_delta" {
						delta, _ := event["delta"].(map[string]any)
						text, _ := delta["text"].(string)
						assembled.WriteString(text)
					}
				}
				observed = assembled.String()
			}
			if response.Code != 200 || !strings.Contains(observed, "bridge-test-answer") {
				t.Fatalf("%s stream=%t: %d %s", path, stream, response.Code, response.Body.String())
			}
		}
	}
}
