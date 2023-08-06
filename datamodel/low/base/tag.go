// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package base

import (
	"crypto/sha256"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/index"
	"github.com/pb33f/libopenapi/utils"
	"github.com/pb33f/libopenapi/utils/typex"
	"gopkg.in/yaml.v3"
)

// Tag represents a low-level Tag instance that is backed by a low-level one.
//
// Adds metadata to a single tag that is used by the Operation Object. It is not mandatory to have a Tag Object per
// tag defined in the Operation Object instances.
//   - v2: https://swagger.io/specification/v2/#tagObject
//   - v3: https://swagger.io/specification/#tag-object
type Tag struct {
	Name         low.NodeReference[string]
	Description  low.NodeReference[string]
	ExternalDocs low.NodeReference[*ExternalDoc]
	Extensions   typex.Pairs[low.KeyReference[string], low.ValueReference[any]]
	*low.Reference
}

// FindExtension returns a ValueReference containing the extension value, if found.
func (t *Tag) FindExtension(ext string) *low.ValueReference[any] {
	return low.FindItemInMap[any](ext, t.Extensions)
}

// Build will extract extensions and external docs for the Tag.
func (t *Tag) Build(root *yaml.Node, idx *index.SpecIndex) error {
	root = utils.NodeAlias(root)
	utils.CheckForMergeNodes(root)
	t.Reference = new(low.Reference)
	t.Extensions = low.ExtractExtensions(root)

	// extract externalDocs
	extDocs, err := low.ExtractObject[*ExternalDoc](ExternalDocsLabel, root, idx)
	t.ExternalDocs = extDocs
	return err
}

// GetExtensions returns all Tag extensions and satisfies the low.HasExtensions interface.
func (t *Tag) GetExtensions() typex.Pairs[low.KeyReference[string], low.ValueReference[any]] {
	return t.Extensions
}

// Hash will return a consistent SHA256 Hash of the Info object
func (t *Tag) Hash() [32]byte {
	var f []string
	if !t.Name.IsEmpty() {
		f = append(f, t.Name.Value)
	}
	if !t.Description.IsEmpty() {
		f = append(f, t.Description.Value)
	}
	if !t.ExternalDocs.IsEmpty() {
		f = append(f, low.GenerateHashString(t.ExternalDocs.Value))
	}
	f = append(f, low.GenerateReferencePairsHashes(t.Extensions)...)
	return sha256.Sum256([]byte(strings.Join(f, "|")))
}

// TODO: future mutation API experiment code is here. this snippet is to re-marshal the object.
//func (t *Tag) MarshalYAML() (interface{}, error) {
//	m := make(map[string]interface{})
//	for i := range t.Extensions {
//		m[i.Value] = t.Extensions[i].Value
//	}
//	if t.Name.Value != "" {
//		m[NameLabel] = t.Name.Value
//	}
//	if t.Description.Value != "" {
//		m[DescriptionLabel] = t.Description.Value
//	}
//	if t.ExternalDocs.Value != nil {
//		m[ExternalDocsLabel] = t.ExternalDocs.Value
//	}
//	return m, nil
//}
//
//func NewTag() *Tag {
//	return new(Tag)
//}
