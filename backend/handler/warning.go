package handler

// Warning represents a structured linting result with associated node IDs.
type Warning struct {
	Rule     string   `json:"rule"`
	Message  string   `json:"message"`
	Solution string   `json:"solution"`
	NodeIDs  []string `json:"nodeIds"`
}

// AnalyzeResponse is returned after parsing and validating the topology.
type AnalyzeResponse struct {
	Success   bool      `json:"success"`
	NodeCount int       `json:"nodeCount"`
	EdgeCount int       `json:"edgeCount"`
	Warnings  []Warning `json:"warnings,omitempty"`
}
