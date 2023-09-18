package uri

import (
	"fmt"
	"io"
	"net/url"
	"testing"
)

type payload struct {
	name    string
	comment string
	tests   []string
}

func payloads() []payload {
	return []payload{
		{
			name:    "simple payload",
			comment: "a payload with vanilla URIs, no IPs, no escaped content",
			tests: []string{
				"foo://example.com:8042/over/there?name=ferret#nose",
				"mailto://user@domain.com",
				"ssh://user@git.openstack.org:29418/openstack/keystone.git",
				"https://willo.io/#yolo",
			},
		},
		{
			name:    "mixed payload",
			comment: "a payload with vanilla URIs, with 20% containing escaped content",
			tests: []string{
				"foo://example.com:8042/over/there?name=ferret#nose",
				"mailto://user@domain.com",
				"ssh://user@git.openstack.org:29418/openstack/keystone.git",
				"https://willo.io/#yolo",
				"http://httpbin.org/get?utf8=%e2%98%83",
			},
		},
		{
			name:    "payload with IPs",
			comment: "a payload with ~ 30% with an IP host specification (v4 & v6)",
			tests: []string{
				"foo://example.com:8042/over/there?name=ferret#nose",
				"http://httpbin.org/get?utf8=%e2%98%83",
				"mailto://user@domain.com",
				"ssh://user@git.openstack.org:29418/openstack/keystone.git",
				"https://willo.io/#yolo",
				"https://user:passwd@[FF02:30:0:0:0:0:0:5%25en1]:8080/a?query=value#fragment",
				"https://user:passwd@127.0.0.1:8080/a?query=value#fragment",
			},
		},
	}
}

func Benchmark_Parse(b *testing.B) {
	// URI.Parse() and net/url.URL.Parse() side by side on different payloads
	for _, payload := range payloads() {
		b.Run(fmt.Sprintf("with URI %s", payload.name), benchParseWithPayload(payload.tests))
		b.Run(fmt.Sprintf("with URL %s", payload.name), benchParseURLStdLib(payload.tests))
	}
}

func benchParseWithPayload(payload []string) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		var v URI
		for i := 0; i < b.N; i++ {
			v, _ = Parse(payload[i%len(payload)])
		}
		fmt.Fprintln(io.Discard, v)
	}
}

func benchParseURLStdLib(payload []string) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		var u *url.URL
		for i := 0; i < b.N; i++ {
			u, _ = url.Parse(payload[i%len(payload)])
		}
		fmt.Fprintln(io.Discard, u)
	}
}

func Benchmark_String(b *testing.B) {
	tests := []*uri{
		{
			"foo", "//example.com:8042/over/there", "name=ferret", "nose",
			authorityInfo{"//", "", "example.com", "8042", "/over/there", false},
		},
		{
			"http", "//httpbin.org/get", "utf8=\xe2\x98\x83", "",
			authorityInfo{"//", "", "httpbin.org", "", "/get", false},
		},
		{
			"mailto", "user@domain.com", "", "",
			authorityInfo{"//", "user", "domain.com", "", "", false},
		},
		{
			"ssh", "//user@git.openstack.org:29418/openstack/keystone.git", "", "",
			authorityInfo{"//", "user", "git.openstack.org", "29418", "/openstack/keystone.git", false},
		},
		{
			"https", "//willo.io/", "", "yolo",
			authorityInfo{"//", "", "willo.io", "", "/", false},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	var s string
	for i := 0; i < b.N; i++ {
		s = tests[i%5].String()
	}
	fmt.Fprintln(io.Discard, s)
}
