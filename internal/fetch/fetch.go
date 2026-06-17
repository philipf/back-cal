package fetch

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

// APIError is returned when the server responds with a non-200 status.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API returned %d: %s", e.StatusCode, e.Body)
}

// Config holds the parameters needed to fetch the calendar image.
type Config struct {
	URL   string
	Slug  string
	Token string
}

// BMP fetches the calendar image as raw BMP bytes.
func BMP(cfg Config) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "image/bmp")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("x-radiator-slug", cfg.Slug)
	req.Header.Set("x-radiator-token", cfg.Token)

	// Setting Accept-Encoding manually disables the Transport's transparent
	// gzip decompression, so we decompress explicitly if needed.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	reader := io.Reader(resp.Body)
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("creating gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Body: snippet}
	}

	return body, nil
}
