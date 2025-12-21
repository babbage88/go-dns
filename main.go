package main

import (
	"fmt"
	"net"
	"os"

	"github.com/alecthomas/kong"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

//
// ─── MAIN ──────────────────────────────────────────────────────────────────────
//

func main() {
	var cli CLI
	kong.Parse(&cli)

	data, err := os.ReadFile(cli.File)
	if err != nil {
		panic(err)
	}

	file, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	root := file.Docs[0].Body.(*ast.MappingNode)

	nsJobs := collectNameserverJobs(root)
	dnsJobs := collectDNSRecordJobs(root)

	allJobs := append(nsJobs, dnsJobs...)

	results := runPingWorkers(allJobs, cli.Timeout, cli.Workers)

	// Apply results SINGLE-THREADED
	for _, r := range results {
		if !r.ok {
			commentOut(r.job.Node, r.job.Why)
		}
	}

	createMissingPTRs(root)

	if cli.DryRun {
		fmt.Println("# dry-run enabled, no output written")
		return
	}

	out, err := yaml.Marshal(file)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}

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

//
// ─── STEP 3: PTR GENERATION ────────────────────────────────────────────────────
//

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

func kv(k, v string) *ast.MappingValueNode {
	return &ast.MappingValueNode{
		Key:   &ast.StringNode{Value: k},
		Value: &ast.StringNode{Value: v},
	}
}
