package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Expired quests must be reported as a terminal error so the runner can drop
// them instead of retrying forever (the old infinite-spam bug).
func TestClassifyQuestFailureIsTerminal(t *testing.T) {
	terminal := []int{
		http.StatusNotFound,
		http.StatusGone,
		http.StatusForbidden,
		http.StatusUnprocessableEntity,
	}
	for _, status := range terminal {
		err := classifyQuestFailure(status, []byte(`{"message":"Quest is expired"}`), "q-1")
		if err == nil {
			t.Fatalf("status %d: expected terminal error, got nil", status)
		}
		if !errors.Is(err, ErrQuestExpired) {
			t.Fatalf("status %d: expected ErrQuestExpired, got %v", status, err)
		}
	}

	// Transient failures stay generic so they can be retried with backoff.
	for _, status := range []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusTooManyRequests} {
		if err := classifyQuestFailure(status, []byte("boom"), "q-1"); err != nil {
			t.Fatalf("status %d should be retryable, got %v", status, err)
		}
	}
}

func TestTruncateBody(t *testing.T) {
	long := make([]byte, 400)
	for i := range long {
		long[i] = 'a'
	}
	got := truncateBody(long)
	if len(got) > 190 {
		t.Fatalf("expected truncated body, got %d chars", len(got))
	}
	if got[len(got)-3:] != "..." {
		t.Fatalf("expected ellipsis suffix, got %q", got[len(got)-3:])
	}
}

// A 410 response from the heartbeat endpoint must surface as ErrQuestExpired.
func TestSendHeartbeatExpiredQuest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte(`{"message":"This quest has ended","code":0}`))
	}))
	defer srv.Close()

	c := NewClient("token", 123)
	c.apiBase = srv.URL

	_, err := c.SendHeartbeat(context.Background(), "quest-42", "call:0:1", false)
	if err == nil {
		t.Fatal("expected an error for an ended quest")
	}
	if !errors.Is(err, ErrQuestExpired) {
		t.Fatalf("expected ErrQuestExpired, got %v", err)
	}
	var expiredErr *QuestExpiredError
	if !errors.As(err, &expiredErr) {
		t.Fatalf("expected *QuestExpiredError, got %T", err)
	}
	if expiredErr.QuestID != "quest-42" {
		t.Fatalf("expected quest ID quest-42, got %s", expiredErr.QuestID)
	}
}

// A 404 from video-progress must also be terminal.
func TestSendVideoProgressExpiredQuest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Unknown Quest"}`))
	}))
	defer srv.Close()

	c := NewClient("token", 123)
	c.apiBase = srv.URL

	_, err := c.SendVideoProgress(context.Background(), "quest-99", 30)
	if !errors.Is(err, ErrQuestExpired) {
		t.Fatalf("expected ErrQuestExpired, got %v", err)
	}
}
