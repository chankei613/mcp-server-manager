package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type UpdateInfo struct {
	Version    string `json:"version"`
	ReleaseURL string `json:"release_url"`
}

const githubRepo = "chankei613/mcp-server-manager"

func checkForUpdate() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)

	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api status: %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	if isNewerVersion(latest, AppVersion) {
		return &UpdateInfo{
			Version:    latest,
			ReleaseURL: release.HTMLURL,
		}, nil
	}
	return nil, nil
}

// isNewerVersion は latest が current より新しいか判定する（semver 大小比較）
func isNewerVersion(latest, current string) bool {
	lp := parseSemver(latest)
	cp := parseSemver(current)
	for i := range lp {
		if lp[i] > cp[i] {
			return true
		}
		if lp[i] < cp[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) [3]int {
	var major, minor, patch int
	fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)
	return [3]int{major, minor, patch}
}
