package sc

// ScType represents SC type
type ScType struct {
	Value int
}

// IsNode checks if type is a node
func (t ScType) IsNode() bool {
	return (t.Value & ScTypeNode) != 0
}

// IsEdge checks if type is an edge
func (t ScType) IsEdge() bool {
	return (t.Value & (ScTypeEdgeAccess | ScTypeDEdgeCommon | ScTypeUEdgeCommon)) != 0
}

// IsLink checks if type is a link
func (t ScType) IsLink() bool {
	return (t.Value & ScTypeLink) != 0
}

// IsConst checks if type is constant
func (t ScType) IsConst() bool {
	return (t.Value & ScTypeConst) != 0
}

// IsVar checks if type is variable
func (t ScType) IsVar() bool {
	return (t.Value & ScTypeVar) != 0
}

// IsValid checks if type is valid
func (t ScType) IsValid() bool {
	return t.Value != 0
}

// Equal checks if types are equal
func (t ScType) Equal(other ScType) bool {
	return t.Value == other.Value
}
