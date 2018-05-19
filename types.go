package gfw

/* (1) Configuration
---------------------------------------------------------*/

type methodParameter struct {
	Description string       `json:"des"`
	Type        string       `json:"typ"`
	Rename      *string      `json:"ren"`
	Optional    *bool        `json:"opt"`
	Default     *interface{} `json:"def"`
}
type method struct {
	Description string                      `json:"des"`
	Permission  [][]string                  `json:"per"`
	Parameters  *map[string]methodParameter `json:"par"`
	Options     *map[string]interface{}     `json:"opt"`
}

type controller struct {
	GET    *method `json:"GET"`
	POST   *method `json:"POST"`
	PUT    *method `json:"PUT"`
	DELETE *method `json:"DELETE"`

	Children map[string]controller `json:"/"`
}
