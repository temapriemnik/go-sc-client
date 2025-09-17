package sc

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Response represents server response
type Response struct {
	ID      int         `json:"id"`
	Status  bool        `json:"status"`
	Event   bool        `json:"event"`
	Payload interface{} `json:"payload"`
}

// Request represents client request
type Request struct {
	ID      int         `json:"id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ScClient represents SC client
type ScClient struct {
	url          string
	conn         *websocket.Conn
	messageQueue []func()
	callbacks    map[int]func(Response)
	events       map[int]*ScEvent
	eventID      int
	mu           sync.Mutex
	done         chan struct{}
}

// NewScClient creates new SC client
func NewScClient(url string) *ScClient {
	client := &ScClient{
		url:       url,
		callbacks: make(map[int]func(Response)),
		events:    make(map[int]*ScEvent),
		done:      make(chan struct{}),
	}

	go client.connect()
	return client
}

// Connect establishes connection to SC-machine
func (c *ScClient) connect() {
	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		log.Printf("Failed to connect: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
		go c.connect()
		return
	}

	defer c.conn.Close()

	// Process incoming messages
	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v. Reconnecting...", err)
				time.Sleep(5 * time.Second)
				go c.connect()
				return
			}

			var response Response
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("Failed to unmarshal response: %v", err)
				continue
			}

			c.mu.Lock()
			if response.Event {
				if event, exists := c.events[response.ID]; exists {
					payload := response.Payload.([]interface{})
					event.Callback(
						ScAddr{Value: int64(payload[0].(float64))},
						ScAddr{Value: int64(payload[1].(float64))},
						ScAddr{Value: int64(payload[2].(float64))},
						response.ID,
					)
				}
			} else {
				if callback, exists := c.callbacks[response.ID]; exists {
					callback(response)
					delete(c.callbacks, response.ID)
				}
			}
			c.mu.Unlock()
		}
	}
}

// SendMessage sends message to SC-machine
func (c *ScClient) sendMessage(actionType string, payload interface{}, callback func(Response)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventID++
	c.callbacks[c.eventID] = callback

	request := Request{
		ID:      c.eventID,
		Type:    actionType,
		Payload: payload,
	}

	message, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return
	}

	if c.conn != nil {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Write error: %v", err)
			// Add to queue for retry
			c.messageQueue = append(c.messageQueue, func() {
				c.sendMessage(actionType, payload, callback)
			})
		}
	} else {
		// Add to queue if connection is not established
		c.messageQueue = append(c.messageQueue, func() {
			c.sendMessage(actionType, payload, callback)
		})
	}
}

// CheckElements checks elements existence
func (c *ScClient) CheckElements(addrs []ScAddr) ([]ScType, error) {
	if len(addrs) == 0 {
		return []ScType{}, nil
	}

	payload := make([]int64, len(addrs))
	for i, addr := range addrs {
		payload[i] = addr.Value
	}

	result := make(chan []ScType, 1)
	c.sendMessage("check_elements", payload, func(response Response) {
		if !response.Status {
			result <- nil
			return
		}

		types := make([]ScType, len(response.Payload.([]interface{})))
		for i, t := range response.Payload.([]interface{}) {
			types[i] = ScType{Value: int(t.(float64))}
		}
		result <- types
	})

	select {
	case res := <-result:
		if res == nil {
			return nil, errors.New("failed to check elements")
		}
		return res, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while checking elements")
	}
}

// CreateElements creates elements
func (c *ScClient) CreateElements(construction *ScConstruction) ([]ScAddr, error) {
	payload := make([]interface{}, len(construction.Commands))
	for i, cmd := range construction.Commands {
		if cmd.Type.IsNode() {
			payload[i] = map[string]interface{}{
				"el":   "node",
				"type": cmd.Type.Value,
			}
		} else if cmd.Type.IsEdge() {
			data := cmd.Data.(map[string]interface{})
			payload[i] = map[string]interface{}{
				"el":   "edge",
				"type": cmd.Type.Value,
				"src":  c.transformEdgeInfo(construction, data["src"]),
				"trg":  c.transformEdgeInfo(construction, data["trg"]),
			}
		} else if cmd.Type.IsLink() {
			data := cmd.Data.(map[string]interface{})
			payload[i] = map[string]interface{}{
				"el":           "link",
				"type":         cmd.Type.Value,
				"content":      data["content"],
				"content_type": data["type"],
			}
		}
	}

	result := make(chan []ScAddr, 1)
	c.sendMessage("create_elements", payload, func(response Response) {
		if !response.Status {
			result <- nil
			return
		}

		addrs := make([]ScAddr, len(response.Payload.([]interface{})))
		for i, a := range response.Payload.([]interface{}) {
			addrs[i] = ScAddr{Value: int64(a.(float64))}
		}
		result <- addrs
	})

	select {
	case res := <-result:
		if res == nil {
			return nil, errors.New("failed to create elements")
		}
		return res, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while creating elements")
	}
}

func (c *ScClient) transformEdgeInfo(construction *ScConstruction, aliasOrAddr interface{}) map[string]interface{} {
	switch v := aliasOrAddr.(type) {
	case ScAddr:
		return map[string]interface{}{
			"type":  "addr",
			"value": v.Value,
		}
	case string:
		if idx, exists := construction.GetIndex(v); exists {
			return map[string]interface{}{
				"type":  "ref",
				"value": idx,
			}
		}
		panic("invalid alias: " + v)
	default:
		panic("invalid type for edge info")
	}
}

// DeleteElements deletes elements
func (c *ScClient) DeleteElements(addrs []ScAddr) (bool, error) {
	payload := make([]int64, len(addrs))
	for i, addr := range addrs {
		payload[i] = addr.Value
	}

	result := make(chan bool, 1)
	c.sendMessage("delete_elements", payload, func(response Response) {
		result <- response.Status
	})

	select {
	case res := <-result:
		return res, nil
	case <-time.After(30 * time.Second):
		return false, errors.New("timeout while deleting elements")
	}
}

// SetLinkContents sets link contents
func (c *ScClient) SetLinkContents(contents []ScLinkContent) ([]bool, error) {
	payload := make([]interface{}, len(contents))
	for i, content := range contents {
		payload[i] = map[string]interface{}{
			"command": "set",
			"type":    content.TypeToStr(),
			"data":    content.Data,
			"addr":    content.Addr.Value,
		}
	}

	result := make(chan []bool, 1)
	c.sendMessage("content", payload, func(response Response) {
		if !response.Status {
			result <- nil
			return
		}

		results := make([]bool, len(response.Payload.([]interface{})))
		for i, r := range response.Payload.([]interface{}) {
			results[i] = r.(bool)
		}
		result <- results
	})

	select {
	case res := <-result:
		if res == nil {
			return nil, errors.New("failed to set link contents")
		}
		return res, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while setting link contents")
	}
}

// GetLinkContents gets link contents
func (c *ScClient) GetLinkContents(addrs []ScAddr) ([]ScLinkContent, error) {
	payload := make([]interface{}, len(addrs))
	for i, addr := range addrs {
		payload[i] = map[string]interface{}{
			"command": "get",
			"addr":    addr.Value,
		}
	}

	result := make(chan []ScLinkContent, 1)
	c.sendMessage("content", payload, func(response Response) {
		if !response.Status {
			result <- nil
			return
		}

		contents := make([]ScLinkContent, 0)
		for _, item := range response.Payload.([]interface{}) {
			itemMap := item.(map[string]interface{})
			if value, exists := itemMap["value"]; exists && value != nil {
				content := ScLinkContent{
					Data: value,
					Type: StringToType(itemMap["type"].(string)),
				}
				contents = append(contents, content)
			}
		}
		result <- contents
	})

	select {
	case res := <-result:
		if res == nil {
			return nil, errors.New("failed to get link contents")
		}
		return res, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while getting link contents")
	}
}

// ResolveKeynodes resolves keynodes
func (c *ScClient) ResolveKeynodes(params map[string]ScType) (map[string]ScAddr, error) {
	payload := make([]interface{}, 0, len(params))
	for id, t := range params {
		if t.IsValid() {
			payload = append(payload, map[string]interface{}{
				"command": "resolve",
				"idtf":    id,
				"elType":  t.Value,
			})
		} else {
			payload = append(payload, map[string]interface{}{
				"command": "find",
				"idtf":    id,
			})
		}
	}

	result := make(chan map[string]ScAddr, 1)
	c.sendMessage("keynodes", payload, func(response Response) {
		if !response.Status {
			result <- nil
			return
		}

		addrs := make([]ScAddr, len(response.Payload.([]interface{})))
		for i, a := range response.Payload.([]interface{}) {
			addrs[i] = ScAddr{Value: int64(a.(float64))}
		}

		res := make(map[string]ScAddr)
		i := 0
		for id := range params {
			res[id] = addrs[i]
			i++
		}
		result <- res
	})

	select {
	case res := <-result:
		if res == nil {
			return nil, errors.New("failed to resolve keynodes")
		}
		return res, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while resolving keynodes")
	}
}

// TemplateSearch searches by template
func (c *ScClient) TemplateSearch(template *ScTemplate) ([]ScTemplateResult, error) {
	payload := c.prepareTemplatePayload(template)

	result := make(chan []ScTemplateResult, 1)
	c.sendMessage("search_template", payload, func(response Response) {
		if !response.Status {
			result <- nil
			return
		}

		responseData := response.Payload.(map[string]interface{})
		aliases := responseData["aliases"].(map[string]interface{})
		addrsData := responseData["addrs"].([]interface{})

		results := make([]ScTemplateResult, len(addrsData))
		for i, addrList := range addrsData {
			addrValues := addrList.([]interface{})
			addrs := make([]ScAddr, len(addrValues))
			for j, addr := range addrValues {
				addrs[j] = ScAddr{Value: int64(addr.(float64))}
			}

			aliasIndices := make(map[string]int)
			for alias, index := range aliases {
				aliasIndices[alias] = int(index.(float64))
			}

			results[i] = ScTemplateResult{
				Addrs:   addrs,
				Indices: aliasIndices,
			}
		}
		result <- results
	})

	select {
	case res := <-result:
		if res == nil {
			return nil, errors.New("failed to search template")
		}
		return res, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while searching template")
	}
}

// TemplateGenerate generates elements by template
func (c *ScClient) TemplateGenerate(template *ScTemplate, params map[string]ScAddr) (*ScTemplateResult, error) {
	payload := map[string]interface{}{
		"templ":  c.prepareTemplatePayload(template),
		"params": c.prepareTemplateParams(params),
	}

	result := make(chan *ScTemplateResult, 1)
	c.sendMessage("generate_template", payload, func(response Response) {
		if !response.Status {
			result <- nil
			return
		}

		responseData := response.Payload.(map[string]interface{})
		aliases := responseData["aliases"].(map[string]interface{})
		addrsData := responseData["addrs"].([]interface{})

		addrs := make([]ScAddr, len(addrsData))
		for i, addr := range addrsData {
			addrs[i] = ScAddr{Value: int64(addr.(float64))}
		}

		aliasIndices := make(map[string]int)
		for alias, index := range aliases {
			aliasIndices[alias] = int(index.(float64))
		}

		result <- &ScTemplateResult{
			Addrs:   addrs,
			Indices: aliasIndices,
		}
	})

	select {
	case res := <-result:
		if res == nil {
			return nil, errors.New("failed to generate template")
		}
		return res, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while generating template")
	}
}

func (c *ScClient) prepareTemplatePayload(template *ScTemplate) []interface{} {
	payload := make([]interface{}, len(template.Triples))
	for i, triple := range template.Triples {
		payload[i] = []interface{}{
			c.processTripleItem(triple.Source),
			c.processTripleItem(triple.Edge),
			c.processTripleItem(triple.Target),
		}
	}
	return payload
}

func (c *ScClient) processTripleItem(item ScTemplateValue) map[string]interface{} {
	result := make(map[string]interface{})
	if item.Alias != "" {
		result["alias"] = item.Alias
	}

	switch v := item.Value.(type) {
	case ScAddr:
		result["type"] = "addr"
		result["value"] = v.Value
	case ScType:
		result["type"] = "type"
		result["value"] = v.Value
	case string:
		result["type"] = "alias"
		result["value"] = v
	default:
		panic("invalid triple item type")
	}
	return result
}

func (c *ScClient) prepareTemplateParams(params map[string]ScAddr) map[string]interface{} {
	result := make(map[string]interface{})
	for key, addr := range params {
		result[key] = addr.Value
	}
	return result
}

// EventsCreate creates events
func (c *ScClient) EventsCreate(events []ScEventParams) ([]ScEvent, error) {
	payload := make([]interface{}, len(events))
	for i, event := range events {
		payload[i] = map[string]interface{}{
			"type": event.Type,
			"addr": event.Addr.Value,
		}
	}

	requestPayload := map[string]interface{}{
		"create": payload,
	}

	result := make(chan []ScEvent, 1)
	c.sendMessage("events", requestPayload, func(response Response) {
		if !response.Status {
			result <- nil
			return
		}

		eventIDs := response.Payload.([]interface{})
		createdEvents := make([]ScEvent, len(events))
		for i, id := range eventIDs {
			eventID := int(id.(float64))
			createdEvents[i] = ScEvent{
				ID:       eventID,
				Type:     events[i].Type,
				Callback: events[i].Callback,
			}
			c.events[eventID] = &createdEvents[i]
		}
		result <- createdEvents
	})

	select {
	case res := <-result:
		if res == nil {
			return nil, errors.New("failed to create events")
		}
		return res, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while creating events")
	}
}

// EventsDestroy destroys events
func (c *ScClient) EventsDestroy(eventIDs []int) error {
	requestPayload := map[string]interface{}{
		"delete": eventIDs,
	}

	result := make(chan bool, 1)
	c.sendMessage("events", requestPayload, func(response Response) {
		result <- response.Status
	})

	select {
	case res := <-result:
		if !res {
			return errors.New("failed to destroy events")
		}

		// Remove events from internal storage
		c.mu.Lock()
		for _, id := range eventIDs {
			delete(c.events, id)
		}
		c.mu.Unlock()

		return nil
	case <-time.After(30 * time.Second):
		return errors.New("timeout while destroying events")
	}
}

// Close closes connection
func (c *ScClient) Close() {
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
}
