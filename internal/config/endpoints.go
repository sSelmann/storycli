package config

// ItrocketEndpoints holds the URLs for Itrocket provider based on pruning mode.
type ItrocketEndpoints struct {
	Pruned  []string // List of pruned server URLs
	Archive []string // List of archive server URLs
}

// Endpoints struct holds all API endpoints for different providers.
type Endpoints struct {
	Itrocket ItrocketEndpoints
	Krews    string
}

// DefaultEndpoints returns the default API endpoints.
func DefaultEndpoints() Endpoints {
	return Endpoints{
		Itrocket: ItrocketEndpoints{
			Pruned: []string{
				"https://server-1.itrocket.net/testnet/story/.current_state.json",
				"https://server-3.itrocket.net/testnet/story/.current_state.json",
			},
			Archive: []string{
				"https://server-5.itrocket.net/testnet/story/.current_state.json",
				"https://server-8.itrocket.net/testnet/story/.current_state.json",
			},
		},
		Krews: "https://snapshots-api.krews.xyz/api/snapshots/story",
	}
}
