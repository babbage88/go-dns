package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

const pingTimeout = 2 * time.Second

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <dns.yaml>\n", os.Args[0])
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	file, err := yaml.Parse(data)
	if err != nil {
		panic(err)
	}

	root := file.Docs[0].Body.(*ast.MappingNode)

	handleNameservers(root)
	handleDNSRecords(root)
	createMissingPTRs(root)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(file); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}

//
// ─── PING LOGIC ────────────────────────────────────────────────────────────────
//

func ping(ip string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", ip)
	return cmd.Run() == nil
}

//
// ─── YAML HELPERS ──────────────────────────────────────────────────────────────
//

func commentOut(node ast.Node, reason string) {
	node.SetComment(&ast.CommentGroup{
		Head: []*ast.Comment{
			{
				Text: fmt.Sprintf("# DISABLED: %s", reason),
			},
		},
	})
}

func mappingValue(m *ast.MappingNode, key string) ast.Node {
	for i := 0; i < len(m.Values); i += 2 {
		k := m.Values[i].(*ast.StringNode)
		if k.Value == key {
			return m.Values[i+1]
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

//
// ─── STEP 1: NAMESERVERS ───────────────────────────────────────────────────────
//

func handleNameservers(root *ast.MappingNode) {
	for i := 0; i < len(root.Values); i += 2 {
		key := root.Values[i].(*ast.StringNode).Value
		if !strings.HasPrefix(key, "nameservers") {
			continue
		}

		seq := root.Values[i+1].(*ast.SequenceNode)
		for _, item := range seq.Values {
			m := item.(*ast.MappingNode)
			ip := stringValue(m, "ip_address")

			if ip != "" && !ping(ip) {
				commentOut(item, "nameserver unreachable")
			}
		}
	}
}

//
// ─── STEP 2: DNS RECORDS ───────────────────────────────────────────────────────
//

func handleDNSRecords(root *ast.MappingNode) {
	for _, section := range []string{"dns_records", "sub_zone_records"} {
		seq := mappingValue(root, section)
		if seq == nil {
			continue
		}

		for _, item := range seq.(*ast.SequenceNode).Values {
			m := item.(*ast.MappingNode)
			rtype := stringValue(m, "type")

			if rtype == "A" || rtype == "AAAA" {
				ip := stringValue(m, "record_value")
				if ip != "" && !ping(ip) {
					commentOut(item, "A/AAAA record unreachable")
				}
			}
		}
	}
}

//
// ─── STEP 3: PTR CREATION ──────────────────────────────────────────────────────
//

func createMissingPTRs(root *ast.MappingNode) {
	dnsSeq := mappingValue(root, "dns_records").(*ast.SequenceNode)

	existingPTR := map[string]bool{}
	for _, item := range dnsSeq.Values {
		m := item.(*ast.MappingNode)
		if stringValue(m, "type") == "PTR" {
			zone := stringValue(m, "zone")
			val := stringValue(m, "record_value")
			existingPTR[zone+":"+val] = true
		}
	}

	for _, item := range dnsSeq.Values {
		m := item.(*ast.MappingNode)
		if stringValue(m, "type") != "A" {
			continue
		}

		ip := net.ParseIP(stringValue(m, "record_value"))
		if ip == nil {
			continue
		}

		ip4 := ip.To4()
		if ip4 == nil {
			continue
		}

		zone := ""
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

		if existingPTR[key] {
			continue
		}

		ptr := &ast.MappingNode{
			Values: []ast.Node{
				&ast.StringNode{Value: "host"},
				&ast.StringNode{Value: stringValue(m, "host")},
				&ast.StringNode{Value: "type"},
				&ast.StringNode{Value: "PTR"},
				&ast.StringNode{Value: "zone"},
				&ast.StringNode{Value: zone},
				&ast.StringNode{Value: "record_value"},
				&ast.StringNode{Value: last},
			},
		}

		dnsSeq.Values = append(dnsSeq.Values, ptr)
	}
}
