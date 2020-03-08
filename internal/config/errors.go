package config

import "git.xdrm.io/go/aicra/internal/cerr"

// ErrRead - a problem ocurred when trying to read the configuration file
const ErrRead = cerr.Error("cannot read config")

// ErrFormat - a invalid format has been detected
const ErrFormat = cerr.Error("invalid config format")

// ErrIllegalServiceName - an illegal character has been found in a service name
const ErrIllegalServiceName = cerr.Error("service must not contain any slash '/' nor '-' symbols")

// ErrMissingMethodDesc - a method is missing its description
const ErrMissingMethodDesc = cerr.Error("missing method description")

// ErrMissingParamDesc - a parameter is missing its description
const ErrMissingParamDesc = cerr.Error("missing parameter description")

// ErrIllegalParamName - a parameter has an illegal name
const ErrIllegalParamName = cerr.Error("illegal parameter name (must not begin/end with '_')")

// ErrMissingParamType - a parameter has an illegal type
const ErrMissingParamType = cerr.Error("missing parameter type")

// ErrParamNameConflict - a parameter has a conflict with its name/rename field
const ErrParamNameConflict = cerr.Error("name conflict for parameter")
