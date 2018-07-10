package controller

/* (1) Configuration
---------------------------------------------------------*/

type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Rename      string `json:"name,omitempty"`
	Optional    bool
	Default     *interface{} `json:"default"`
}
type Method struct {
	Description string                `json:"info"`
	Permission  [][]string            `json:"scope"`
	Parameters  map[string]*Parameter `json:"in"`
	Download    *bool                 `json:"download"`
}

type Controller struct {
	GET    *Method `json:"GET"`
	POST   *Method `json:"POST"`
	PUT    *Method `json:"PUT"`
	DELETE *Method `json:"DELETE"`

	Children map[string]*Controller `json:"/"`
}
