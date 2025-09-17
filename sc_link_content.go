package sc

// ScLinkContent represents SC link content
type ScLinkContent struct {
	Data interface{}
	Type ScLinkContentType
	Addr *ScAddr
}

// TypeToStr returns string representation of content type
func (c ScLinkContent) TypeToStr() string {
	switch c.Type {
	case ScLinkContentBinary:
		return "binary"
	case ScLinkContentFloat:
		return "float"
	case ScLinkContentInt:
		return "int"
	default:
		return "string"
	}
}

// StringToType converts string to content type
func StringToType(s string) ScLinkContentType {
	switch s {
	case "binary":
		return ScLinkContentBinary
	case "float":
		return ScLinkContentFloat
	case "int":
		return ScLinkContentInt
	default:
		return ScLinkContentString
	}
}
