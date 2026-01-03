package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_EmptyToken(t *testing.T) {
	_, err := NewClient("", "")
	if err == nil {
		t.Error("NewClient() with empty token should return error")
	}
}

func TestNewClient_WithToken(t *testing.T) {
	// Create a mock server that returns a valid response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"id":123,"is_bot":true,"first_name":"TestBot","username":"test_bot"}}`))
	}))
	defer server.Close()

	token := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	client, err := NewClient(token, server.URL)
	if err != nil {
		t.Errorf("NewClient() with valid token failed: %v", err)
		return
	}
	if client == nil {
		t.Error("NewClient() returned nil client")
	}
}

func TestClient_GetAPIDomain(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"id":123,"is_bot":true,"first_name":"TestBot","username":"test_bot"}}`))
	}))
	defer server.Close()

	token := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	apiDomain := server.URL

	client, err := NewClient(token, apiDomain)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	if got := client.GetAPIDomain(); got != apiDomain {
		t.Errorf("GetAPIDomain() = %v, want %v", got, apiDomain)
	}
}

func TestClient_SetDebug(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"id":123,"is_bot":true,"first_name":"TestBot","username":"test_bot"}}`))
	}))
	defer server.Close()

	token := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	client, err := NewClient(token, server.URL)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	// Test enabling debug
	client.SetDebug(true)
	if !client.BotAPI.Debug {
		t.Error("SetDebug(true) did not enable debug mode")
	}

	// Test disabling debug
	client.SetDebug(false)
	if client.BotAPI.Debug {
		t.Error("SetDebug(false) did not disable debug mode")
	}
}

func TestNewDefaultHTTPClient(t *testing.T) {
	client := NewDefaultHTTPClient()
	if client == nil {
		t.Fatal("NewDefaultHTTPClient() returned nil")
	}
	if client.Timeout == 0 {
		t.Error("NewDefaultHTTPClient() timeout is not set")
	}
	if client.Transport == nil {
		t.Error("NewDefaultHTTPClient() transport is not set")
	}
}
