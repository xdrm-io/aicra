package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParamKind defines supported parameter kinds
type ParamKind uint8

// supported kinds
const (
	KindURI ParamKind = iota
	KindQuery
	KindForm
)

// Parameter represents a parameter definition (from api.json)
type Parameter struct {
	Type     string `json:"type"`
	Rename   string `json:"name,omitempty"`
	Optional bool   `json:"-"`

	ValidatorName   string   `json:"-"`
	ValidatorParams []string `json:"-"`

	// filled by the endpoint containing the parameter
	Kind        ParamKind `json:"-"`
	ExtractName string    `json:"-"`
}

var (
	typeRe      = regexp.MustCompile(`^([^\(]+)(?:\(([^\),]+(?:, ?[^\),]+)*)\))?$`)
	paramNameRe = regexp.MustCompile(`^[A-Z][A-Za-z0-9_-]*$`)
)

// UnmarshalJSON with custom validation and parsing
func (p *Parameter) UnmarshalJSON(b []byte) error {
	type receiver Parameter
	var r receiver
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}
	p.Type = r.Type
	p.Rename = r.Rename

	if p.Rename != "" && !paramNameRe.MatchString(p.Rename) {
		return fmt.Errorf("param '%s': %w", p.Rename, ErrParamRenameInvalid)
	}

	if p.Type == "" {
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
	matches := typeRe.FindStringSubmatch(p.Type)
	if len(matches) != 3 {
		return ErrParamTypeInvalid
	}

	p.ValidatorParams = make([]string, 0)
	p.ValidatorName = matches[1]
	if matches[2] != "" {
		p.ValidatorParams = append(p.ValidatorParams, strings.Split(strings.ReplaceAll(matches[2], ", ", ","), ",")...)
	}
	return nil
}

// RuntimeCheck fails when the config is invalid with the code-generated
// validators
func (p Parameter) RuntimeCheck(avail Validators) error {
	v, ok := avail[p.ValidatorName]
	if !ok || v == nil {
		return ErrParamTypeUnknown
	}
	if v.Validate(p.ValidatorParams) == nil {
		return ErrParamTypeParamsInvalid
	}
	return nil
}
