package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"m365-copilot2api/internal/outbound"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ConsumerCookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain,omitempty"`
}

func (s *Store) refreshConsumer(account AccountToken) (AccountToken, error) {
	s.mu.Lock()
	if s.inflight == nil {
		s.inflight = map[string]*inflightRefresh{}
	}
	if flight, ok := s.inflight[account.ID]; ok {
		s.mu.Unlock()
		<-flight.done
		return flight.acc, flight.err
	}
	flight := &inflightRefresh{done: make(chan struct{})}
	s.inflight[account.ID] = flight
	s.mu.Unlock()
	refreshed, err := s.exchangeConsumer(account)
	s.mu.Lock()
	flight.acc, flight.err = refreshed, err
	delete(s.inflight, account.ID)
	close(flight.done)
	s.mu.Unlock()
	return refreshed, err
}

func (s *Store) exchangeConsumer(account AccountToken) (AccountToken, error) {
	var credentials ConsumerCredentials
	if json.Unmarshal([]byte(account.ConsumerAuth), &credentials) != nil {
		return account, errors.New("invalid consumer credentials")
	}
	if credentials.RefreshToken == "" {
		return account, errors.New("consumer token expired: import a new personal credential snapshot")
	}
	if !strings.EqualFold(credentials.AccountID, credentials.RefreshAccountID) {
		return account, errors.New("consumer refresh identity mismatch")
	}
	clients, err := outbound.New(account.BoundProxy)
	if err != nil {
		return account, errors.New("consumer refresh proxy invalid")
	}
	form := url.Values{"client_id": {credentials.RefreshClientID}, "grant_type": {"refresh_token"}, "refresh_token": {credentials.RefreshToken}, "scope": {credentials.RefreshScope + " openid profile offline_access"}}
	request, err := http.NewRequest(http.MethodPost, "https://login.microsoftonline.com/consumers/oauth2/v2.0/token", strings.NewReader(form.Encode()))
	if err != nil {
		return account, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Origin", "https://copilot.microsoft.com")
	client := *clients.HTTP
	client.Timeout = 20 * time.Second
	response, err := client.Do(request)
	if err != nil {
		return account, errors.New("consumer refresh connection failed; reimport credentials if needed")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return account, errors.New("consumer refresh rejected by Microsoft; reimport personal credentials")
	}
	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ClientInfo   string `json:"client_info"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result) != nil || result.AccessToken == "" || result.ExpiresIn <= 0 {
		return account, errors.New("invalid consumer refresh response")
	}
	info, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(result.ClientInfo, "="))
	if err != nil {
		return account, errors.New("consumer refresh missing identity")
	}
	var identity struct {
		UID  string `json:"uid"`
		UTID string `json:"utid"`
	}
	if json.Unmarshal(info, &identity) != nil || identity.UID == "" || identity.UTID == "" || !strings.EqualFold("home:"+identity.UID+"."+identity.UTID, credentials.AccountID) {
		return account, errors.New("consumer refresh identity mismatch")
	}
	credentials.AccessToken = result.AccessToken
	if result.RefreshToken != "" {
		credentials.RefreshToken = result.RefreshToken
	}
	credentials.ExpiresAt = float64(time.Now().Unix() + result.ExpiresIn)
	encoded, _ := json.Marshal(credentials)
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, current := range s.data.Accounts {
		if current.ID == account.ID {
			if current.ConsumerAuth != account.ConsumerAuth {
				return current, nil
			}
			current.ConsumerAuth = string(encoded)
			current.ExpiresAt = time.Unix(int64(credentials.ExpiresAt), 0)
			current.UpdatedAt = time.Now()
			current.Status = "online"
			s.data.Accounts[index] = current
			return current, s.saveLocked()
		}
	}
	return account, errors.New("consumer account removed during refresh")
}

type ConsumerCredentials struct {
	AccountID        string           `json:"consumer_account_id"`
	Email            string           `json:"email"`
	Username         string           `json:"username"`
	AccessToken      string           `json:"access_token"`
	IdentityType     string           `json:"identity_type"`
	Cookies          []ConsumerCookie `json:"cookies"`
	ExpiresAt        float64          `json:"expires_at,omitempty"`
	RefreshToken     string           `json:"refresh_token,omitempty"`
	RefreshClientID  string           `json:"refresh_token_client_id,omitempty"`
	RefreshScope     string           `json:"refresh_token_scope,omitempty"`
	RefreshAccountID string           `json:"refresh_token_account_id,omitempty"`
}

func (s *Store) ImportConsumer(credentials ConsumerCredentials) (AccountToken, error) {
	subject := strings.TrimSpace(credentials.AccountID)
	if subject == "" || len(subject) > 256 || len(credentials.AccessToken) < 10 || len(credentials.AccessToken) > 32768 || len(credentials.Cookies) == 0 || len(credentials.Cookies) > 500 {
		return AccountToken{}, errors.New("personal credentials require account identity, access_token and cookies")
	}
	if len(credentials.Email) > 320 || len(credentials.Username) > 320 {
		return AccountToken{}, errors.New("name too long")
	}
	for _, cookie := range credentials.Cookies {
		domain := strings.ToLower(strings.TrimPrefix(cookie.Domain, "."))
		if cookie.Name == "" || len(cookie.Name) > 256 || len(cookie.Value) > 32768 || (domain != "" && domain != "copilot.microsoft.com" && domain != "microsoft.com" && domain != "live.com" && !strings.HasSuffix(domain, ".live.com")) {
			return AccountToken{}, errors.New("invalid consumer cookie")
		}
	}
	if credentials.RefreshToken != "" && (!strings.EqualFold(subject, credentials.RefreshAccountID) || credentials.RefreshClientID == "" || credentials.RefreshScope == "" || len(credentials.RefreshToken) > 8192) {
		return AccountToken{}, errors.New("refresh token binding does not match the personal account")
	}
	data, err := json.Marshal(credentials)
	if err != nil {
		return AccountToken{}, err
	}
	digest := sha256.Sum256([]byte(strings.ToLower(subject)))
	id := "consumer-" + hex.EncodeToString(digest[:12])
	expiry := time.Time{}
	if credentials.ExpiresAt > 0 {
		expiry = time.Unix(int64(credentials.ExpiresAt), 0)
	}
	account := AccountToken{ID: id, OID: id, TID: "consumer", Provider: "consumer", Email: credentials.Email, DisplayName: credentials.Username, Status: "online", AccessToken: "consumer-managed", ConsumerAuth: string(data), ExpiresAt: expiry, UpdatedAt: time.Now()}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, existing := range s.data.Accounts {
		if existing.ID == id {
			account.ScheduleDisabled, account.WebSearchDisabled = existing.ScheduleDisabled, existing.WebSearchDisabled
			account.SystemPrompt, account.BoundProxy = existing.SystemPrompt, existing.BoundProxy
			s.data.Accounts[index] = account
			return account, s.saveLocked()
		}
	}
	s.data.Accounts = append(s.data.Accounts, account)
	return account, s.saveLocked()
}
