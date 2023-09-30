package config

import (
	"fmt"

	"github.com/xdrm-io/aicra/validator"
)

type captureValidation struct {
	ValidatorName   string
	ValidatorParams []string
}

// captures returns the captures' validators for an endpoint indexed by their
// index in the URI
func captureSpec(endpoint *Endpoint) map[int]captureValidation {
	captures := make(map[int]captureValidation, len(endpoint.Captures))
	for _, capture := range endpoint.Captures {
		p, ok := endpoint.Input[`{`+capture.Name+`}`]
		if !ok {
			panic(fmt.Errorf("(%s %q) capture %d %q not found in inputs %v", endpoint.Method, endpoint.Pattern, capture.Index, capture.Name, endpoint.Input))
		}
		captures[capture.Index] = captureValidation{
			ValidatorName:   p.ValidatorName,
			ValidatorParams: p.ValidatorParams,
		}
	}
	return captures
}

// check if uri of services A and B collide
func checkURICollision(aFragments, bFragments []string, aCaptures, bCaptures map[int]captureValidation, avail Validators) error {
	var err error

	// for each part
	for i, aSeg := range aFragments {
		var (
			bSeg = bFragments[i]

			aCapture, aIsCapture = aCaptures[i]
			bCapture, bIsCapture = bCaptures[i]
		)

		// both captures -> as we cannot check, consider a collision
		if aIsCapture && bIsCapture {
			err = fmt.Errorf("%w (path %s and %s)", ErrPatternCollision, aSeg, bSeg)
			continue
		}

		// no capture -> check strict equality
		if !aIsCapture && !bIsCapture && aSeg == bSeg {
			err = fmt.Errorf("%w (same path %q)", ErrPatternCollision, aSeg)
			continue
		}

		// A captures B -> fail if B is a valid A value
		if aIsCapture && validates(avail[aCapture.ValidatorName], aCapture.ValidatorParams, bSeg) {
			err = fmt.Errorf("%w (%s captures %q)", ErrPatternCollision, aSeg, bSeg)
			continue
		}
		// B captures A -> fail is A is a valid B value
		if bIsCapture && validates(avail[bCapture.ValidatorName], bCapture.ValidatorParams, aSeg) {
			err = fmt.Errorf("%w (%s captures %q)", ErrPatternCollision, bSeg, aSeg)
			continue
		}
		// no match for at least one segment -> no collision
		return nil
	}
	return err
}

// validates returns whether a parameter validates a given value
func validates(v validator.Validator[any], params []string, value string) bool {
	extractFn := v.Validate(params)
	if extractFn == nil {
		// must not happen
		return false
	}
	_, valid := extractFn(value)
	return valid
}
