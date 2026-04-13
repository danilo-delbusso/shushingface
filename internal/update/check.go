package update

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Release represents a release from the Codeberg/Forgejo API.
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// Check queries the Codeberg API for the latest release and returns it
// if it's newer than currentVersion. Returns nil if already up to date.
func Check(ctx context.Context, currentVersion string) (*Release, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://codeberg.org/api/v1/repos/dbus/shushingface/releases/latest", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close release response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, nil // no stable release yet, that's fine
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}

	if rel.TagName == "" {
		return nil, nil
	}

	if isNewer(currentVersion, rel.TagName) {
		return &rel, nil
	}
	return nil, nil
}

// isNewer returns true if remote is a newer version than current.
// Compares by stripping the "v" prefix and doing a simple string compare.
// For proper semver comparison, use a library — this handles the common
// case of v0.1.0 < v0.2.0 < v0.10.0 correctly via split-and-compare.
func isNewer(current, remote string) bool {
	current = strings.TrimPrefix(current, "v")
	remote = strings.TrimPrefix(remote, "v")

	if current == "" || current == "dev" {
		return false // dev builds never prompt for updates
	}

	cParts := strings.Split(current, ".")
	rParts := strings.Split(remote, ".")

	for i := 0; i < len(cParts) && i < len(rParts); i++ {
		c := cParts[i]
		r := rParts[i]
		// Pad to same length for numeric comparison
		for len(c) < len(r) {
			c = "0" + c
		}
		for len(r) < len(c) {
			r = "0" + r
		}
		if r > c {
			return true
		}
		if r < c {
			return false
		}
	}
	return len(rParts) > len(cParts)
}
