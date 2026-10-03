package captcha

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"discord-quest-completer/pkg/api"
)

func TestGetOutboundIP(t *testing.T) {
	ip := GetOutboundIP()
	if ip == "" {
		t.Fatalf("expected non-empty outbound IP, got empty")
	}
}

func TestPortalSubmission(t *testing.T) {
	challenge := &api.CaptchaRequiredError{
		QuestID:        "test-quest-123",
		CaptchaService: "turnstile",
		CaptchaSitekey: "1x00000000000000000000AA",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	port := 18080

	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		tok, err := StartPortal(ctx, port, challenge, "Test Quest")
		if err != nil {
			errChan <- err
		} else {
			resultChan <- tok
		}
	}()

	// Wait for server to start listening
	time.Sleep(100 * time.Millisecond)

	// Simulate browser submitting solved token
	resp, err := http.Post("http://127.0.0.1:18080/submit", "application/json", strings.NewReader(`{"token":"solved-captcha-token-abc"}`))
	if err != nil {
		t.Fatalf("failed to post solve token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	select {
	case tok := <-resultChan:
		if tok != "solved-captcha-token-abc" {
			t.Fatalf("expected token 'solved-captcha-token-abc', got '%s'", tok)
		}
	case err := <-errChan:
		t.Fatalf("portal returned error: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for portal to complete")
	}
}
