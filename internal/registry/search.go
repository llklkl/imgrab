package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type SearchResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Stars       int    `json:"star_count"`
	IsOfficial  bool   `json:"is_official"`
	IsAutomated bool   `json:"is_automated"`
}

type SearchResponse struct {
	Results    []SearchResult `json:"results"`
	Count      int            `json:"count"`
	Next       string         `json:"next"`
	Previous   string         `json:"previous"`
}

type TagResult struct {
	Name    string    `json:"name"`
	LastUpdated time.Time `json:"last_updated"`
}

type TagsResponse struct {
	Results []TagResult `json:"results"`
	Count   int         `json:"count"`
	Next    string      `json:"next"`
}

const (
	dockerHubSearchAPI = "https://hub.docker.com/v2/search/repositories"
	dockerHubTagsAPI   = "https://hub.docker.com/v2/repositories"
)

func SearchImages(query string, page, pageSize int) (*SearchResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	u, err := url.Parse(dockerHubSearchAPI)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("page_size", fmt.Sprintf("%d", pageSize))
	u.RawQuery = q.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func ListTags(repository string, page, pageSize int) (*TagsResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if !containsSlash(repository) {
		repository = "library/" + repository
	}

	u, err := url.Parse(fmt.Sprintf("%s/%s/tags", dockerHubTagsAPI, repository))
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("page_size", fmt.Sprintf("%d", pageSize))
	u.RawQuery = q.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("tags request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get tags failed: %s", resp.Status)
	}

	var result TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}
