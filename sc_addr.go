package sc

// ScAddr represents SC address
type ScAddr struct {
	Value int64
}

// IsValid checks if address is valid
func (a ScAddr) IsValid() bool {
	return a.Value != 0
}

// Equal checks if addresses are equal
func (a ScAddr) Equal(other ScAddr) bool {
	return a.Value == other.Value
}
