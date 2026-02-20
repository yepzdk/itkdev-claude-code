package console

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/itk-dev/itkdev-claude-code/internal/db"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	database, err := db.OpenInMemory(logger)
	if err != nil {
		t.Fatalf("OpenInMemory: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return NewWithDB(0, logger, database)
}

func doRequest(t *testing.T, srv *Server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	return rr
}

func TestHealth(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "GET", "/health", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want ok", resp["status"])
	}
}

func TestObservationRoundTrip(t *testing.T) {
	srv := testServer(t)

	// Create
	rr := doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1",
		"type":       "discovery",
		"title":      "test obs",
		"text":       "some text about authentication",
		"project":    "proj",
		"metadata":   "{}",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var created map[string]any
	json.NewDecoder(rr.Body).Decode(&created)
	id := int64(created["id"].(float64))

	// Get
	rr = doRequest(t, srv, "GET", fmt.Sprintf("/api/observations/%d", id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("get status = %d", rr.Code)
	}

	// Search
	rr = doRequest(t, srv, "GET", "/api/observations/search?q=authentication", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("search status = %d", rr.Code)
	}
	var results []any
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) != 1 {
		t.Errorf("search returned %d results, want 1", len(results))
	}
}

func TestRecentObservations(t *testing.T) {
	srv := testServer(t)

	// Empty database returns empty list
	rr := doRequest(t, srv, "GET", "/api/observations/recent", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	// Create some observations
	for i := 0; i < 3; i++ {
		doRequest(t, srv, "POST", "/api/observations", map[string]string{
			"session_id": "s1",
			"type":       "discovery",
			"title":      fmt.Sprintf("obs %d", i),
			"text":       fmt.Sprintf("text %d", i),
			"project":    "proj",
		})
	}

	// Fetch recent
	rr = doRequest(t, srv, "GET", "/api/observations/recent", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var results []any
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
	}

	// With limit
	rr = doRequest(t, srv, "GET", "/api/observations/recent?limit=2", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) != 2 {
		t.Errorf("got %d results with limit=2, want 2", len(results))
	}

	// With project filter
	doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1",
		"type":       "feature",
		"title":      "other proj obs",
		"text":       "text",
		"project":    "other",
	})
	rr = doRequest(t, srv, "GET", "/api/observations/recent?project=other", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) != 1 {
		t.Errorf("got %d results with project=other, want 1", len(results))
	}
}

func TestObservationNotFound(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "GET", "/api/observations/99999", nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestObservationInvalidID(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "GET", "/api/observations/abc", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSearchWithFilters(t *testing.T) {
	srv := testServer(t)

	// Create observations with different types and projects
	doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1", "type": "bugfix", "title": "auth bug",
		"text": "Fixed authentication", "project": "backend",
	})
	doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1", "type": "feature", "title": "auth feature",
		"text": "Added authentication rate limiting", "project": "backend",
	})
	doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1", "type": "bugfix", "title": "ui auth bug",
		"text": "Fixed authentication dialog", "project": "frontend",
	})

	// Search with type filter
	rr := doRequest(t, srv, "GET", "/api/observations/search?q=authentication&type=bugfix", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var results []any
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) != 2 {
		t.Errorf("got %d results with type=bugfix, want 2", len(results))
	}

	// Search with project filter
	rr = doRequest(t, srv, "GET", "/api/observations/search?q=authentication&project=frontend", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) != 1 {
		t.Errorf("got %d results with project=frontend, want 1", len(results))
	}
}

func TestSearchMissingQuery(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "GET", "/api/observations/search", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSessionRoundTrip(t *testing.T) {
	srv := testServer(t)

	// Create
	rr := doRequest(t, srv, "POST", "/api/sessions", map[string]string{
		"id":      "sess-1",
		"project": "myproj",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rr.Code, rr.Body.String())
	}

	// Get
	rr = doRequest(t, srv, "GET", "/api/sessions/sess-1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("get status = %d", rr.Code)
	}

	// List
	rr = doRequest(t, srv, "GET", "/api/sessions", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status = %d", rr.Code)
	}

	// End
	rr = doRequest(t, srv, "POST", "/api/sessions/sess-1/end", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("end status = %d", rr.Code)
	}
}

func TestSessionIncrementMessageCount(t *testing.T) {
	srv := testServer(t)

	doRequest(t, srv, "POST", "/api/sessions", map[string]string{
		"id": "sess-1", "project": "proj",
	})

	rr := doRequest(t, srv, "POST", "/api/sessions/sess-1/message-count", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}

	// Verify count increased
	rr = doRequest(t, srv, "GET", "/api/sessions/sess-1", nil)
	var sess map[string]any
	json.NewDecoder(rr.Body).Decode(&sess)
	if count := sess["MessageCount"].(float64); count != 1 {
		t.Errorf("MessageCount = %v, want 1", count)
	}
}

func TestSessionCleanup(t *testing.T) {
	srv := testServer(t)

	rr := doRequest(t, srv, "POST", "/api/sessions/cleanup", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rr.Body).Decode(&resp)
	if _, ok := resp["cleaned"]; !ok {
		t.Error("response missing 'cleaned' field")
	}
}

func TestSessionNotFound(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "GET", "/api/sessions/nonexistent", nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestSummaryRoundTrip(t *testing.T) {
	srv := testServer(t)

	doRequest(t, srv, "POST", "/api/sessions", map[string]string{
		"id": "sess-1",
	})

	rr := doRequest(t, srv, "POST", "/api/summaries", map[string]string{
		"session_id": "sess-1",
		"text":       "did some work",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, srv, "GET", "/api/summaries/recent", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("recent status = %d", rr.Code)
	}
}

func TestContextInject(t *testing.T) {
	srv := testServer(t)

	// Create a session and some observations
	doRequest(t, srv, "POST", "/api/sessions", map[string]string{
		"id": "sess-1", "project": "proj",
	})
	doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "sess-1", "type": "discovery",
		"title": "Config system", "text": "Config loads from YAML",
	})
	doRequest(t, srv, "POST", "/api/summaries", map[string]string{
		"session_id": "sess-1", "text": "Built config subsystem",
	})

	rr := doRequest(t, srv, "GET", "/api/context/inject?session_id=sess-1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rr.Body).Decode(&resp)
	ctx, ok := resp["context"].(string)
	if !ok || ctx == "" {
		t.Error("expected non-empty context string")
	}
}

func TestContextInjectEmpty(t *testing.T) {
	srv := testServer(t)

	rr := doRequest(t, srv, "GET", "/api/context/inject?session_id=nonexistent", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp map[string]any
	json.NewDecoder(rr.Body).Decode(&resp)
	ctx := resp["context"].(string)
	if ctx != "" {
		t.Errorf("expected empty context for nonexistent session, got %q", ctx)
	}
}

func TestPlanRoundTrip(t *testing.T) {
	srv := testServer(t)

	// Create
	rr := doRequest(t, srv, "POST", "/api/plans", map[string]string{
		"path":       "docs/plans/test.md",
		"session_id": "sess-1",
		"status":     "PENDING",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var created map[string]any
	json.NewDecoder(rr.Body).Decode(&created)
	id := int64(created["id"].(float64))

	// List plans
	rr = doRequest(t, srv, "GET", "/api/plans", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status = %d", rr.Code)
	}
	var plans []any
	json.NewDecoder(rr.Body).Decode(&plans)
	if len(plans) != 1 {
		t.Errorf("list returned %d plans, want 1", len(plans))
	}

	// Get by path
	rr = doRequest(t, srv, "GET", "/api/plans/by-path?path=docs/plans/test.md", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("get status = %d", rr.Code)
	}

	// Update status
	rr = doRequest(t, srv, "PATCH", fmt.Sprintf("/api/plans/%d/status", id), map[string]string{
		"status": "COMPLETE",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", rr.Code, rr.Body.String())
	}
}

func TestPlanSSEEvents(t *testing.T) {
	srv := testServer(t)

	// Subscribe to SSE events before creating a plan
	ch := srv.sse.Subscribe()
	defer srv.sse.Unsubscribe(ch)

	// Create plan — should broadcast a "plan" event
	rr := doRequest(t, srv, "POST", "/api/plans", map[string]string{
		"path":       "docs/plans/sse-test.md",
		"session_id": "sess-sse",
		"status":     "PENDING",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var created map[string]any
	json.NewDecoder(rr.Body).Decode(&created)
	id := int64(created["id"].(float64))

	// Read the create event
	e := <-ch
	if e.Type != "plan" {
		t.Errorf("event type = %q, want plan", e.Type)
	}
	var eventData map[string]any
	json.Unmarshal([]byte(e.Data), &eventData)
	if eventData["action"] != "created" {
		t.Errorf("action = %q, want created", eventData["action"])
	}
	if eventData["path"] != "docs/plans/sse-test.md" {
		t.Errorf("path = %q, want docs/plans/sse-test.md", eventData["path"])
	}

	// Update status — should broadcast another "plan" event
	rr = doRequest(t, srv, "PATCH", fmt.Sprintf("/api/plans/%d/status", id), map[string]string{
		"status": "IN_PROGRESS",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("update status = %d", rr.Code)
	}

	e = <-ch
	if e.Type != "plan" {
		t.Errorf("event type = %q, want plan", e.Type)
	}
	json.Unmarshal([]byte(e.Data), &eventData)
	if eventData["action"] != "status_changed" {
		t.Errorf("action = %q, want status_changed", eventData["action"])
	}
	if eventData["status"] != "IN_PROGRESS" {
		t.Errorf("status = %q, want IN_PROGRESS", eventData["status"])
	}
}

func TestPlanListWithLimit(t *testing.T) {
	srv := testServer(t)

	// Create multiple plans
	for i := 0; i < 3; i++ {
		doRequest(t, srv, "POST", "/api/plans", map[string]string{
			"path":       fmt.Sprintf("docs/plans/plan-%d.md", i),
			"session_id": "sess-1",
			"status":     "PENDING",
		})
	}

	// List with limit
	rr := doRequest(t, srv, "GET", "/api/plans?limit=2", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status = %d", rr.Code)
	}
	var plans []any
	json.NewDecoder(rr.Body).Decode(&plans)
	if len(plans) != 2 {
		t.Errorf("list returned %d plans with limit=2, want 2", len(plans))
	}
}

func TestPlanNotFound(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "GET", "/api/plans/by-path?path=nonexistent.md", nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestPlanMissingPath(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "GET", "/api/plans/by-path", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestTimeline(t *testing.T) {
	srv := testServer(t)

	for i := 0; i < 5; i++ {
		doRequest(t, srv, "POST", "/api/observations", map[string]string{
			"session_id": "s1",
			"title":      fmt.Sprintf("obs %d", i),
			"text":       "text",
		})
	}

	rr := doRequest(t, srv, "GET", "/api/observations/timeline/3", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("timeline status = %d, body = %s", rr.Code, rr.Body.String())
	}
}
