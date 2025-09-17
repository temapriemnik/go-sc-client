package sc

// SC_TYPE constants
const (
	ScTypeNode            = 0x1
	ScTypeLink            = 0x2
	ScTypeUEdgeCommon     = 0x4
	ScTypeDEdgeCommon     = 0x8
	ScTypeEdgeAccess      = 0x10
	ScTypeConst           = 0x20
	ScTypeVar             = 0x40
	ScTypeEdgePos         = 0x80
	ScTypeEdgeNeg         = 0x100
	ScTypeEdgeFuz         = 0x200
	ScTypeEdgeTemp        = 0x400
	ScTypeEdgePerm        = 0x800
	ScTypeNodeTuple       = 0x80
	ScTypeNodeStruct      = 0x100
	ScTypeNodeRole        = 0x200
	ScTypeNodeNoRole      = 0x400
	ScTypeNodeClass       = 0x800
	ScTypeNodeAbstract    = 0x1000
	ScTypeNodeMaterial    = 0x2000
	ScTypeArcPosConstPerm = ScTypeEdgeAccess | ScTypeConst | ScTypeEdgePos | ScTypeEdgePerm
	ScTypeArcPosVarPerm   = ScTypeEdgeAccess | ScTypeVar | ScTypeEdgePos | ScTypeEdgePerm
)

// ScLinkContentType represents link content type
type ScLinkContentType int

const (
	ScLinkContentInt ScLinkContentType = iota
	ScLinkContentFloat
	ScLinkContentString
	ScLinkContentBinary
)

// ScEventType represents event types
type ScEventType string

const (
	ScEventUnknown            ScEventType = "unknown"
	ScEventAddOutgoingEdge    ScEventType = "add_outgoing_edge"
	ScEventAddIngoingEdge     ScEventType = "add_ingoing_edge"
	ScEventRemoveOutgoingEdge ScEventType = "remove_outgoing_edge"
	ScEventRemoveIngoingEdge  ScEventType = "remove_ingoing_edge"
	ScEventRemoveElement      ScEventType = "delete_element"
	ScEventChangeContent      ScEventType = "content_change"
)
