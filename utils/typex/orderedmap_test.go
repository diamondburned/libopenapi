package typex

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOrderedMapUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		parse   func(b []byte) (any, error)
		want    any
		wantErr bool
	}{
		{
			name:  "json_empty",
			input: `{}`,
			parse: func(b []byte) (any, error) {
				m := Pairs[string, int]{}
				return &m, json.Unmarshal(b, &m)
			},
			want: &Pairs[string, int]{},
		},
		{
			name:  "json_string:int",
			input: `{ "b": 1, "a": 2 }`,
			parse: func(b []byte) (any, error) {
				var m Pairs[string, int]
				return &m, json.Unmarshal(b, &m)
			},
			want: &Pairs[string, int]{
				{Key: "b", Value: 1},
				{Key: "a", Value: 2},
			},
		},
		{
			name:  "json_string:any",
			input: `{ "b": 1, "a": "2" }`,
			parse: func(b []byte) (any, error) {
				var m Pairs[string, any]
				return &m, json.Unmarshal(b, &m)
			},
			want: &Pairs[string, any]{
				{Key: "b", Value: 1.0},
				{Key: "a", Value: "2"},
			},
		},
		{
			name:  "json_string:any-object",
			input: `{ "b": { "c": { "d": 3 } } }`,
			parse: func(b []byte) (any, error) {
				var m Pairs[string, map[string]map[string]int]
				return &m, json.Unmarshal(b, &m)
			},
			want: &Pairs[string, map[string]map[string]int]{
				{Key: "b", Value: map[string]map[string]int{
					"c": {"d": 3},
				}},
			},
		},
		{
			name:  "json_bad_string:any-object",
			input: `{ "b": 3 }`,
			parse: func(b []byte) (any, error) {
				var m Pairs[string, string]
				return &m, json.Unmarshal(b, &m)
			},
			wantErr: true,
		},
		{
			name:  "json_bad_object",
			input: `{ "b": 3,`,
			parse: func(b []byte) (any, error) {
				var m Pairs[string, any]
				return &m, json.Unmarshal(b, &m)
			},
			wantErr: true,
		},
		{
			name:  "yaml_empty",
			input: ``,
			parse: func(b []byte) (any, error) {
				m := Pairs[string, int]{}
				return &m, yaml.Unmarshal(b, &m)
			},
			want: &Pairs[string, int]{},
		},
		{
			name:  "yaml_string:int",
			input: "b: 1\na: 2\n",
			parse: func(b []byte) (any, error) {
				var m Pairs[string, int]
				return &m, yaml.Unmarshal(b, &m)
			},
			want: &Pairs[string, int]{
				{Key: "b", Value: 1},
				{Key: "a", Value: 2},
			},
		},
		{
			name:  "yaml_string:any",
			input: "b: 1\na: [2]\n",
			parse: func(b []byte) (any, error) {
				var m Pairs[string, any]
				return &m, yaml.Unmarshal(b, &m)
			},
			want: &Pairs[string, any]{
				{Key: "b", Value: 1},
				{Key: "a", Value: []any{2}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.parse([]byte(test.input))
			if err != nil {
				if test.wantErr {
					return
				}
				t.Errorf("parse error: %v", err)
			}
			if test.wantErr {
				t.Errorf("expected error, got none")
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected result:\n"+
					"got  %#v\n"+
					"want %#v", got, test.want)
			}
		})
	}
}

func TestOrderedMapMarshal(t *testing.T) {
	type testKind string

	const (
		jsonTest testKind = "json"
		yamlTest testKind = "yaml"
	)

	tests := []struct {
		name  string
		kind  testKind
		input any
		want  string
	}{
		{
			name:  "json_empty",
			kind:  jsonTest,
			input: &Pairs[string, int]{},
			want:  `{}`,
		},
		{
			name: "json_string:int",
			kind: jsonTest,
			input: &Pairs[string, int]{
				{Key: "b", Value: 1},
				{Key: "a", Value: 2},
			},
			want: `{"b":1,"a":2}`,
		},
		{
			name: "json_string:any",
			kind: jsonTest,
			input: &Pairs[string, any]{
				{Key: "b", Value: 1.0},
				{Key: "a", Value: map[string]int{"c": 3}},
			},
			want: `{"b":1,"a":{"c":3}}`,
		},
		{
			name:  "yaml_empty",
			kind:  yamlTest,
			input: &Pairs[string, int]{},
			want:  `{}`,
		},
		{
			name: "yaml_string:int",
			kind: yamlTest,
			input: &Pairs[string, int]{
				{Key: "b", Value: 1},
				{Key: "a", Value: 2},
			},
			want: "b: 1\na: 2\n",
		},
		{
			name: "yaml_string:any",
			kind: yamlTest,
			input: &Pairs[string, any]{
				{Key: "b", Value: 1},
				{Key: "a", Value: []any{2}},
			},
			want: "b: 1\na:\n    - 2\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got []byte
			var err error
			switch test.kind {
			case jsonTest:
				got, err = json.Marshal(test.input)
			case yamlTest:
				got, err = yaml.Marshal(test.input)
			default:
				t.Fatalf("unknown test kind: %v", test.kind)
			}
			if err != nil {
				t.Errorf("marshal error: %v", err)
			}

			got = bytes.TrimSpace(got)
			want := strings.TrimSpace(test.want)
			if string(got) != want {
				t.Errorf("unexpected result:\n"+
					"got  %q\n"+
					"want %q", got, test.want)
			}
		})
	}
}
