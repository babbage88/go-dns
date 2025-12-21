package main

import (
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

//
// ─── YAML HELPERS ──────────────────────────────────────────────────────────────
//

func mappingValue(m *ast.MappingNode, key string) ast.Node {
	for _, mv := range m.Values {
		if mv.Key.(*ast.StringNode).Value == key {
			return mv.Value
		}
	}
	return nil
}

func stringValue(m *ast.MappingNode, key string) string {
	n := mappingValue(m, key)
	if n == nil {
		return ""
	}
	return n.(*ast.StringNode).Value
}

func commentOut(n ast.Node, reason string) {
	if n == nil {
		return
	}

	text := "DISABLED"
	if reason != "" {
		text = "DISABLED: " + reason
	}

	n.SetComment(
		ast.CommentGroup([]*token.Token{
			{
				Type:  token.CommentType,
				Value: "# " + text,
			},
		}),
	)
}
