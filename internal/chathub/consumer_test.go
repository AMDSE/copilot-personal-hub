package chathub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConsumerBridgeStreamsAndPreservesPrompt(t *testing.T) {
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-bridge-secret" {
			t.Error("missing private auth")
		}
		var payload map[string]any
		if json.NewDecoder(r.Body).Decode(&payload) != nil {
			t.Error("invalid payload")
		}
		if payload["prompt"] != "full conversation history" || payload["mode"] != "reasoning" {
			t.Error("prompt or mode lost")
		}
		fmt.Fprintln(w, `{"delta":"Hello "}`)
		fmt.Fprintln(w, `{"delta":"world"}`)
		fmt.Fprintln(w, `{"done":true}`)
	}))
	defer endpoint.Close()
	t.Setenv("M365_CONSUMER_TRANSPORT_URL", endpoint.URL)
	t.Setenv("M365_CONSUMER_TRANSPORT_KEY", "test-bridge-secret")
	var streamed strings.Builder
	result, err := NewClient().ChatWithDelta(context.Background(), Account{Provider: "consumer", ConsumerAuth: `{"access_token":"not-real","cookies":[]}`}, Request{Text: "full conversation history", Tone: "reasoning"}, func(delta string) error { streamed.WriteString(delta); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "Hello world" || streamed.String() != result.Text {
		t.Fatal("stream differs from result")
	}
	streamed.Reset()
	_, err = NewClient().ChatWithEvents(context.Background(), Account{Provider: "consumer", ConsumerAuth: "{}"}, Request{Text: "full conversation history", Tone: "reasoning"}, func(event StreamEvent) error { streamed.WriteString(event.Text); return nil })
	if err != nil || streamed.String() != "Hello world" {
		t.Fatalf("duplicated event stream: %q %v", streamed.String(), err)
	}
}

func TestConsumerBridgeRejectsTruncatedStream(t *testing.T) {
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, `{"delta":"incomplete"}`) }))
	defer endpoint.Close()
	t.Setenv("M365_CONSUMER_TRANSPORT_URL", endpoint.URL)
	t.Setenv("M365_CONSUMER_TRANSPORT_KEY", "test")
	_, err := NewClient().Chat(context.Background(), Account{Provider: "consumer", ConsumerAuth: "{}"}, Request{Text: "test"})
	if err == nil {
		t.Fatal("truncated stream reported success")
	}
}
