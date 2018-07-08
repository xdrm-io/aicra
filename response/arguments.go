package response

// Has checks whether a key exists in the arguments
func (i Arguments) Has(key string) bool {
	_, exists := i[key]
	return exists
}
