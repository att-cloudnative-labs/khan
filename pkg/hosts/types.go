package hosts

// Host contains metadata for each target host
type Host struct {
	Type      string `json:"type"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	App       string `json:"app"`
	NodeIP    string `json:"nodeIP"`
}

// HostCache map of IP to
type HostCache map[string]Host
