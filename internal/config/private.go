package config

import (
	"fmt"
	"strings"
)

// format checks for format errors and missing required fields
// it also sets default values to optional fields
func (c *Controller) format(controllerName string) error {

	/* (1) Check each method
	---------------------------------------------------------*/
	methods := []struct {
		Name string
		Ptr  *Method
	}{
		{"GET", c.GET},
		{"POST", c.POST},
		{"PUT", c.PUT},
		{"DELETE", c.DELETE},
	}

	for _, method := range methods {

		/* (1) ignore non-defined method */
		if method.Ptr == nil {
			continue
		}

		/* (2) Fail on missing description */
		if len(method.Ptr.Description) < 1 {
			return fmt.Errorf("Missing %s.%s description", controllerName, method.Name)
		}

		/* (3) stop if no parameter */
		if method.Ptr.Parameters == nil || len(method.Ptr.Parameters) < 1 {
			method.Ptr.Parameters = make(map[string]*MethodParameter, 0)
			continue
		}

		/* check parameters */
		for pName, pData := range method.Ptr.Parameters {

			/* (4) Fail on invalid rename (set but empty) */
			if pData.Rename != nil && len(*pData.Rename) < 1 {
				return fmt.Errorf("Empty rename for %s.%s parameter '%s'", controllerName, method.Name, pName)
			}

			/* (5) Check for name/rename conflict */
			for paramName, param := range method.Ptr.Parameters {

				// ignore self
				if pName == paramName {
					continue
				}

				// 1. Same rename field
				if pData.Rename != nil && param.Rename != nil && *pData.Rename == *param.Rename {
					return fmt.Errorf("Rename conflict for %s.%s parameter '%s'", controllerName, method.Name, *pData.Rename)
				}

				// 2. Not-renamed field matches a renamed field
				if pData.Rename == nil && param.Rename != nil && pName == *param.Rename {
					return fmt.Errorf("Name conflict for %s.%s parameter '%s'", controllerName, method.Name, pName)
				}

				// 3. Renamed field matches name
				if pData.Rename != nil && param.Rename == nil && *pData.Rename == paramName {
					return fmt.Errorf("Name conflict for %s.%s parameter '%s'", controllerName, method.Name, pName)
				}

			}

			/* (6) Fail on missing description */
			if len(pData.Description) < 1 {
				return fmt.Errorf("Missing description for %s.%s parameter '%s'", controllerName, method.Name, pName)
			}

			/* (7) Fail on missing type */
			if len(pData.Type) < 1 {
				return fmt.Errorf("Missing type for %s.%s parameter '%s'", controllerName, method.Name, pName)
			}
		}

	}

	/* (2) Check child controllers
	---------------------------------------------------------*/
	/* (1) Stop if no child */
	if c.Children == nil || len(c.Children) < 1 {
		return nil
	}

	/* (2) For each controller */
	for ctlName, ctl := range c.Children {

		/* (3) Invalid name */
		if strings.Contains(ctlName, "/") {
			return fmt.Errorf("Controller '%s' must not contain any slash '/'", ctlName)
		}

		/* (4) Check recursively */
		err := ctl.format(ctlName)
		if err != nil {
			return err
		}

	}

	return nil

}

// setDefaults sets the defaults for optional configuration fields
func (c *Controller) setDefaults() {

	/* (1) Get methods */
	methods := []*Method{
		c.GET,
		c.POST,
		c.PUT,
		c.DELETE,
	}

	/* (2) Browse methods */
	for _, m := range methods {

		// ignore if not set
		if m == nil {
			continue
		}

		/* (3) Browse parameters */
		for name, param := range m.Parameters {

			// 1. Default 'opt': required //
			if param.Optional == nil {
				param.Optional = new(bool)
			}

			// 2. Default 'rename': same as name
			if param.Rename == nil {
				param.Rename = &name
			}

		}

	}

	/* (4) Stop here if no children */
	if c.Children == nil || len(c.Children) < 1 {
		return
	}

	/* (5) Iterate over children */
	for _, child := range c.Children {
		child.setDefaults()
	}

	return

}
