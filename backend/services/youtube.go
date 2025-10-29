package services

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// ResolveYouTubeChannelID takes a YouTube URL or handle and returns the actual channel ID
func ResolveYouTubeChannelID(input string) (string, error) {
	input = strings.TrimSpace(input)

	// If it's already a channel ID (starts with UC and is 24 chars), return it
	if regexp.MustCompile(`^UC[\w-]{22}$`).MatchString(input) {
		return input, nil
	}

	// Extract channel identifier from various URL formats
	patterns := []struct {
		regex   *regexp.Regexp
		urlType string
	}{
		{regexp.MustCompile(`youtube\.com/channel/(UC[\w-]{22})`), "channel_id"},
		{regexp.MustCompile(`youtube\.com/@([\w-]+)`), "handle"},
		{regexp.MustCompile(`youtube\.com/c/([\w-]+)`), "custom"},
		{regexp.MustCompile(`youtube\.com/user/([\w-]+)`), "user"},
	}

	var identifier, urlType string
	for _, p := range patterns {
		if match := p.regex.FindStringSubmatch(input); match != nil {
			identifier = match[1]
			urlType = p.urlType
			break
		}
	}

	// If no URL pattern matched, assume it's a handle or channel name
	if identifier == "" {
		identifier = strings.TrimPrefix(input, "@")
		urlType = "handle"
	}

	// If we already have a channel ID from the URL, return it
	if urlType == "channel_id" {
		return identifier, nil
	}

	// For handles, custom URLs, or usernames, we need to fetch the page to get the channel ID
	return fetchChannelIDFromPage(identifier, urlType)
}

// fetchChannelIDFromPage fetches the YouTube page and extracts the channel ID
func fetchChannelIDFromPage(identifier, urlType string) (string, error) {
	var pageURL string
	switch urlType {
	case "handle":
		pageURL = fmt.Sprintf("https://www.youtube.com/@%s", identifier)
	case "custom":
		pageURL = fmt.Sprintf("https://www.youtube.com/c/%s", identifier)
	case "user":
		pageURL = fmt.Sprintf("https://www.youtube.com/user/%s", identifier)
	default:
		pageURL = fmt.Sprintf("https://www.youtube.com/@%s", identifier)
	}

	// Fetch the page
	resp, err := http.Get(pageURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch YouTube page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("YouTube page returned status %d", resp.StatusCode)
	}

	// Read the page content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read page content: %w", err)
	}

	pageContent := string(body)

	// Try multiple patterns to extract channel ID
	patterns := []*regexp.Regexp{
		// Pattern 1: "channelId":"UCxxxx"
		regexp.MustCompile(`"channelId":"(UC[\w-]{22})"`),
		// Pattern 2: "externalId":"UCxxxx"
		regexp.MustCompile(`"externalId":"(UC[\w-]{22})"`),
		// Pattern 3: channel_id=UCxxxx in URLs
		regexp.MustCompile(`channel_id=(UC[\w-]{22})`),
		// Pattern 4: /channel/UCxxxx in links
		regexp.MustCompile(`/channel/(UC[\w-]{22})`),
	}

	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(pageContent); match != nil {
			return match[1], nil
		}
	}

	return "", fmt.Errorf("could not find channel ID for %s", identifier)
}
