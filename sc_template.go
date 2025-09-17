package sc

import (
	"fmt"
)

// ScTemplateValue represents template value
type ScTemplateValue struct {
	Value interface{}
	Alias string
}

// ScTemplateTriple represents template triple
type ScTemplateTriple struct {
	Source ScTemplateValue
	Edge   ScTemplateValue
	Target ScTemplateValue
}

// ScTemplate represents SC template
type ScTemplate struct {
	Triples []ScTemplateTriple
}

// Triple adds a triple to template
func (t *ScTemplate) Triple(param1, param2, param3 interface{}) *ScTemplate {
	p1 := t.splitTemplateParam(param1)
	p2 := t.splitTemplateParam(param2)
	p3 := t.splitTemplateParam(param3)

	t.Triples = append(t.Triples, ScTemplateTriple{
		Source: p1,
		Edge:   p2,
		Target: p3,
	})
	return t
}

// TripleWithRelation adds a triple with relation to template
func (t *ScTemplate) TripleWithRelation(param1, param2, param3, param4, param5 interface{}) *ScTemplate {
	p2 := t.splitTemplateParam(param2)
	alias := p2.Alias
	if alias == "" {
		alias = fmt.Sprintf("edge_1_%d", len(t.Triples))
	}

	t.Triple(param1, ScTemplateValue{Value: p2.Value, Alias: alias}, param3)
	t.Triple(param5, param4, alias)
	return t
}

func (t *ScTemplate) splitTemplateParam(param interface{}) ScTemplateValue {
	switch v := param.(type) {
	case []interface{}:
		if len(v) != 2 {
			panic("invalid number of values for replacement. Use [ScType | ScAddr, string]")
		}

		value := v[0]
		alias, ok := v[1].(string)
		if !ok {
			panic("second parameter should be string")
		}

		_, isScAddr := value.(ScAddr)
		_, isScType := value.(ScType)

		if !isScAddr && !isScType {
			panic("first parameter should be ScAddr or ScType")
		}

		return ScTemplateValue{
			Value: value,
			Alias: alias,
		}
	default:
		return ScTemplateValue{
			Value: param,
			Alias: "",
		}
	}
}

// ScTemplateResult represents template search result
type ScTemplateResult struct {
	Addrs   []ScAddr
	Indices map[string]int
}

// Get returns address by alias or index
func (r ScTemplateResult) Get(aliasOrIndex interface{}) ScAddr {
	switch v := aliasOrIndex.(type) {
	case string:
		return r.Addrs[r.Indices[v]]
	case int:
		return r.Addrs[v]
	default:
		panic("aliasOrIndex should be string or int")
	}
}

// Size returns number of addresses in result
func (r ScTemplateResult) Size() int {
	return len(r.Addrs)
}

// ForEachTriple executes function for each triple in result
func (r ScTemplateResult) ForEachTriple(f func(src, edge, trg ScAddr)) {
	for i := 0; i < len(r.Addrs); i += 3 {
		f(r.Addrs[i], r.Addrs[i+1], r.Addrs[i+2])
	}
}
