package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckUpdateMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"tag_name": "v1.2.0",
			"html_url": "https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.2.0",
			"name": "v1.2.0 Release"
		}`))
	}))
	defer server.Close()

	ctx := context.Background()
	rel, hasUpdate := checkUpdateFromURL(ctx, server.URL, "1.1.0")
	if !hasUpdate {
		t.Fatalf("expected hasUpdate to be true for 1.1.0 vs 1.2.0")
	}
	if rel.TagName != "v1.2.0" {
		t.Fatalf("expected tag v1.2.0, got %s", rel.TagName)
	}

	_, hasUpdateSame := checkUpdateFromURL(ctx, server.URL, "1.2.0")
	if hasUpdateSame {
		t.Fatalf("expected hasUpdate to be false for 1.2.0 vs 1.2.0")
	}
}
