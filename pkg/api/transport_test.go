package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestRegionInfoDrivesLocaleHeaders(t *testing.T) {
	var gotLocale, gotTimezone, clientLocale, systemLocale string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLocale = r.Header.Get("X-Discord-Locale")
		gotTimezone = r.Header.Get("X-Discord-Timezone")
		raw, _ := base64.StdEncoding.DecodeString(r.Header.Get("X-Super-Properties"))
		var props map[string]interface{}
		if err := json.Unmarshal(raw, &props); err == nil {
			clientLocale, _ = props["client_locale"].(string)
			systemLocale, _ = props["system_locale"].(string)
		}
		_, _ = w.Write([]byte(`{"quests":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", 500000)
	client.SetAPIBaseForTesting(server.URL)
	ctx := WithRegionInfo(context.Background(), RegionInfo{Code: "US", Locale: "ja", Timezone: "Asia/Tokyo"})
	if _, err := client.FetchQuests(ctx); err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if gotLocale != "ja" {
		t.Fatalf("expected X-Discord-Locale ja, got %q", gotLocale)
	}
	if gotTimezone != "Asia/Tokyo" {
		t.Fatalf("expected X-Discord-Timezone Asia/Tokyo, got %q", gotTimezone)
	}
	if clientLocale != "ja" {
		t.Fatalf("expected super properties client_locale ja, got %q", clientLocale)
	}
	// The OS locale is a property of the device: using a different egress must
	// never claim the machine itself switched language.
	if systemLocale != "en-US" {
		t.Fatalf("expected super properties system_locale to stay en-US, got %q", systemLocale)
	}
}

func TestRegionProxyRoutesOnlyRegionRequests(t *testing.T) {
	var proxied int32
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&proxied, 1)
		_, _ = w.Write([]byte(`{"quests":[{"id":"q-proxy"}]}`))
	}))
	defer proxyServer.Close()

	client := NewClient("test-token", 500000)
	client.SetAPIBaseForTesting("http://127.0.0.1:1")
	if err := client.SetRegionProxy("US", proxyServer.URL); err != nil {
		t.Fatalf("unexpected proxy setup error: %v", err)
	}

	regionCtx := WithRegionInfo(context.Background(), RegionInfo{Code: "US", Locale: "en-US", Timezone: "America/New_York"})
	quests, err := client.FetchQuests(regionCtx)
	if err != nil {
		t.Fatalf("regional request through the proxy failed: %v", err)
	}
	if len(quests) != 1 || quests[0].ID != "q-proxy" {
		t.Fatalf("expected the proxy response, got %+v", quests)
	}
	if atomic.LoadInt32(&proxied) == 0 {
		t.Fatal("proxy server was never used")
	}

	if _, err := client.FetchQuests(context.Background()); err == nil {
		t.Fatal("expected the direct request to fail without a proxy")
	}
}

func TestSetRegionProxyValidation(t *testing.T) {
	client := NewClient("test-token", 500000)
	if err := client.SetRegionProxy("US", "ftp://127.0.0.1:1"); err == nil {
		t.Fatal("expected an unsupported scheme to be rejected")
	}
	if err := client.SetRegionProxy("US", "socks5://127.0.0.1:1080"); err != nil {
		t.Fatalf("unexpected socks5 proxy error: %v", err)
	}
	if !client.HasRegionProxy("us") {
		t.Fatal("expected the proxy to be registered case-insensitively")
	}
	client.RemoveRegionProxy("US")
	if client.HasRegionProxy("US") {
		t.Fatal("expected the proxy to be removed")
	}
}
