package gfw

import "git.xdrm.io/gfw/internal/config"

// Init initilises a new framework instance
func Init(path string) (*Server, error) {

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

	return inst, nil
}
