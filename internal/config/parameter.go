package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Parameter represents a parameter definition (from api.json)
type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Rename      string `json:"name,omitempty"`
	Optional    bool   `json:"-"`

	ValidatorName   string   `json:"-"`
	ValidatorParams []string `json:"-"`
}

var typenameRe = regexp.MustCompile(`^([^\(]+)(\([^\),]+(?:, ?[^\),]+)*\))?$`)

// UnmarshalJSON with custom validation and parsing
func (p *Parameter) UnmarshalJSON(b []byte) error {
	type receiver Parameter
	var r receiver
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}
	p.Description = r.Description
	p.Type = r.Type
	p.Rename = r.Rename

	if len(p.Type) < 1 || p.Type == "?" {
		return fmt.Errorf("param '%s': %w", p.Rename, ErrParamTypeMissing)
	}
	if p.Type[0] == '?' {
		p.Optional = true
		p.Type = strings.TrimPrefix(p.Type, "?")
	}
	if err := p.parseType(); err != nil {
		return fmt.Errorf("param '%s': %w", p.Rename, err)
	}
	return nil
}

// parseType extracts the validator name and its parameters from the config
// format.
//
// The format is :
// * 'validatorName' when there is no param
// * 'validatorName(abc)' when there is 1 param "abc"
// * 'validatorName(abc,123)' when there is 2 params "abc" and "123"
func (p *Parameter) parseType() error {
	matches := typenameRe.FindStringSubmatch(p.Type)
	if len(matches) == 2 {
		p.ValidatorName = matches[1]
		return nil
	}
	if len(matches) == 3 {
		p.ValidatorName = matches[1]
		p.ValidatorParams = strings.Split(strings.ReplaceAll(matches[2], ", ", ","), ",")
		return nil
	}
	return ErrParamTypeInvalid
}

// Validate fails when the validator is not found or invalid from the provided
// map
func (p *Parameter) Validate(avail Validators) error {
	validator, ok := avail[p.ValidatorName]
	if !ok {
		return ErrParamTypeUnknown
	}
	if validator(p.ValidatorParams) == nil {
		return ErrParamTypeParamsInvalid
	}
	return nil
}
