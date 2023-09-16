package uri

import "testing"

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

	for i := 0; i < b.N; i++ {
		_, _ = Parse(tests[i%5])
	}
}

func Benchmark_String(b *testing.B) {
	tests := []*uri{
		{
			"foo", "//example.com:8042/over/there", "name=ferret", "nose",
			&authorityInfo{"//", "", "example.com", "8042", "/over/there"},
		},
		{
			"http", "//httpbin.org/get", "utf8=\xe2\x98\x83", "",
			&authorityInfo{"//", "", "httpbin.org", "", "/get"},
		},
		{
			"mailto", "user@domain.com", "", "",
			&authorityInfo{"//", "user", "domain.com", "", ""},
		},
		{
			"ssh", "//user@git.openstack.org:29418/openstack/keystone.git", "", "",
			&authorityInfo{"//", "user", "git.openstack.org", "29418", "/openstack/keystone.git"},
		},
		{
			"https", "//willo.io/", "", "yolo",
			&authorityInfo{"//", "", "willo.io", "", "/"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tests[i%5].String()
	}
}
