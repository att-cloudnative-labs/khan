package mappings

// Target contains metadata for each target host
type Target struct {
	Type      string `json:"type"`
	Namespace string `json:"namespace"`
	AppName   string `json:"appName"`
	PodName   string `json:"podName"`
	NodeIP    string `json:"nodeIP"`
}
