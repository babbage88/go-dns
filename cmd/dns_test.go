package cmd

import (
	"testing"

	"github.com/goccy/go-yaml/ast"
)

func mapNode(fields ...[2]string) *ast.MappingNode {
	values := make([]*ast.MappingValueNode, 0, len(fields))
	for _, f := range fields {
		values = append(values, &ast.MappingValueNode{
			Key:   &ast.StringNode{Value: f[0]},
			Value: &ast.StringNode{Value: f[1]},
		})
	}
	return &ast.MappingNode{Values: values}
}

func seqNode(items ...*ast.MappingNode) *ast.SequenceNode {
	values := make([]ast.Node, 0, len(items))
	for _, item := range items {
		values = append(values, item)
	}
	return &ast.SequenceNode{Values: values}
}

func rootWithSection(name string, seq *ast.SequenceNode) *ast.MappingNode {
	return &ast.MappingNode{
		Values: []*ast.MappingValueNode{{
			Key:   &ast.StringNode{Value: name},
			Value: seq,
		}},
	}
}

func TestCollectNameserverJobs_EmptyRoot(t *testing.T) {
	root := &ast.MappingNode{Values: []*ast.MappingValueNode{}}

	jobs := collectNameserverJobs(root)
	if len(jobs) != 0 {
		t.Fatalf("collectNameserverJobs returned %d jobs, want 0", len(jobs))
	}
}

func TestCollectNameserverJobs_CollectsAndDeduplicates(t *testing.T) {
	root := &ast.MappingNode{Values: []*ast.MappingValueNode{
		{
			Key: &ast.StringNode{Value: "nameservers"},
			Value: seqNode(
				mapNode([2]string{"ip_address", "8.8.8.8"}, [2]string{"name", "google"}),
				mapNode([2]string{"ip_address", "1.1.1.1"}, [2]string{"name", "cloudflare"}),
				mapNode([2]string{"ip_address", "8.8.8.8"}, [2]string{"name", "google-duplicate"}),
				mapNode([2]string{"name", "missing-ip"}),
			),
		},
	}}

	jobs := collectNameserverJobs(root)
	if len(jobs) != 2 {
		t.Fatalf("collectNameserverJobs returned %d jobs, want 2", len(jobs))
	}

	got := map[string]bool{}
	for _, j := range jobs {
		got[j.IP] = true
	}
	if !got["8.8.8.8"] || !got["1.1.1.1"] {
		t.Fatalf("collectNameserverJobs IPs = %#v, want 8.8.8.8 and 1.1.1.1", got)
	}
}

func TestCollectNameserverJobs_PrefixKeyIncluded(t *testing.T) {
	root := &ast.MappingNode{Values: []*ast.MappingValueNode{
		{
			Key:   &ast.StringNode{Value: "nameservers_internal"},
			Value: seqNode(mapNode([2]string{"ip_address", "10.0.0.53"})),
		},
	}}

	jobs := collectNameserverJobs(root)
	if len(jobs) != 1 {
		t.Fatalf("collectNameserverJobs returned %d jobs, want 1", len(jobs))
	}
	if jobs[0].IP != "10.0.0.53" {
		t.Fatalf("collectNameserverJobs IP = %q, want 10.0.0.53", jobs[0].IP)
	}
}

func TestCollectDNSRecordJobs_EmptyRoot(t *testing.T) {
	root := &ast.MappingNode{Values: []*ast.MappingValueNode{}}

	jobs := collectDNSRecordJobs(root)
	if len(jobs) != 0 {
		t.Fatalf("collectDNSRecordJobs returned %d jobs, want 0", len(jobs))
	}
}

func TestCollectDNSRecordJobs_CollectsFromBothSectionsAndDeduplicates(t *testing.T) {
	root := &ast.MappingNode{Values: []*ast.MappingValueNode{
		{
			Key: &ast.StringNode{Value: "dns_records"},
			Value: seqNode(
				mapNode([2]string{"type", "A"}, [2]string{"record_value", "10.0.0.10"}),
				mapNode([2]string{"type", "AAAA"}, [2]string{"record_value", "2001:db8::1"}),
				mapNode([2]string{"type", "MX"}, [2]string{"record_value", "mail.example.com"}),
				mapNode([2]string{"type", "A"}, [2]string{"record_value", ""}),
			),
		},
		{
			Key: &ast.StringNode{Value: "sub_zone_records"},
			Value: seqNode(
				mapNode([2]string{"type", "A"}, [2]string{"record_value", "10.0.0.10"}),
				mapNode([2]string{"type", "A"}, [2]string{"record_value", "10.1.0.20"}),
			),
		},
	}}

	jobs := collectDNSRecordJobs(root)
	if len(jobs) != 3 {
		t.Fatalf("collectDNSRecordJobs returned %d jobs, want 3", len(jobs))
	}

	got := map[string]bool{}
	for _, j := range jobs {
		got[j.IP] = true
	}
	for _, ip := range []string{"10.0.0.10", "10.1.0.20", "2001:db8::1"} {
		if !got[ip] {
			t.Fatalf("collectDNSRecordJobs missing IP %q in %#v", ip, got)
		}
	}
}

func TestCreateMissingPTRs_NoRecordsSection(t *testing.T) {
	root := &ast.MappingNode{Values: []*ast.MappingValueNode{}}
	createMissingPTRs(root)
}

func TestCreateMissingPTRs_SkipsNonAAndOutOfRange(t *testing.T) {
	root := rootWithSection("dns_records", seqNode(
		mapNode([2]string{"type", "AAAA"}, [2]string{"record_value", "2001:db8::1"}),
		mapNode([2]string{"type", "A"}, [2]string{"record_value", "8.8.8.8"}),
		mapNode([2]string{"type", "A"}, [2]string{"record_value", "invalid"}),
	))

	seq := root.Values[0].Value.(*ast.SequenceNode)
	before := len(seq.Values)
	createMissingPTRs(root)
	if len(seq.Values) != before {
		t.Fatalf("createMissingPTRs changed record count to %d, want %d", len(seq.Values), before)
	}
}

func TestCreateMissingPTRs_CreatesPTRFor10Dot0And10Dot1(t *testing.T) {
	root := rootWithSection("dns_records", seqNode(
		mapNode(
			[2]string{"host", "svc-a"},
			[2]string{"type", "A"},
			[2]string{"record_value", "10.0.0.5"},
		),
		mapNode(
			[2]string{"host", "svc-b"},
			[2]string{"type", "A"},
			[2]string{"record_value", "10.0.1.6"},
		),
	))

	seq := root.Values[0].Value.(*ast.SequenceNode)
	createMissingPTRs(root)

	if len(seq.Values) != 4 {
		t.Fatalf("createMissingPTRs created %d records, want 4", len(seq.Values))
	}

	ptr1 := seq.Values[2].(*ast.MappingNode)
	if stringValue(ptr1, "type") != "PTR" || stringValue(ptr1, "host") != "svc-a" ||
		stringValue(ptr1, "zone") != "0.0.10.in-addr.arpa." || stringValue(ptr1, "record_value") != "5" {
		t.Fatalf("unexpected PTR1: type=%q host=%q zone=%q record_value=%q",
			stringValue(ptr1, "type"), stringValue(ptr1, "host"), stringValue(ptr1, "zone"), stringValue(ptr1, "record_value"))
	}

	ptr2 := seq.Values[3].(*ast.MappingNode)
	if stringValue(ptr2, "type") != "PTR" || stringValue(ptr2, "host") != "svc-b" ||
		stringValue(ptr2, "zone") != "1.0.10.in-addr.arpa." || stringValue(ptr2, "record_value") != "6" {
		t.Fatalf("unexpected PTR2: type=%q host=%q zone=%q record_value=%q",
			stringValue(ptr2, "type"), stringValue(ptr2, "host"), stringValue(ptr2, "zone"), stringValue(ptr2, "record_value"))
	}
}

func TestCreateMissingPTRs_DoesNotDuplicateExistingPTR(t *testing.T) {
	root := rootWithSection("dns_records", seqNode(
		mapNode(
			[2]string{"type", "PTR"},
			[2]string{"zone", "0.0.10.in-addr.arpa."},
			[2]string{"record_value", "5"},
		),
		mapNode(
			[2]string{"host", "svc"},
			[2]string{"type", "A"},
			[2]string{"record_value", "10.0.0.5"},
		),
	))

	seq := root.Values[0].Value.(*ast.SequenceNode)
	before := len(seq.Values)
	createMissingPTRs(root)
	if len(seq.Values) != before {
		t.Fatalf("createMissingPTRs created duplicate PTR: got %d records, want %d", len(seq.Values), before)
	}
}
