package config

// Default contains the default values when omitted in json
var Default = Schema{
	Root:       ".",
	Host:       "0.0.0.0",
	Port:       80,
	DriverName: "",
	Types: &builder{
		Default: true,
		Folder:  "type",
	},
	Controllers: &builder{
		Default: false,
		Folder:  "controller",
	},
	Middlewares: &builder{
		Default: false,
		Folder:  "middleware",
	},
}
