package sc

import "errors"

// ScConstructionCommand represents construction command
type ScConstructionCommand struct {
	Type ScType
	Data interface{}
}

// ScConstruction represents SC construction
type ScConstruction struct {
	Commands []ScConstructionCommand
	Aliases  map[string]int
}

// CreateNode creates a node in construction
func (c *ScConstruction) CreateNode(t ScType, alias string) error {
	if !t.IsNode() {
		return errors.New("you should pass node type there")
	}

	cmd := ScConstructionCommand{Type: t}
	if alias != "" {
		if c.Aliases == nil {
			c.Aliases = make(map[string]int)
		}
		c.Aliases[alias] = len(c.Commands)
	}
	c.Commands = append(c.Commands, cmd)
	return nil
}

// CreateEdge creates an edge in construction
func (c *ScConstruction) CreateEdge(t ScType, src interface{}, trg interface{}, alias string) error {
	if !t.IsEdge() {
		return errors.New("you should pass edge type there")
	}

	cmd := ScConstructionCommand{
		Type: t,
		Data: map[string]interface{}{
			"src": src,
			"trg": trg,
		},
	}

	if alias != "" {
		if c.Aliases == nil {
			c.Aliases = make(map[string]int)
		}
		c.Aliases[alias] = len(c.Commands)
	}
	c.Commands = append(c.Commands, cmd)
	return nil
}

// CreateLink creates a link in construction
func (c *ScConstruction) CreateLink(t ScType, content ScLinkContent, alias string) error {
	if !t.IsLink() {
		return errors.New("you should pass link type there")
	}

	cmd := ScConstructionCommand{
		Type: t,
		Data: map[string]interface{}{
			"content": content.Data,
			"type":    content.Type,
		},
	}

	if alias != "" {
		if c.Aliases == nil {
			c.Aliases = make(map[string]int)
		}
		c.Aliases[alias] = len(c.Commands)
	}
	c.Commands = append(c.Commands, cmd)
	return nil
}

// GetIndex returns index by alias
func (c *ScConstruction) GetIndex(alias string) (int, bool) {
	if c.Aliases == nil {
		return 0, false
	}
	idx, exists := c.Aliases[alias]
	return idx, exists
}
