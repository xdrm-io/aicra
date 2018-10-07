package api

/* (1) Configuration
---------------------------------------------------------*/
// Parameter represents a parameter definition (from api.json)
type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Rename      string `json:"name,omitempty"`
	Optional    bool
	Default     *interface{} `json:"default"`
}

// Method represents a method definition (from api.json)
type Method struct {
	Description string                `json:"info"`
	Permission  [][]string            `json:"scope"`
	Parameters  map[string]*Parameter `json:"in"`
	Download    *bool                 `json:"download"`
}

// Controller represents a controller definition (from api.json)
type Controller struct {
	GET    *Method `json:"GET"`
	POST   *Method `json:"POST"`
	PUT    *Method `json:"PUT"`
	DELETE *Method `json:"DELETE"`

	Children map[string]*Controller `json:"/"`
}
