package meta

// Default contains the default values when ommited in json
var Default = Schema{
	Host:       "0.0.0.0",
	Port:       80,
	DriverName: "",
	Types: &builder{
		IgnoreBuiltIn: false,
		Folder:        "",
		Map:           nil,
	},
	Controllers: &builder{
		IgnoreBuiltIn: true,
		Folder:        "",
		Map:           nil,
	},
	Middlewares: &builder{
		IgnoreBuiltIn: true,
		Folder:        "",
		Map:           nil,
	},
}
