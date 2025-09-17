package sc

import (
	"fmt"
)

// ScSet represents SC set
type ScSet struct {
	Client       *ScClient
	Addr         ScAddr
	Elements     map[int64]ScAddr
	OnAdd        func(ScAddr) error
	OnRemove     func(ScAddr) error
	OnInitialize func([]ScAddr) error
	FilterType   *ScType
	AddEvent     *ScEvent
	RemoveEvent  *ScEvent
}

// NewScSet creates new SC set
func NewScSet(client *ScClient, addr ScAddr, onInitialize, onAdd, onRemove func([]ScAddr) error, filterType *ScType) (*ScSet, error) {
	if !addr.IsValid() {
		return nil, fmt.Errorf("invalid addr of set: %v", addr)
	}

	set := &ScSet{
		Client:     client,
		Addr:       addr,
		Elements:   make(map[int64]ScAddr),
		FilterType: filterType,
	}

	// Adapt callback functions
	set.OnInitialize = func(addrs []ScAddr) error {
		return onInitialize(addrs)
	}
	set.OnAdd = func(addr ScAddr) error {
		return onAdd([]ScAddr{addr})
	}
	set.OnRemove = func(addr ScAddr) error {
		return onRemove([]ScAddr{addr})
	}

	return set, nil
}

// Initialize initializes set
func (s *ScSet) Initialize() error {
	// Create events for adding and removing elements
	events, err := s.Client.EventsCreate([]ScEventParams{
		{
			Addr:     s.Addr,
			Type:     ScEventAddOutgoingEdge,
			Callback: s.onEventAddElement,
		},
		{
			Addr:     s.Addr,
			Type:     ScEventRemoveOutgoingEdge,
			Callback: s.onEventRemoveElement,
		},
	})
	if err != nil {
		return err
	}

	s.AddEvent = &events[0]
	s.RemoveEvent = &events[1]

	// Iterate existing elements
	return s.iterateExistingElements()
}

func (s *ScSet) onEventAddElement(elAddr, edge, other ScAddr, eventID int) {
	if _, exists := s.Elements[edge.Value]; !exists {
		if elAddr.IsValid() {
			if shouldAppend, _ := s.shouldAppend([]ScAddr{elAddr}); shouldAppend[0] {
				s.Elements[edge.Value] = elAddr
				s.OnAdd(elAddr)
			}
		}
	}
}

func (s *ScSet) onEventRemoveElement(elAddr, edge, other ScAddr, eventID int) {
	if trg, exists := s.Elements[edge.Value]; exists {
		s.OnRemove(trg)
		delete(s.Elements, edge.Value)
	}
}

func (s *ScSet) shouldAppend(addrs []ScAddr) ([]bool, error) {
	if s.FilterType == nil {
		result := make([]bool, len(addrs))
		for i := range result {
			result[i] = true
		}
		return result, nil
	}

	types, err := s.Client.CheckElements(addrs)
	if err != nil {
		return nil, err
	}

	result := make([]bool, len(types))
	for i, t := range types {
		result[i] = (s.FilterType.Value & t.Value) == s.FilterType.Value
	}
	return result, nil
}

func (s *ScSet) iterateExistingElements() error {
	template := &ScTemplate{}
	template.Triple(
		s.Addr,
		[]interface{}{ScType{Value: ScTypeArcPosVarPerm}, "_edge"},
		[]interface{}{ScType{Value: 0}, "_item"},
	)

	results, err := s.Client.TemplateSearch(template)
	if err != nil {
		return err
	}

	var elements []ScAddr
	for _, result := range results {
		edge := result.Get("_edge")
		item := result.Get("_item")

		if shouldAppend, _ := s.shouldAppend([]ScAddr{item}); shouldAppend[0] {
			s.Elements[edge.Value] = item
			elements = append(elements, item)
		}
	}

	return s.OnInitialize(elements)
}

// AddItem adds item to set
func (s *ScSet) AddItem(addr ScAddr) (bool, error) {
	template := &ScTemplate{}
	template.Triple(
		s.Addr,
		[]interface{}{ScType{Value: ScTypeArcPosVarPerm}, "_edge"},
		addr,
	)

	results, err := s.Client.TemplateSearch(template)
	if err != nil {
		return false, err
	}

	if len(results) == 0 {
		genResult, err := s.Client.TemplateGenerate(template, map[string]ScAddr{})
		if err != nil {
			return false, err
		}
		return genResult != nil, nil
	}

	return true, nil
}
