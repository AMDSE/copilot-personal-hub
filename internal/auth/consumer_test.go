package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConsumerEncryptedImportAndReopen(t *testing.T) {
	t.Setenv("M365_MASTER_KEY", "test-only-encryption-key")
	path := filepath.Join(t.TempDir(), "accounts.json")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	credentials := ConsumerCredentials{AccountID: "home:test.subject", Email: "sample@example.test", AccessToken: "private-test-access-token", Cookies: []ConsumerCookie{{Name: "session", Value: "private-test-cookie", Domain: ".copilot.microsoft.com"}}}
	account, err := store.ImportConsumer(credentials)
	if err != nil {
		t.Fatal(err)
	}
	if account.Provider != "consumer" {
		t.Fatal("wrong provider")
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), credentials.AccessToken) || strings.Contains(string(raw), "private-test-cookie") {
		t.Fatal("credential stored without encryption")
	}
	reopened, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := reopened.EnsureValid(account.ID)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot ConsumerCredentials
	if json.Unmarshal([]byte(restored.ConsumerAuth), &snapshot) != nil || snapshot.AccessToken != credentials.AccessToken {
		t.Fatal("credentials not restored")
	}
	credentials.RefreshToken = "test-refresh-token"
	credentials.RefreshAccountID = "different-subject"
	if _, err := store.ImportConsumer(credentials); err == nil {
		t.Fatal("accepted mismatched refresh identity")
	}
	if len(store.List()) != 1 {
		t.Fatal("failed import mutated store")
	}
}
