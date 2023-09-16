package uri

import (
	"fmt"
	"io"
	"net/url"
	"testing"
)

func Benchmark_Parse(b *testing.B) {
	tests := []string{
		"foo://example.com:8042/over/there?name=ferret#nose",
		"http://httpbin.org/get?utf8=%e2%98%83",
		"mailto://user@domain.com",
		"ssh://user@git.openstack.org:29418/openstack/keystone.git",
		"https://willo.io/#yolo",
	}

	b.ReportAllocs()
	b.ResetTimer()

	var v URI
	for i := 0; i < b.N; i++ {
		v, _ = Parse(tests[i%len(tests)])
	}
	fmt.Fprintln(io.Discard, v)
}

func Benchmark_ParseURLAsURI(b *testing.B) {
	tests := []string{
		"http://httpbin.org/get?utf8=%e2%98%83",
		"ssh://user@git.openstack.org:29418/openstack/keystone.git",
		"https://willo.io/#yolo",
	}

	b.ReportAllocs()
	b.ResetTimer()

	var v URI
	for i := 0; i < b.N; i++ {
		v, _ = Parse(tests[i%len(tests)])
	}
	fmt.Fprintln(io.Discard, v)
}

func Benchmark_ParseURLStdLib(b *testing.B) {
	tests := []string{
		"http://httpbin.org/get?utf8=%e2%98%83",
		"ssh://user@git.openstack.org:29418/openstack/keystone.git",
		"https://willo.io/#yolo",
	}

	b.ReportAllocs()
	b.ResetTimer()

	var u *url.URL
	for i := 0; i < b.N; i++ {
		u, _ = url.Parse(tests[i%len(tests)])
	}
	fmt.Fprintln(io.Discard, u)
}

func Benchmark_String(b *testing.B) {
	tests := []*uri{
		{
			"foo", "//example.com:8042/over/there", "name=ferret", "nose",
			&authorityInfo{"//", "", "example.com", "8042", "/over/there", false},
		},
		{
			"http", "//httpbin.org/get", "utf8=\xe2\x98\x83", "",
			&authorityInfo{"//", "", "httpbin.org", "", "/get", false},
		},
		{
			"mailto", "user@domain.com", "", "",
			&authorityInfo{"//", "user", "domain.com", "", "", false},
		},
		{
			"ssh", "//user@git.openstack.org:29418/openstack/keystone.git", "", "",
			&authorityInfo{"//", "user", "git.openstack.org", "29418", "/openstack/keystone.git", false},
		},
		{
			"https", "//willo.io/", "", "yolo",
			&authorityInfo{"//", "", "willo.io", "", "/", false},
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
