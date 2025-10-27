package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type FeedCandidate struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Type  string `json:"type"` // "rss", "atom", or "unknown"
}

// DiscoverFeeds attempts to find RSS/Atom feeds for a given website URL
func DiscoverFeeds(websiteURL string) ([]FeedCandidate, error) {
	// Parse and normalize the URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure scheme is present
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
		websiteURL = parsedURL.String()
	}

	var candidates []FeedCandidate
	seen := make(map[string]bool)

	// Strategy 1: Try common feed URLs
	commonPaths := []string{
		"/feed",
		"/feed.xml",
		"/rss",
		"/rss.xml",
		"/atom.xml",
		"/feed.atom",
		"/feeds/posts/default", // Blogger
		"/?feed=rss2",          // WordPress
		"/?feed=atom",          // WordPress
		"/index.xml",           // Hugo
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	for _, path := range commonPaths {
		feedURL := baseURL + path
		if !seen[feedURL] && isValidFeed(feedURL) {
			candidates = append(candidates, FeedCandidate{
				URL:   feedURL,
				Title: "",
				Type:  detectFeedType(feedURL),
			})
			seen[feedURL] = true
		}
	}

	// Strategy 2: Parse HTML and look for <link> tags
	htmlFeeds, err := findFeedsInHTML(websiteURL)
	if err == nil {
		for _, feed := range htmlFeeds {
			// Resolve relative URLs
			feedURL := resolveURL(baseURL, feed.URL)
			if !seen[feedURL] {
				feed.URL = feedURL
				candidates = append(candidates, feed)
				seen[feedURL] = true
			}
		}
	}

	// Strategy 3: Look for links in the page content
	contentFeeds, err := findFeedsInContent(websiteURL, baseURL)
	if err == nil {
		for _, feed := range contentFeeds {
			feedURL := resolveURL(baseURL, feed.URL)
			if !seen[feedURL] {
				feed.URL = feedURL
				candidates = append(candidates, feed)
				seen[feedURL] = true
			}
		}
	}

	return candidates, nil
}

// findFeedsInHTML parses HTML and extracts feed links from <link> tags
func findFeedsInHTML(websiteURL string) ([]FeedCandidate, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(websiteURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var candidates []FeedCandidate
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" {
			var rel, href, title, feedType string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "rel":
					rel = strings.ToLower(attr.Val)
				case "href":
					href = attr.Val
				case "title":
					title = attr.Val
				case "type":
					feedType = strings.ToLower(attr.Val)
				}
			}

			// Check if it's a feed link
			if rel == "alternate" && href != "" {
				if strings.Contains(feedType, "rss") || strings.Contains(feedType, "atom") || strings.Contains(feedType, "xml") {
					fType := "unknown"
					if strings.Contains(feedType, "rss") {
						fType = "rss"
					} else if strings.Contains(feedType, "atom") {
						fType = "atom"
					}

					candidates = append(candidates, FeedCandidate{
						URL:   href,
						Title: title,
						Type:  fType,
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return candidates, nil
}

// findFeedsInContent looks for feed URLs in the page content (anchors, etc.)
func findFeedsInContent(websiteURL, baseURL string) ([]FeedCandidate, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(websiteURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	var candidates []FeedCandidate

	// Look for common feed URL patterns
	patterns := []string{
		`href=["']([^"']*(?:feed|rss|atom)[^"']*)["']`,
		`href=["']([^"']*\.xml)["']`,
	}

	seen := make(map[string]bool)
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				feedURL := match[1]
				if !seen[feedURL] && !strings.HasPrefix(feedURL, "javascript:") && !strings.HasPrefix(feedURL, "#") {
					candidates = append(candidates, FeedCandidate{
						URL:  feedURL,
						Type: detectFeedType(feedURL),
					})
					seen[feedURL] = true
				}
			}
		}
	}

	return candidates, nil
}

// isValidFeed checks if a URL actually returns a valid feed
func isValidFeed(feedURL string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(feedURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	// Read first 1KB to check if it looks like a feed
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	content := string(buf[:n])

	// Check for feed indicators
	content = strings.ToLower(content)
	return strings.Contains(content, "<rss") ||
		strings.Contains(content, "<feed") ||
		strings.Contains(content, "<atom") ||
		strings.Contains(content, "<?xml")
}

// detectFeedType tries to determine if a URL is RSS or Atom based on the URL
func detectFeedType(feedURL string) string {
	lower := strings.ToLower(feedURL)
	if strings.Contains(lower, "atom") {
		return "atom"
	}
	if strings.Contains(lower, "rss") {
		return "rss"
	}
	return "unknown"
}

// resolveURL resolves relative URLs to absolute URLs
func resolveURL(baseURL, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}

	ref, err := url.Parse(href)
	if err != nil {
		return href
	}

	return base.ResolveReference(ref).String()
}
