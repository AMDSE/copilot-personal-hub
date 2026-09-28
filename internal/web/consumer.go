package web

import (
	"encoding/json"
	"m365-copilot2api/internal/auth"
	"m365-copilot2api/internal/chathub"
	"net/http"
	"os"
	"strings"
)

func (s *Server) withConsumerAccount(id string, account chathub.Account) chathub.Account {
	if stored, ok := s.tokens.Get(id); ok && stored.Provider == "consumer" {
		account.Provider, account.ConsumerAuth, account.ProxyURL = stored.Provider, stored.ConsumerAuth, stored.BoundProxy
	}
	return account
}

func (s *Server) importConsumer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeOpenAIError(w, 405, "invalid_request_error", "POST required")
		return
	}
	var credentials auth.ConsumerCredentials
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&credentials) != nil {
		writeOpenAIError(w, 400, "invalid_request_error", "invalid credential JSON")
		return
	}
	account, err := s.tokens.ImportConsumer(credentials)
	if err != nil {
		writeOpenAIError(w, 400, "invalid_request_error", err.Error())
		return
	}
	jsonOut(w, map[string]any{"status": "imported", "id": account.ID, "provider": "consumer", "message": "Credentials saved; real chat has not yet been verified"})
}

func (s *Server) consumerScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET required", 405)
		return
	}
	data, err := webFS.ReadFile("web/consumer-capture.user.js")
	if err != nil {
		http.Error(w, "script unavailable", 500)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}

func consumerMode(model string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "copilot", "copilot-smart", "auto":
		return "smart", true
	case "copilot-reasoning", "copilot-thinking":
		return "reasoning", true
	case "copilot-chat":
		return "chat", true
	case "copilot-search":
		return "search", true
	case "copilot-research":
		return "research", true
	case "copilot-study":
		return "study", true
	case "copilot-coco":
		return "coco", true
	}
	return "", false
}

func consumerModelSpecs() []modelSpec {
	modes := []string{"copilot", "copilot-reasoning", "copilot-thinking", "copilot-chat", "copilot-search", "copilot-research", "copilot-study", "copilot-coco"}
	result := make([]modelSpec, 0, len(modes))
	for _, id := range modes {
		result = append(result, modelSpec{ID: id, Owner: "microsoft-consumer", DisplayName: id, Tools: true})
	}
	return result
}

func providerDefaultMappings() []modelMapping {
	if os.Getenv("M365_CONSUMER_ONLY") != "1" {
		return append([]modelMapping(nil), defaultModelMappings...)
	}
	result := []modelMapping{}
	for _, spec := range consumerModelSpecs() {
		mode, _ := consumerMode(spec.ID)
		result = append(result, modelMapping{PublicModel: spec.ID, UpstreamTone: mode, DisplayName: spec.ID})
	}
	return result
}

func consumerMappedTone(model string, mappings []modelMapping) (string, bool) {
	mapping, ok := configuredModelMapping(model, mappings)
	if !ok {
		return "", false
	}
	switch mapping.UpstreamTone {
	case "smart", "reasoning", "chat", "search", "research", "study", "coco":
		return mapping.UpstreamTone, true
	}
	return "", false
}
