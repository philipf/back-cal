package fetch_test

import (
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/philipf/back-cal/internal/fetch"
)

func TestBMP_Success(t *testing.T) {
	payload := []byte("BM fake bmp bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-radiator-slug") != "test-slug" {
			t.Errorf("missing or wrong slug header: %q", r.Header.Get("x-radiator-slug"))
		}
		if r.Header.Get("x-radiator-token") != "test-token" {
			t.Errorf("missing or wrong token header: %q", r.Header.Get("x-radiator-token"))
		}
		if r.Header.Get("Accept") != "image/bmp" {
			t.Errorf("wrong Accept header: %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "image/bmp")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer srv.Close()

	got, err := fetch.BMP(fetch.Config{URL: srv.URL, Slug: "test-slug", Token: "test-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("got %q, want %q", got, payload)
	}
}

func TestBMP_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := fetch.BMP(fetch.Config{URL: srv.URL, Slug: "s", Token: "t"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*fetch.APIError)
	if !ok {
		t.Fatalf("expected *fetch.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestBMP_GzipResponse(t *testing.T) {
	payload := []byte("BM gzip bmp bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/bmp")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		gz := gzip.NewWriter(w)
		gz.Write(payload)
		gz.Close()
	}))
	defer srv.Close()

	got, err := fetch.BMP(fetch.Config{URL: srv.URL, Slug: "s", Token: "t"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("got %q, want %q", got, payload)
	}
}

func TestBMP_BodyTruncatedIn400Error(t *testing.T) {
	longBody := make([]byte, 500)
	for i := range longBody {
		longBody[i] = 'x'
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(longBody)
	}))
	defer srv.Close()

	_, err := fetch.BMP(fetch.Config{URL: srv.URL, Slug: "s", Token: "t"})
	apiErr, ok := err.(*fetch.APIError)
	if !ok {
		t.Fatalf("expected *fetch.APIError, got %T", err)
	}
	if len(apiErr.Body) > 210 {
		t.Errorf("body not truncated, length %d", len(apiErr.Body))
	}
}

// TestBMP_Integration hits the real calendar API.
// Skipped when -short is set. Uses env vars BACK_CAL_SLUG and BACK_CAL_TOKEN,
// falling back to the test credentials from the project README.
func TestBMP_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	slug := os.Getenv("BACK_CAL_SLUG")
	if slug == "" {
		slug = "test-daytime_calendar"
	}
	token := os.Getenv("BACK_CAL_TOKEN")
	if token == "" {
		token = "prod-token-123"
	}

	data, err := fetch.BMP(fetch.Config{
		URL:   "https://gotta-go.notnot.uk/v1/frame",
		Slug:  slug,
		Token: token,
	})
	if err != nil {
		t.Fatalf("integration fetch failed: %v", err)
	}
	if len(data) < 4 {
		t.Fatalf("response too short to be a BMP: %d bytes", len(data))
	}
	// BMP files start with the magic bytes 'BM'
	if data[0] != 'B' || data[1] != 'M' {
		t.Errorf("response does not look like a BMP: first bytes %q", data[:4])
	}
	t.Logf("integration test ok: received %d byte BMP", len(data))
}
