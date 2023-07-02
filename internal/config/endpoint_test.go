package config_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/config"
)

func TestEndpointUnmarshal(t *testing.T) {
	t.Parallel()

	type TC struct {
		name     string
		conf     string
		err      error
		endpoint config.Endpoint
	}

	tt := map[string][]TC{
		"name": {
			{
				name: "missing",
				conf: `{}`,
				err:  config.ErrNameMissing,
			},
			{
				name: "unexported",
				conf: `{ "name": "a" }`,
				err:  config.ErrNameUnexported,
			},
			{
				name: "invalid",
				conf: `{ "name": "A*" }`,
				err:  config.ErrNameInvalid,
			},
			{
				name: "ok",
				conf: `{ "name": "Va_l1d"  }`,
				err:  config.ErrMethodUnknown,
			},
		},

		"method": {
			{
				name: "missing",
				conf: `{ "name": "Va_l1d" }`,
				err:  config.ErrMethodUnknown,
			},
			{
				name: "unknown",
				conf: `{ "name": "Va_l1d", "method": "unknown" }`,
				err:  config.ErrMethodUnknown,
			},
			{
				name: "GET ok",
				conf: `{ "name": "Va_l1d", "method": "GET" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "POST ok",
				conf: `{ "name": "Va_l1d", "method": "POST" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "PUT ok",
				conf: `{ "name": "Va_l1d", "method": "PUT" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "PATCH ok",
				conf: `{ "name": "Va_l1d", "method": "PATCH" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "DELETE ok",
				conf: `{ "name": "Va_l1d", "method": "DELETE" }`,
				err:  config.ErrPatternInvalid,
			},
		},

		"path": {
			{
				name: "missing",
				conf: `{ "name": "OK", "method": "GET" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "empty",
				conf: `{ "name": "OK", "method": "GET", "path": "" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "root",
				conf: `{ "name": "OK", "method": "GET", "path": "/" }`,
				err:  config.ErrDescMissing,
			},
			{
				name: "empty fragment",
				conf: `{ "name": "OK", "method": "GET", "path": "/a//b" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "no starting slash",
				conf: `{ "name": "OK", "method": "GET", "path": "no-starting-slash" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "ending slash",
				conf: `{ "name": "OK", "method": "GET", "path": "ending-slash/" }`,
				err:  config.ErrPatternInvalid,
			},
			{
				name: "simple ok",
				conf: `{ "name": "OK", "method": "GET", "path": "/valid-name" }`,
				err:  config.ErrDescMissing,
			},
			{
				name: "nested ok",
				conf: `{ "name": "OK", "method": "GET", "path": "/valid/nested/name" }`,
				err:  config.ErrDescMissing,
			},
			{
				name: "capture not after slash",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/s{braces}" }`,
				err:  config.ErrPatternInvalidBraceCapture,
			},
			{
				name: "capture not before slash",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/{braces}a" }`,
				err:  config.ErrPatternInvalidBraceCapture,
			},
			{
				name: "ending capture ok",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/{braces}" }`,
				err:  config.ErrDescMissing,
			},
			{
				name: "ending capture case ok",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/{BrAcEs}" }`,
				err:  config.ErrDescMissing,
			},
			{
				name: "invalid middle capture before slash",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/s{braces}/abc" }`,
				err:  config.ErrPatternInvalidBraceCapture,
			},
			{
				name: "invalid middle capture after slash",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/{braces}s/abc" }`,
				err:  config.ErrPatternInvalidBraceCapture,
			},
			{
				name: "middle capture ok",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/{braces}/abc" }`,
				err:  config.ErrDescMissing,
			},
			{
				name: "middle capture with {",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/{b{races}s/abc" }`,
				err:  config.ErrPatternInvalidBraceCapture,
			},
			{
				name: "middle capture invalid } after slash",
				conf: `{ "name": "OK", "method": "GET", "path": "/invalid/{braces}/}abc" }`,
				err:  config.ErrPatternInvalidBraceCapture,
			},
		},

		"description": {
			{
				name: "missing",
				conf: `{ "name": "OK", "method": "GET", "path": "/" }`,
				err:  config.ErrDescMissing,
			},
			{
				name: "empty",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "" }`,
				err:  config.ErrDescMissing,
			},
			{
				name: "ok",
				conf: `{ "name": "OK", "method": "GET", "path": "/{a}", "info": "description" }`,
				err:  config.ErrBraceCaptureUndefined,
			},
		},

		"in": {
			{
				name: "empty name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"": { "type": "string" }
					}
				}`,
				err: config.ErrParamNameIllegal,
			},
		},

		"in:uri": {
			{
				name: "empty",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"{a}": {}
					}
				}`,
				err: config.ErrParamTypeMissing,
			},
			{
				name: "missing name",
				conf: `{ "name": "OK", "method": "GET", "path": "/{uri}", "info": "description",
					"in": {
						"{uri}": { "type": "string" }
					}
				}`,
				err: config.ErrRenameMandatory,
			},
			{
				name: "empty name",
				conf: `{ "name": "OK", "method": "GET", "path": "/{uri}", "info": "description",
					"in": {
						"{}": { "type": "string" }
					}
				}`,
				err: config.ErrParamNameIllegal,
			},
			{
				name: "optional",
				conf: `{ "name": "OK", "method": "GET", "path": "/{uri}", "info": "description",
					"in": {
						"{uri}": { "type": "?string", "name": "URI" }
					}
				}`,
				err: config.ErrParamOptionalIllegalURI,
			},
			{
				name: "renamed",
				conf: `{ "name": "OK", "method": "GET", "path": "/{uri}", "info": "description",
					"in": {
						"{uri}": { "type": "string", "name": "URI" }
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/{uri}",
					Description: "description",
					Input: map[string]*config.Parameter{
						"{uri}": {
							Kind:            config.KindURI,
							Type:            "string",
							Rename:          "URI",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "0",
						},
					},
					Output: make(map[string]*config.Parameter),
					Captures: []*config.BraceCapture{
						{Name: "uri", Index: 0, Defined: true},
					},
					Fragments: []string{"{uri}"},
				},
			},
		},

		"in:query": {
			{
				name: "missing name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"?query": { "type": "string" }
					}
				}`,
				err: config.ErrRenameMandatory,
			},
			{
				name: "renamed",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"?query": { "type": "string", "name": "Query" }
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/",
					Description: "description",
					Input: map[string]*config.Parameter{
						"?query": {
							Kind:            config.KindQuery,
							Type:            "string",
							Rename:          "Query",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "query",
						},
					},
					Output:    map[string]*config.Parameter{},
					Fragments: []string{},
				},
			},
			{
				name: "renamed optional",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"?query": { "type": "?string", "name": "Query" }
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/",
					Description: "description",
					Input: map[string]*config.Parameter{
						"?query": {
							Kind:            config.KindQuery,
							Type:            "string",
							Rename:          "Query",
							Optional:        true,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "query",
						},
					},
					Output:    map[string]*config.Parameter{},
					Fragments: []string{},
				},
			},
		},

		"in:form": {
			{
				name: "missing type",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"form": {  }
					}
				}`,
				err: config.ErrParamTypeMissing,
			},
			{
				name: "unexported name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"form": { "type": "string" }
					}
				}`,
				err: config.ErrRenameUnexported,
			},
			{
				name: "unexported rename unexported",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"form": { "type": "string", "name": "unexported" }
					}
				}`,
				err: config.ErrNameUnexported,
			},
			{
				name: "unexported renamed",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"form": { "type": "string", "name": "Exported" }
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/",
					Description: "description",
					Input: map[string]*config.Parameter{
						"form": {
							Kind:            config.KindForm,
							Type:            "string",
							Rename:          "Exported",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "form",
						},
					},
					Output:    map[string]*config.Parameter{},
					Fragments: []string{},
				},
			},
			{
				name: "exported name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"Exported": { "type": "string"}
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/",
					Description: "description",
					Input: map[string]*config.Parameter{
						"Exported": {
							Kind:            config.KindForm,
							Type:            "string",
							Rename:          "Exported",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "Exported",
						},
					},
					Output:    map[string]*config.Parameter{},
					Fragments: []string{},
				},
			},
			{
				name: "optional",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"form": { "type": "?string", "name": "Exported" }
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/",
					Description: "description",
					Input: map[string]*config.Parameter{
						"form": {
							Kind:            config.KindForm,
							Type:            "string",
							Rename:          "Exported",
							Optional:        true,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "form",
						},
					},
					Output:    map[string]*config.Parameter{},
					Fragments: []string{},
				},
			},
		},

		"in:name conflicts": {
			{
				name: "name = rename",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"Name1": { "type": "string" },
						"Name2": { "type": "string", "name": "Name1" }
					}
				}`,
				err: config.ErrParamNameConflict,
			},
			{
				name: "rename = name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"Name1": { "type": "string", "name": "Name2" },
						"Name2": { "type": "string" }
					}
				}`,
				err: config.ErrParamNameConflict,
			},
			{
				name: "rename = rename",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": {
						"Name1": { "type": "string", "name": "Name3" },
						"Name2": { "type": "string", "name": "Name3" }
					}
				}`,
				err: config.ErrParamNameConflict,
			},
		},

		"capture": {
			{
				name: "pattern not in input",
				conf: `{ "name": "OK", "method": "GET", "path": "/{a}", "info": "description" }`,
				err:  config.ErrBraceCaptureUndefined,
			},
			{
				name: "input uri not in pattern",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"in": { "{a}": { "type": "string", "name": "A" } }
				}`,
				err: config.ErrBraceCaptureUnspecified,
			},
			{
				name: "complex",
				conf: `{ "name": "OK", "method": "GET", "path": "/a/{b}/{c}/d/e/{f}", "info": "description",
					"in": {
						"{b}": { "type": "string", "name": "URI1" },
						"{c}": { "type": "string", "name": "URI2" },
						"{f}": { "type": "string", "name": "URI3" }
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/a/{b}/{c}/d/e/{f}",
					Description: "description",
					Input: map[string]*config.Parameter{
						"{b}": {
							Kind:            config.KindURI,
							Type:            "string",
							Rename:          "URI1",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "1",
						},
						"{c}": {
							Kind:            config.KindURI,
							Type:            "string",
							Rename:          "URI2",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "2",
						},
						"{f}": {
							Kind:            config.KindURI,
							Type:            "string",
							Rename:          "URI3",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "5",
						},
					},
					Output: make(map[string]*config.Parameter),
					Captures: []*config.BraceCapture{
						{Name: "b", Index: 1, Defined: true},
						{Name: "c", Index: 2, Defined: true},
						{Name: "f", Index: 5, Defined: true},
					},
					Fragments: []string{
						"a",
						"{b}",
						"{c}",
						"d",
						"e",
						"{f}",
					},
				},
			},
		},

		"out": {
			{
				name: "empty name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": { "": { "type": "string" } }
				}`,
				err: config.ErrParamNameIllegal,
			},
			{
				name: "missing type",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"form": {  }
					}
				}`,
				err: config.ErrParamTypeMissing,
			},
			{
				name: "unexported name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"form": { "type": "string" }
					}
				}`,
				err: config.ErrRenameUnexported,
			},
			{
				name: "unexported rename unexported",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"form": { "type": "string", "name": "unexported" }
					}
				}`,
				err: config.ErrNameUnexported,
			},
			{
				name: "unexported renamed",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"form": { "type": "string", "name": "Exported" }
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/",
					Description: "description",
					Output: map[string]*config.Parameter{
						"form": {
							Kind:            config.KindForm,
							Type:            "string",
							Rename:          "Exported",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "form",
						},
					},
					Input:     map[string]*config.Parameter{},
					Fragments: []string{},
				},
			},
			{
				name: "exported name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"Exported": { "type": "string"}
					}
				}`,
				err: nil,
				endpoint: config.Endpoint{
					Name:        "OK",
					Method:      "GET",
					Pattern:     "/",
					Description: "description",
					Output: map[string]*config.Parameter{
						"Exported": {
							Kind:            config.KindForm,
							Type:            "string",
							Rename:          "Exported",
							Optional:        false,
							ValidatorName:   "string",
							ValidatorParams: []string{},
							ExtractName:     "Exported",
						},
					},
					Input:     map[string]*config.Parameter{},
					Fragments: []string{},
				},
			},
			{
				name: "optional",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"form": { "type": "?string", "name": "Exported" }
					}
				}`,
				err: config.ErrOutputOptional,
			},
			{
				name: "uri",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"{uri}": { "type": "string" }
					}
				}`,
				err: config.ErrOutputURIForbidden,
			},
			{
				name: "query",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"?query": { "type": "string" }
					}
				}`,
				err: config.ErrOutputQueryForbidden,
			},
		},

		"out:name conflicts": {
			{
				name: "name = rename",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"Name1": { "type": "string" },
						"Name2": { "type": "string", "name": "Name1" }
					}
				}`,
				err: config.ErrParamNameConflict,
			},
			{
				name: "rename = name",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"Name1": { "type": "string", "name": "Name2" },
						"Name2": { "type": "string" }
					}
				}`,
				err: config.ErrParamNameConflict,
			},
			{
				name: "rename = rename",
				conf: `{ "name": "OK", "method": "GET", "path": "/", "info": "description",
					"out": {
						"Name1": { "type": "string", "name": "Name3" },
						"Name2": { "type": "string", "name": "Name3" }
					}
				}`,
				err: config.ErrParamNameConflict,
			},
		},
	}

	for group, tg := range tt {
		for _, tc := range tg {
			t.Run(group+"/"+tc.name, func(t *testing.T) {
				var endpoint config.Endpoint
				err := json.Unmarshal([]byte(tc.conf), &endpoint)
				if tc.err != nil {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.err)
					return
				}
				require.NoError(t, err)
				require.EqualValues(t, tc.endpoint, endpoint)
			})
		}
	}
}
