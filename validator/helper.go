package validator

type num interface {
	uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64 | float32 | float64
}

// casts any number of type N into a fixed number type C. Returns the cast value
// and whether the cast is valid (precision lost)
func castNumber[N num, C num](val N) (C, bool) {
	var (
		cast       = C(val)
		reversible = val == N(cast)
		overflow   = float64(cast) != float64(val)
	)
	return cast, reversible && !overflow
}
