package main

// Match matches the string 'any'
func Match(name string) bool {
	return name == "any"
}

// Check always returns true
func Check(value interface{}) bool {
	return true
}
