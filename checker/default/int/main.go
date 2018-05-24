package main

func Match(name string) bool {
	return name == "int"
}

func Check(value interface{}) bool {
	_, intOK := value.(int)
	_, uintOK := value.(uint)

	return intOK || uintOK
}
