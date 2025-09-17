package sc

// ScEventCallbackFunc represents event callback function
type ScEventCallbackFunc func(elAddr, edge, other ScAddr, eventID int)

// ScEvent represents SC event
type ScEvent struct {
	ID       int
	Type     ScEventType
	Callback ScEventCallbackFunc
}

// IsValid checks if event is valid
func (e ScEvent) IsValid() bool {
	return e.ID > 0
}

// ScEventParams represents event parameters
type ScEventParams struct {
	Addr     ScAddr
	Type     ScEventType
	Callback ScEventCallbackFunc
}
