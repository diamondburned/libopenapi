// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package v2

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/index"
	"github.com/pb33f/libopenapi/utils"
	"github.com/pb33f/libopenapi/utils/typex"
	"gopkg.in/yaml.v3"
)

// Responses is a low-level representation of a Swagger / OpenAPI 2 Responses object.
type Responses struct {
	Codes      typex.Pairs[low.KeyReference[string], low.ValueReference[*Response]]
	Default    low.NodeReference[*Response]
	Extensions typex.Pairs[low.KeyReference[string], low.ValueReference[any]]
}

// GetExtensions returns all Responses extensions and satisfies the low.HasExtensions interface.
func (r *Responses) GetExtensions() typex.Pairs[low.KeyReference[string], low.ValueReference[any]] {
	return r.Extensions
}

// Build will extract default value and extensions from node.
func (r *Responses) Build(root *yaml.Node, idx *index.SpecIndex) error {
	root = utils.NodeAlias(root)
	utils.CheckForMergeNodes(root)
	r.Extensions = low.ExtractExtensions(root)

	if utils.IsNodeMap(root) {
		codes, err := low.ExtractMapNoLookup[*Response](root, idx)
		if err != nil {
			return err
		}
		if codes != nil {
			r.Codes = codes
		}
		def := r.getDefault()
		if def != nil {
			// default is bundled into codes, pull it out
			r.Default = *def
			// remove default from codes
			r.deleteCode(DefaultLabel)
		}
	} else {
		return fmt.Errorf("responses build failed: vn node is not a map! line %d, col %d",
			root.Line, root.Column)
	}
	return nil
}

func (r *Responses) getDefault() *low.NodeReference[*Response] {
	for _, p := range r.Codes {
		n := p.Key
		o := p.Value
		if strings.ToLower(n.Value) == DefaultLabel {
			return &low.NodeReference[*Response]{
				ValueNode: o.ValueNode,
				KeyNode:   n.KeyNode,
				Value:     o.Value,
			}
		}
	}
	return nil
}

// used to remove default from codes extracted by Build()
func (r *Responses) deleteCode(code string) {
	var key *low.KeyReference[string]
	if r.Codes != nil {
		for _, p := range r.Codes {
			if p.Key.Value == code {
				key = &p.Key
				break
			}
		}
	}
	// should never be nil, but, you never know... science and all that!
	if key != nil {
		r.Codes.Delete(*key)
	}
}

// FindResponseByCode will attempt to locate a Response instance using an HTTP response code string.
func (r *Responses) FindResponseByCode(code string) *low.ValueReference[*Response] {
	return low.FindItemInMap[*Response](code, r.Codes)
}

// Hash will return a consistent SHA256 Hash of the Examples object
func (r *Responses) Hash() [32]byte {
	var f []string
	var keys []string
	keys = make([]string, len(r.Codes))
	cmap := make(map[string]*Response, len(keys))
	for i, p := range r.Codes {
		k := p.Key
		keys[i] = k.Value
		cmap[k.Value] = p.Value.Value
	}
	sort.Strings(keys)
	for k := range keys {
		f = append(f, fmt.Sprintf("%s-%s", keys[k], low.GenerateHashString(cmap[keys[k]])))
	}
	if !r.Default.IsEmpty() {
		f = append(f, low.GenerateHashString(r.Default.Value))
	}
	f = append(f, low.GenerateReferencePairsHashes(r.Extensions)...)
	return sha256.Sum256([]byte(strings.Join(f, "|")))
}
