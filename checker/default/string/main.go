package main

func Match(name string) bool {
	return name == "string"
}

func Check(value interface{}) bool {
	_, OK := value.(string)

	return OK
}
