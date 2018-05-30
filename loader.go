package gfw

import (
	"git.xdrm.io/xdrm-brackets/gfw/checker"
	"git.xdrm.io/xdrm-brackets/gfw/internal/config"
)

// Init initilises a new framework instance
//
// - path is the configuration path
//
// - if typeChecker is nil, defaults will be used (all *.so files
//   inside ./types local directory)
//
func Init(path string, typeChecker *checker.TypeRegistry) (*Server, error) {

	/* (1) Init instance */
	inst := &Server{
		config: nil,
		Params: make(map[string]interface{}),
		err:    ErrSuccess,
	}

	/* (2) Load configuration */
	config, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	inst.config = config

	/* (3) Store registry if not nil */
	if typeChecker != nil {
		inst.Checker = typeChecker
		return inst, nil
	}

	/* (4) Default registry creation */
	inst.Checker = checker.CreateRegistry(true)

	return inst, nil
}
