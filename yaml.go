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

func commentOut(node ast.Node, reason string) {
	node.SetComment(
		ast.CommentGroup(
			[]*token.Token{
				{
					Type:  token.CommentType,
					Value: "# DISABLED: " + reason,
				},
			},
		),
	)
}
