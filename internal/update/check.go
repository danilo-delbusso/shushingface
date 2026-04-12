package update

import (
	"context"
	"encoding/json"
	"fmt"
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
		"https://codeberg.org/api/v1/repos/dbus/shushingface/releases?limit=1", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release API returned %d", resp.StatusCode)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	if len(releases) == 0 || releases[0].TagName == "" {
		return nil, nil
	}

	if isNewer(currentVersion, releases[0].TagName) {
		return &releases[0], nil
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
