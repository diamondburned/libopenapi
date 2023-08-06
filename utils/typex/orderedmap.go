package typex

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Pairs is a list of key-value pairs. It is used to implement ordered maps.
type Pairs[K comparable, V any] []Pair[K, V]

// Pair is a key-value pair.
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// Push adds a new key-value pair to the map.
// It does not check for duplicate keys.
func (m *Pairs[K, V]) Push(key K, value V) {
	*m = append(*m, Pair[K, V]{key, value})
}

// Get returns the value associated with the given key.
func (m Pairs[K, V]) Get(key K) (V, bool) {
	for _, p := range m {
		if p.Key == key {
			return p.Value, true
		}
	}
	var z V
	return z, false
}

// Getz returns the value associated with the given key. If the key is not
// found, it returns the zero value of the value type.
func (m Pairs[K, V]) Getz(key K) V {
	v, _ := m.Get(key)
	return v
}

// Delete removes one key-value pair with the given key.
func (m *Pairs[K, V]) Delete(key K) {
	for i, p := range *m {
		if p.Key == key {
			m.DeleteAt(i)
			return
		}
	}
}

// DeleteAt removes the key-value pair at the given index.
func (m *Pairs[K, V]) DeleteAt(i int) {
	*m = append((*m)[:i], (*m)[i+1:]...)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *Pairs[K, V]) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))

	if err := expectToken(dec, json.Delim('{')); err != nil {
		return err
	}

	newMap := Pairs[K, V]{}
	for dec.More() {
		var p Pair[K, V]

		kt, err := dec.Token()
		if err != nil {
			return fmt.Errorf("cannot decode key: %w", err)
		}

		// Do this in case the key implements json.Unmarshaler.
		jsonKey, err := json.Marshal(kt)
		if err != nil {
			return fmt.Errorf("cannot encode key: %w", err)
		}
		if err := json.Unmarshal(jsonKey, &p.Key); err != nil {
			return fmt.Errorf("cannot decode key: %w", err)
		}

		if err := dec.Decode(&p.Value); err != nil {
			return fmt.Errorf("cannot decode value: %w", err)
		}

		newMap = append(newMap, p)
	}

	if err := expectToken(dec, json.Delim('}')); err != nil {
		return err
	}

	*m = newMap
	return nil
}

func expectToken(dec *json.Decoder, expected json.Token) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if t != expected {
		return fmt.Errorf("expected %v, got %v", expected, t)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (m Pairs[K, V]) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	for i, p := range m {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := json.NewEncoder(buf).Encode(p.Key); err != nil {
			return nil, fmt.Errorf("cannot encode key: %w", err)
		}
		buf.WriteString(`:`)
		if err := json.NewEncoder(buf).Encode(p.Value); err != nil {
			return nil, fmt.Errorf("cannot encode value: %w", err)
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (m *Pairs[K, V]) UnmarshalYAML(obj *yaml.Node) error {
	if obj.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %v", obj.Kind)
	}

	if len(obj.Content)%2 != 0 {
		return fmt.Errorf("expected even number of nodes, got %v", len(obj.Content))
	}

	newMap := Pairs[K, V]{}
	for i := 0; i < len(obj.Content); i += 2 {
		k := obj.Content[i]
		v := obj.Content[i+1]

		var p Pair[K, V]
		if err := k.Decode(&p.Key); err != nil {
			return fmt.Errorf("cannot decode key: %w", err)
		}
		if err := v.Decode(&p.Value); err != nil {
			return fmt.Errorf("cannot decode value: %w", err)
		}

		newMap = append(newMap, p)
	}

	*m = newMap
	return nil
}

func (m Pairs[K, V]) MarshalYAML() (interface{}, error) {
	var obj yaml.Node
	obj.Kind = yaml.MappingNode
	obj.Content = make([]*yaml.Node, 0, len(m)*2)
	for _, p := range m {
		var k yaml.Node
		var v yaml.Node
		if err := k.Encode(p.Key); err != nil {
			return nil, fmt.Errorf("cannot encode key: %w", err)
		}
		if err := v.Encode(p.Value); err != nil {
			return nil, fmt.Errorf("cannot encode value: %w", err)
		}
		obj.Content = append(obj.Content, &k, &v)
	}
	return obj, nil
}
