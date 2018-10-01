package meta

// Default contains the default values when ommited in json
var Default = Schema{
	Host:       "0.0.0.0",
	Port:       80,
	DriverName: "",
	Types: &builder{
		Default: true,
		Folder:  "",
		Map:     nil,
	},
	Controllers: &builder{
		Default: false,
		Folder:  "",
		Map:     nil,
	},
	Middlewares: &builder{
		Default: false,
		Folder:  "",
		Map:     nil,
	},
}
