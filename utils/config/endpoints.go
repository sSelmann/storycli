package config

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ItrocketEndpoints holds the URLs for Itrocket provider based on pruning mode.
type ItrocketEndpoints struct {
	Pruned  []string // List of pruned server URLs
	Archive []string // List of archive server URLs
}

// Endpoints struct holds all API endpoints for different providers.
type Endpoints struct {
	Itrocket ItrocketEndpoints
	Krews    string
	Jnode    string
}

type itrocketAPIResponse struct {
	Archive map[string]string `json:"archive"`
	Pruned  map[string]string `json:"pruned"`
}

// fetchItrocketEndpointsFromAPI fetches the dynamic endpoints from the external API
// and converts them to ItrocketEndpoints (pruned + archive URLs).
func fetchItrocketEndpointsFromAPI() (ItrocketEndpoints, error) {
	var result ItrocketEndpoints

	apiURL := "https://snapshot-external-providers-api.krews.xyz/snapshots/itrocket"
	resp, err := http.Get(apiURL)
	if err != nil {
		return result, fmt.Errorf("failed to GET Itrocket endpoint API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("got non-OK status code %d from Itrocket endpoint API", resp.StatusCode)
	}

	var apiResp itrocketAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return result, fmt.Errorf("failed to decode Itrocket endpoint API response: %v", err)
	}

	// Convert the map data (e.g. "server-3.itrocket.net") to full URLs.
	// For pruned:
	prunedList := []string{}
	for _, endpoint := range apiResp.Pruned {
		// e.g. "server-3.itrocket.net" => "https://server-3.itrocket.net/testnet/story/.current_state.json"
		fullURL := fmt.Sprintf("https://%s/testnet/story/.current_state.json", endpoint)
		prunedList = append(prunedList, fullURL)
	}

	// For archive:
	archiveList := []string{}
	for _, endpoint := range apiResp.Archive {
		fullURL := fmt.Sprintf("https://%s/testnet/story/.current_state.json", endpoint)
		archiveList = append(archiveList, fullURL)
	}

	result = ItrocketEndpoints{
		Pruned:  prunedList,
		Archive: archiveList,
	}
	return result, nil
}

// DefaultEndpoints returns the default API endpoints.
// It dynamically fetches Itrocket endpoints via fetchItrocketEndpointsFromAPI,
// and if there's a failure, it panics (no fallback).
func DefaultEndpoints() Endpoints {
	dynamicItrocket, err := fetchItrocketEndpointsFromAPI()
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch Itrocket endpoints from API: %v", err))
	}

	return Endpoints{
		Itrocket: dynamicItrocket,
		Krews:    "https://snapshots-api.krews.xyz/api/snapshots/story",
		Jnode:    "https://snapshot-external-providers-api.krews.xyz/snapshots/jnode",
	}
}
