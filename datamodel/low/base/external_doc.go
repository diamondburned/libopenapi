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

// ExternalDoc represents a low-level External Documentation object as defined by OpenAPI 2 and 3
//
// Allows referencing an external resource for extended documentation.
//
//	v2 - https://swagger.io/specification/v2/#externalDocumentationObject
//	v3 - https://spec.openapis.org/oas/v3.1.0#external-documentation-object
type ExternalDoc struct {
	Description low.NodeReference[string]
	URL         low.NodeReference[string]
	Extensions  typex.Pairs[low.KeyReference[string], low.ValueReference[any]]
	*low.Reference
}

// FindExtension returns a ValueReference containing the extension value, if found.
func (ex *ExternalDoc) FindExtension(ext string) *low.ValueReference[any] {
	return low.FindItemInMap[any](ext, ex.Extensions)
}

// Build will extract extensions from the ExternalDoc instance.
func (ex *ExternalDoc) Build(root *yaml.Node, idx *index.SpecIndex) error {
	root = utils.NodeAlias(root)
	utils.CheckForMergeNodes(root)
	ex.Reference = new(low.Reference)
	ex.Extensions = low.ExtractExtensions(root)
	return nil
}

// GetExtensions returns all ExternalDoc extensions and satisfies the low.HasExtensions interface.
func (ex *ExternalDoc) GetExtensions() typex.Pairs[low.KeyReference[string], low.ValueReference[any]] {
	return ex.Extensions
}

func (ex *ExternalDoc) Hash() [32]byte {
	// calculate a hash from every property.
	f := []string{
		ex.Description.Value,
		ex.URL.Value,
	}
	f = append(f, low.GenerateReferencePairsHashes(ex.Extensions)...)
	return sha256.Sum256([]byte(strings.Join(f, "|")))
}
