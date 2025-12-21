package main

import (
	"fmt"
	"net"

	"github.com/goccy/go-yaml/ast"
)

// ──────── PTR Record Generation ────────
func createMissingPTRs(root *ast.MappingNode) {
	dnsNode := mappingValue(root, "dns_records")
	if dnsNode == nil {
		return
	}

	seq := dnsNode.(*ast.SequenceNode)

	existing := map[string]bool{}
	for _, item := range seq.Values {
		m := item.(*ast.MappingNode)
		if stringValue(m, "type") == "PTR" {
			key := stringValue(m, "zone") + ":" + stringValue(m, "record_value")
			existing[key] = true
		}
	}

	for _, item := range seq.Values {
		m := item.(*ast.MappingNode)
		if stringValue(m, "type") != "A" {
			continue
		}

		ip := net.ParseIP(stringValue(m, "record_value"))
		if ip == nil || ip.To4() == nil {
			continue
		}

		ip4 := ip.To4()
		var zone string
		switch ip4[2] {
		case 0:
			zone = "0.0.10.in-addr.arpa."
		case 1:
			zone = "1.0.10.in-addr.arpa."
		default:
			continue
		}

		last := fmt.Sprintf("%d", ip4[3])
		key := zone + ":" + last
		if existing[key] {
			continue
		}

		ptr := &ast.MappingNode{
			Values: []*ast.MappingValueNode{
				kv("host", stringValue(m, "host")),
				kv("type", "PTR"),
				kv("zone", zone),
				kv("record_value", last),
			},
		}

		seq.Values = append(seq.Values, ptr)
	}
}
