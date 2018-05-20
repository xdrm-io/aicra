package config

/* (1) Configuration
---------------------------------------------------------*/

type MethodParameter struct {
	Description string       `json:"des"`
	Type        string       `json:"typ"`
	Rename      *string      `json:"ren"`
	Optional    *bool        `json:"opt"`
	Default     *interface{} `json:"def"`
}
type Method struct {
	Description string                      `json:"des"`
	Permission  [][]string                  `json:"per"`
	Parameters  *map[string]MethodParameter `json:"par"`
	Options     *map[string]interface{}     `json:"opt"`
}

type Controller struct {
	GET    *Method `json:"GET"`
	POST   *Method `json:"POST"`
	PUT    *Method `json:"PUT"`
	DELETE *Method `json:"DELETE"`

	Children map[string]Controller `json:"/"`
}
