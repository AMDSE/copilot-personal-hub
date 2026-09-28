package chathub

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func (c *Client) consumerChat(ctx context.Context, account Account, request Request, onDelta func(string) error, onEvent StreamHandler) (Result, error) {
	endpoint := strings.TrimRight(os.Getenv("M365_CONSUMER_TRANSPORT_URL"), "/")
	secret := os.Getenv("M365_CONSUMER_TRANSPORT_KEY")
	if endpoint == "" || secret == "" {
		return Result{}, errors.New("consumer transport is not configured")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "http" || (parsed.Hostname() != "consumer" && parsed.Hostname() != "127.0.0.1" && parsed.Hostname() != "localhost") {
		return Result{}, errors.New("consumer transport must be local or the private consumer service")
	}
	var credentials map[string]any
	if err := json.Unmarshal([]byte(account.ConsumerAuth), &credentials); err != nil {
		return Result{}, errors.New("invalid stored personal credentials")
	}
	images := []string{}
	for _, attachment := range request.Attachments {
		if attachment.Type != "image" {
			return Result{}, errors.New("consumer currently accepts image attachments only")
		}
		if len(images) >= maxAttachments {
			return Result{}, errors.New("too many images")
		}
		data := attachment.URL
		if !strings.HasPrefix(data, "data:") {
			if err := validateRemoteDownloadURL(data); err != nil {
				return Result{}, err
			}
			download, err := http.NewRequestWithContext(ctx, http.MethodGet, data, nil)
			if err != nil {
				return Result{}, err
			}
			response, err := c.downloadClient().Do(download)
			if err != nil {
				return Result{}, err
			}
			raw, readErr := io.ReadAll(io.LimitReader(response.Body, (maxAttachmentMiB<<20)+1))
			response.Body.Close()
			if readErr != nil || response.StatusCode != 200 || len(raw) > maxAttachmentMiB<<20 {
				return Result{}, errors.New("image download failed or exceeds size limit")
			}
			mime := response.Header.Get("Content-Type")
			if !strings.HasPrefix(mime, "image/") {
				return Result{}, errors.New("attachment is not an image")
			}
			data = "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw)
		}
		if len(data) > 15<<20 {
			return Result{}, errors.New("image exceeds size limit")
		}
		images = append(images, data)
	}
	var history strings.Builder
	for _, previous := range request.PreviousMessages {
		history.WriteString(previous.Author + ": " + previous.Description + "\n")
	}
	prompt := history.String() + request.Text
	if len(request.Tools) > 0 {
		tools, _ := json.Marshal(request.Tools)
		prompt = "Available tools (emit JSON tool calls when needed): " + string(tools) + "\n" + prompt
	}
	mode := strings.ToLower(request.Tone)
	switch mode {
	case "smart", "reasoning", "chat", "search", "research", "study", "coco":
	default:
		mode = "smart"
	}
	if request.DisableWebSearch && mode == "search" {
		mode = "chat"
	}
	payload, err := json.Marshal(map[string]any{"credentials": credentials, "prompt": prompt, "mode": mode, "proxy": account.ProxyURL, "images": images})
	if err != nil {
		return Result{}, err
	}
	upstream, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/turn", bytes.NewReader(payload))
	if err != nil {
		return Result{}, err
	}
	upstream.Header.Set("Authorization", "Bearer "+secret)
	upstream.Header.Set("Content-Type", "application/json")
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	response, err := (&http.Client{Transport: transport}).Do(upstream)
	if err != nil {
		return Result{}, errors.New("personal transport connection failed")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return Result{}, fmt.Errorf("personal transport returned HTTP %d", response.StatusCode)
	}
	result := Result{ConversationID: request.ConversationID, SessionID: request.SessionID, RequestID: uuid.NewString()}
	if result.ConversationID == "" {
		result.ConversationID = uuid.NewString()
	}
	if result.SessionID == "" {
		result.SessionID = uuid.NewString()
	}
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	completed := false
	var output strings.Builder
	for scanner.Scan() {
		var event struct {
			Delta string `json:"delta"`
			Error string `json:"error"`
			Done  bool   `json:"done"`
		}
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			return Result{}, errors.New("invalid personal transport event")
		}
		if event.Error != "" {
			return Result{}, fmt.Errorf("consumer: %s", event.Error)
		}
		if event.Done {
			completed = true
			break
		}
		if event.Delta != "" {
			if output.Len()+len(event.Delta) > 32<<20 {
				return Result{}, errors.New("personal response exceeds limit")
			}
			output.WriteString(event.Delta)
			if onDelta != nil {
				if err := onDelta(event.Delta); err != nil {
					return Result{}, err
				}
			}
			if onEvent != nil && onDelta == nil {
				if err := onEvent(StreamEvent{Kind: "text", Text: event.Delta}); err != nil {
					return Result{}, err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if !completed {
		return Result{}, errors.New("personal stream ended before completion")
	}
	result.Text = output.String()
	result.Timestamps.LastTokenReceived = time.Now().UTC().Format(time.RFC3339Nano)
	return result, nil
}
