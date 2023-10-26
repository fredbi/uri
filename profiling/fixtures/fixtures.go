package fixtures

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fredbi/uri"
)

type (
	URITest struct {
		URIRaw      string
		Err         error
		Comment     string
		IsReference bool
		IsNotURI    bool
	}

	testGenerator func() []URITest
)

var AllGenerators = []testGenerator{
	rawParsePassTests,
	rawParseFailTests,
	rawParseReferenceTests,
	rawParseStructureTests,
	rawParseSchemeTests,
	rawParseUserInfoTests,
	rawParsePathTests,
	rawParseHostTests,
	rawParseIPHostTests,
	rawParsePortTests,
	rawParseQueryTests,
	rawParseFragmentTests,
}

func rawParseReferenceTests() []URITest {
	return []URITest{
		{
			Comment:     "valid missing scheme for an URI reference",
			URIRaw:      "//foo.bar/?baz=qux#quux",
			IsReference: true,
		},
		{
			Comment:     "valid URI reference (not a valid URI)",
			URIRaw:      "//host.domain.com/a/b",
			IsReference: true,
		},
		{
			Comment:     "valid URI reference with port (not a valid URI)",
			URIRaw:      "//host.domain.com:8080/a/b",
			IsReference: true,
		},
		{
			Comment:     "absolute reference with port",
			URIRaw:      "//host.domain.com:8080/a/b",
			IsReference: true,
		},
		{
			Comment:     "absolute reference with query params",
			URIRaw:      "//host.domain.com:8080?query=x/a/b",
			IsReference: true,
		},
		{
			Comment:     "absolute reference with query params (reversed)",
			URIRaw:      "//host.domain.com:8080/a/b?query=x",
			IsReference: true,
		},
		{
			Comment:     "invalid URI which is a valid reference",
			URIRaw:      "//not.a.user@not.a.host/just/a/path",
			IsReference: true,
		},
		{
			Comment:     "not an URI but a valid reference",
			URIRaw:      "/",
			IsReference: true,
		},
		{
			Comment:     "URL is an URI reference",
			URIRaw:      "//not.a.user@not.a.host/just/a/path",
			IsReference: true,
		},
		{
			Comment:     "URL is an URI reference, with escaped host",
			URIRaw:      "//not.a.user@%66%6f%6f.com/just/a/path/also",
			IsReference: true,
		},
		{
			Comment:     "non letter is an URI reference",
			URIRaw:      "*",
			IsReference: true,
		},
		{
			Comment:     "file name is an URI reference",
			URIRaw:      "foo.html",
			IsReference: true,
		},
		{
			Comment:     "directory is an URI reference",
			URIRaw:      "../dir/",
			IsReference: true,
		},
		{
			Comment:     "empty string is an URI reference",
			URIRaw:      "",
			IsReference: true,
		},
	}
}

func rawParseStructureTests() []URITest {
	return []URITest{
		{
			Comment: "// without // prefix, this is parsed as a path",
			URIRaw:  "mailto:user@domain.com",
		},
		{
			Comment: "with // prefix, this parsed as a user + host",
			URIRaw:  "mailto://user@domain.com",
		},
		{
			Comment: "pathological input (1)",
			URIRaw:  "?//x",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "pathological input (2)",
			URIRaw:  "#//x",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "pathological input (3)",
			URIRaw:  "://x",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "pathological input (4)",
			URIRaw:  ".?:",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "pathological input (5)",
			URIRaw:  ".#:",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "pathological input (6)",
			URIRaw:  "?",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "pathological input (7)",
			URIRaw:  "#",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "pathological input (8)",
			URIRaw:  "?#",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "invalid empty URI",
			URIRaw:  "",
			Err:     uri.ErrNoSchemeFound,
		},
		{
			Comment: "invalid with blank",
			URIRaw:  " ",
			Err:     uri.ErrNoSchemeFound,
		},
		{
			Comment: "no separator",
			URIRaw:  "foo",
			Err:     uri.ErrNoSchemeFound,
		},
		{
			Comment: "no ':' separator",
			URIRaw:  "foo@bar",
			Err:     uri.ErrNoSchemeFound,
		},
	}
}

func rawParseSchemeTests() []URITest {
	return []URITest{
		{
			Comment: "urn scheme",
			URIRaw:  "urn://example-bin.org/path",
		},
		{
			Comment: "only scheme (DNS host), valid!",
			URIRaw:  "http:",
		},
		{
			Comment: "only scheme (registered name empty), valid!",
			URIRaw:  "foo:",
		},
		{
			Comment: "scheme without prefix (urn)",
			URIRaw:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
		},
		{
			Comment: "scheme without prefix (urn-like)",
			URIRaw:  "news:comp.infosystems.www.servers.unix",
		},
		{
			Comment: "+ and - in scheme (e.g. tel resource)",
			URIRaw:  "tel:+1-816-555-1212",
		},
		{
			Comment: "should assert scheme",
			URIRaw:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
		},
		{
			Comment: "legit separator in scheme",
			URIRaw:  "http+unix://%2Fvar%2Frun%2Fsocket/path?key=value",
		},
		{
			Comment: "with scheme only",
			URIRaw:  "https:",
		},
		{
			Comment: "empty scheme",
			URIRaw:  "://bob/",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "invalid scheme (should start with a letter) (2)",
			URIRaw:  "?invalidscheme://www.example.com",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "invalid scheme (invalid character) (2)",
			URIRaw:  "ht?tps:",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "relative URIs with a colon (':') in their first segment are not considered well-formed",
			URIRaw:  "2013.05.29_14:33:41",
			Err:     uri.ErrInvalidScheme,
		},
		{
			Comment: "invalid scheme (should start with a letter) (1)",
			URIRaw:  "1http://bob",
			Err:     uri.ErrInvalidScheme,
		},
		{
			Comment: "invalid scheme (too short)",
			URIRaw:  "x://bob",
			Err:     uri.ErrInvalidScheme,
		},
		{
			Comment: "invalid scheme (invalid character) (1)",
			URIRaw:  "x{}y://bob",
			Err:     uri.ErrInvalidScheme,
		},
		{
			Comment: "invalid scheme (invalid character) (3)",
			URIRaw:  "inv;alidscheme://www.example.com",
			Err:     uri.ErrInvalidScheme,
		},
		{
			Comment: "absolute URI that represents an implicit file URI.",
			URIRaw:  "c:\\directory\filename",
			Err:     uri.ErrInvalidScheme,
		},
		{
			Comment: "represents a hierarchical absolute URI and does not contain '://'",
			URIRaw:  "www.contoso.com/path/file",
			Err:     uri.ErrNoSchemeFound,
		},
	}
}

func rawParsePathTests() []URITest {
	return []URITest{
		{
			Comment: "legitimate use of several starting /'s in path'",
			URIRaw:  "file://hostname//etc/hosts",
		},
		{
			Comment: "authority is not empty: valid path with double '/' (see issue#3)",
			URIRaw:  "http://host:8080//foo.html",
		},
		{
			Comment: "path",
			URIRaw:  "https://example-bin.org/path?",
		},
		{
			Comment: "empty path, query and fragment",
			URIRaw:  "mailto://u:p@host.domain.com?#",
		},
		{
			Comment: "empty path",
			URIRaw:  "http://foo.com",
		},
		{
			Comment: "path only, no query, no fragmeny",
			URIRaw:  "http://foo.com/path",
		},
		{
			Comment: "path with escaped spaces",
			URIRaw:  "http://example.w3.org/path%20with%20spaces.html",
		},
		{
			Comment: "path is just an escape space",
			URIRaw:  "http://example.w3.org/%20",
		},
		{
			Comment: "dots in path",
			URIRaw:  "ftp://ftp.is.co.za/../../../rfc/rfc1808.txt",
		},
		{
			Comment: "= in path",
			URIRaw:  "ldap://[2001:db8::7]/c=GB?objectClass?one",
		},
		{
			Comment: "path with drive letter (e.g. windows) (1)",
			// this one is dubious: Microsoft (.Net) recognizes the C:/... string as a path and
			// states this as incorrect uri -- all other validators state a host "c" and state this uri as a valid one
			URIRaw: "file://c:/directory/filename",
		},
		{
			Comment: "path with drive letter (e.g. windows) (2)",
			// The unambiguous correct URI notation is  file:///c:/directory/filename
			URIRaw: "file:///c:/directory/filename",
		},
		{
			Comment: `if a URI does not contain an authority component,
		then the path cannot begin with two slash characters ("//").`,
			URIRaw: "https:////a?query=value#fragment",
			Err:    uri.ErrInvalidPath,
		},
		{
			Comment: "contains unescaped backslashes even if they will be treated as forward slashes",
			URIRaw:  "http:\\host/path/file",
			Err:     uri.ErrInvalidPath,
		},
		{
			Comment: "invalid path (invalid characters)",
			URIRaw:  "http://www.example.org/hello/{}yzx;=1.1/world.txt/?id=5&part=three#there-you-go",
			Err:     uri.ErrInvalidPath,
		},
		{
			Comment: "should detect a path starting with several /'s",
			URIRaw:  "file:////etc/hosts",
			Err:     uri.ErrInvalidPath,
		},
		{
			Comment: "empty host  => double '/' invalid in this context",
			URIRaw:  "http:////foo.html",
			Err:     uri.ErrInvalidPath,
		},
		{
			Comment: "trailing empty fragment, invalid path",
			URIRaw:  "http://example.w3.org/%legit#",
			Err:     uri.ErrInvalidPath,
		},
		{
			Comment: "partial escape (1)",
			URIRaw:  "http://example.w3.org/%a",
			Err:     uri.ErrInvalidPath,
		},
		{
			Comment: "partial escape (2)",
			URIRaw:  "http://example.w3.org/%a/foo",
			Err:     uri.ErrInvalidPath,
		},
		{
			Comment: "partial escape (3)",
			URIRaw:  "http://example.w3.org/%illegal",
			Err:     uri.ErrInvalidPath,
		},
	}
}

func rawParseHostTests() []URITest {
	return []URITest{
		{
			Comment: "authorized dash '-' in host",
			URIRaw:  "https://example-bin.org/path",
		},
		{
			Comment: "host with many segments",
			URIRaw:  "ftp://ftp.is.co.za/rfc/rfc1808.txt",
		},
		{
			Comment: "percent encoded host is valid, with encoded character not being valid",
			URIRaw:  "urn://user:passwd@ex%7Cample.com:8080/a?query=value#fragment",
		},
		{
			Comment: "valid percent-encoded host (dash character is allowed in registered name)",
			URIRaw:  "urn://user:passwd@ex%2Dample.com:8080/a?query=value#fragment",
		},
		{
			Comment: "check percent encoding with DNS hostname, dash allowed in DNS name",
			URIRaw:  "https://user:passwd@ex%2Dample.com:8080/a?query=value#fragment",
		},
		{
			Comment: "should error on empty host",
			URIRaw:  "https://user:passwd@:8080/a?query=value#fragment",
			Err:     uri.ErrMissingHost,
		},
		{
			Comment: "unicode host",
			URIRaw:  "http://www.詹姆斯.org/",
		},
		{
			Comment: "illegal characters",
			URIRaw:  "http://<foo>",
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "should detect an invalid host (DNS rule) (1)",
			URIRaw:  "https://user:passwd@286;0.0.1:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "should detect an invalid host (DNS rule) (2)",
			URIRaw:  "https://user:passwd@256.256.256.256:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "registered name containing unallowed character",
			URIRaw:  "bob://x|y/",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "invalid host (contains blank space)",
			URIRaw:  "http://www.exa mple.org",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "DNS hostname is too long",
			URIRaw:  fmt.Sprintf("https://%s/", strings.Repeat("x", 256)),
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "DNS segment in hostname is empty",
			URIRaw:  "https://seg..com/",
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "DNS last segment in hostname is empty",
			URIRaw:  "https://seg.empty.com./",
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "DNS segment ends with unallowed character",
			URIRaw:  "https://x-.y.com/",
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "DNS segment in hostname too long",
			URIRaw:  fmt.Sprintf("https://%s.%s.com/", strings.Repeat("x", 63), strings.Repeat("y", 64)),
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "DNS with all segments empty",
			URIRaw:  "https://........./",
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "DNS segment ends with incomplete escape sequence",
			URIRaw:  "https://x.y.com%/",
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "DNS segment contains an invalid rune",
			URIRaw:  fmt.Sprintf("https://x.y.com%s/", string([]rune{utf8.RuneError})),
			Err:     uri.ErrInvalidDNSName,
		},
	}
}

func rawParseIPHostTests() []URITest {
	return []URITest{
		{
			Comment: "IPv6 host",
			URIRaw:  "mailto://user@[fe80::1]",
		},
		{
			Comment: "IPv6 host with zone",
			URIRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en1]:8080/a?query=value#fragment",
		},
		{
			Comment: "ipv4 host",
			URIRaw:  "https://user:passwd@127.0.0.1:8080/a?query=value#fragment",
		},
		{
			Comment: "IPv4 host",
			URIRaw:  "http://192.168.0.1/",
		},
		{
			Comment: "IPv4 host with port",
			URIRaw:  "http://192.168.0.1:8080/",
		},
		{
			Comment: "IPv6 host",
			URIRaw:  "http://[fe80::1]/",
		},
		{
			Comment: "IPv6 host with port",
			URIRaw:  "http://[fe80::1]:8080/",
		},
		{
			Comment: "IPv6 host with zone",
			URIRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25lo]:8080/a?query=value#fragment",
		},
		// Tests exercising RFC 6874 compliance:
		{
			Comment: "IPv6 host with (escaped) zone identifier",
			URIRaw:  "http://[fe80::1%25en0]/",
		},
		{
			Comment: "IPv6 host with zone identifier and port",
			URIRaw:  "http://[fe80::1%25en0]:8080/",
		},
		{
			Comment: "IPv6 host with percent-encoded+unreserved zone identifier",
			URIRaw:  "http://[fe80::1%25%65%6e%301-._~]/",
		},
		{
			Comment: "IPv6 host with percent-encoded+unreserved zone identifier",
			URIRaw:  "http://[fe80::1%25%65%6e%301-._~]:8080/",
		},
		{
			Comment: "IPv6 host with invalid percent-encoding in zone identifier",
			URIRaw:  "http://[fe80::1%25%C3~]:8080/",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPv6 host with invalid percent-encoding in zone identifier",
			URIRaw:  "http://[fe80::1%25%F3~]:8080/",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IP v4 host (escaped) %31 is percent-encoded for '1'",
			URIRaw:  "http://192.168.0.%31/",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "IPv4 address with percent-encoding is not allowed",
			URIRaw:  "http://192.168.0.%31:8080/",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "invalid IPv4 with port (2)",
			URIRaw:  "https://user:passwd@127.256.0.1:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "invalid IPv4 with port (3)",
			URIRaw:  "https://user:passwd@127.0127.0.1:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "valid IPv4 with port (1)",
			URIRaw:  "https://user:passwd@127.0.0.1:8080/a?query=value#fragment",
		},
		{
			Comment: "invalid IPv4: part>255",
			URIRaw:  "https://user:passwd@256.256.256.256:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "IPv6 percent-encoding is limited to ZoneID specification, mus be %25",
			URIRaw:  "http://[fe80::%31]/",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPv6 percent-encoding is limited to ZoneID specification, mus be %25 (2))",
			URIRaw:  "http://[fe80::%31]:8080/",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPv6 percent-encoding is limited to ZoneID specification, mus be %25 (2))",
			URIRaw:  "http://[fe80::%31%25en0]/",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPv6 percent-encoding is limited to ZoneID specification, mus be %25 (2))",
			URIRaw:  "http://[%310:fe80::%25en0]/",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			URIRaw: "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment",
		},
		{
			URIRaw: "https://user:passwd@[FF02:30:0:0:0:0:0:5%25lo]:8080/a?query=value#fragment",
		},
		{
			Comment: "IPv6 with wrong percent encoding",
			URIRaw:  "http://[fe80::%%31]:8080/",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPv6 with wrong percent encoding",
			URIRaw:  "http://[fe80::%26lo]:8080/",
			Err:     uri.ErrInvalidHostAddress,
		},
		// These two cases are valid as textual representations as
		// described in RFC 4007, but are not valid as address
		// literals with IPv6 zone identifiers in URIs as described in
		// RFC 6874.
		{
			Comment: "invalid IPv6 (double empty ::)",
			URIRaw:  "https://user:passwd@[FF02::3::5]:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "invalid IPv6 host with empty (escaped) zone",
			URIRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25]:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "invalid IPv6 with unescaped zone (bad percent-encoding)",
			URIRaw:  "https://user:passwd@[FADF:01%en0]:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "should not parse IPv6 host with empty zone (bad percent encoding)",
			URIRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%]:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "empty IPv6",
			URIRaw:  "scheme://user:passwd@[]/valid",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "zero IPv6",
			URIRaw:  "scheme://user:passwd@[::]/valid",
		},
		{
			Comment: "invalid IPv6 (lack closing bracket) (1)",
			URIRaw:  "http://[fe80::1/",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "invalid IPv6 (lack closing bracket) (2)",
			URIRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidURI,
		},
		{
			Comment: "invalid IPv6 (lack closing bracket) (3)",
			URIRaw:  "https://user:passwd@[FADF:01%en0:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "missing closing bracket for IPv6 litteral (1)",
			URIRaw:  "https://user:passwd@[FF02::3::5:8080",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "missing closing bracket for IPv6 litteral (2)",
			URIRaw:  "https://user:passwd@[FF02::3::5:8080/?#",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "missing closing bracket for IPv6 litteral (3)",
			URIRaw:  "https://user:passwd@[FF02::3::5:8080#",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "missing closing bracket for IPv6 litteral (4)",
			URIRaw:  "https://user:passwd@[FF02::3::5:8080#abc",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPv6 empty zone",
			URIRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25]:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPv6 unescaped zone with reserved characters",
			URIRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25:lo]:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPv6 addresses not between square brackets are invalid hosts (1)",
			URIRaw:  "https://0%3A0%3A0%3A0%3A0%3A0%3A0%3A1/a",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "IPv6 addresses not between square brackets are invalid hosts (2)",
			URIRaw:  "https://FF02:30:0:0:0:0:0:5%25/a",
			Err:     uri.ErrInvalidPort, // ':30' parses as a port
		},
		{
			Comment: "IP addresses between square brackets should not be ipv4 addresses",
			URIRaw:  "https://[192.169.224.1]/a",
			Err:     uri.ErrInvalidHostAddress,
		},
		// Just for fun: IPvFuture...
		{
			Comment: "IPvFuture address",
			URIRaw:  "http://[v6.fe80::a_en1]",
		},
		{
			Comment: "IPvFuture address",
			URIRaw:  "http://[vFFF.fe80::a_en1]",
		},
		{
			Comment: "IPvFuture address (invalid version)",
			URIRaw:  "http://[vZ.fe80::a_en1]",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPvFuture address (invalid version)",
			URIRaw:  "http://[v]",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPvFuture address (empty address)",
			URIRaw:  "http://[vB.]",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPvFuture address (invalid characters)",
			URIRaw:  "http://[vAF.{}]",
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPvFuture address (invalid rune) (1)",
			URIRaw:  fmt.Sprintf("http://[v6.fe80::a_en1%s]", string([]rune{utf8.RuneError})),
			Err:     uri.ErrInvalidHostAddress,
		},
		{
			Comment: "IPvFuture address (invalid rune) (2)",
			URIRaw:  fmt.Sprintf("http://[v6%s.fe80::a_en1]", string([]rune{utf8.RuneError})),
			Err:     uri.ErrInvalidHostAddress,
		},
	}
}

func rawParsePortTests() []URITest {
	return []URITest{
		{
			Comment: "multiple ports",
			URIRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment",
			Err:     uri.ErrInvalidPort,
		},
		{
			Comment: "should detect an invalid port",
			URIRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment",
			Err:     uri.ErrInvalidPort,
		},
		{
			Comment: "path must begin with / or it collides with port",
			URIRaw:  "https://host:8080a?query=value#fragment",
			Err:     uri.ErrInvalidPort,
		},
	}
}

func rawParseQueryTests() []URITest {
	return []URITest{
		{
			Comment: "valid empty query after '?'",
			URIRaw:  "https://example-bin.org/path?",
		},
		{
			Comment: "valid query (separator character)",
			URIRaw:  "http://www.example.org/hello/world.txt/?id=5@part=three#there-you-go",
		},
		{
			Comment: "query contains invalid characters",
			URIRaw:  "http://httpbin.org/get?utf8=\xe2\x98\x83",
			Err:     uri.ErrInvalidQuery,
		},
		{
			Comment: "invalid query (invalid character) (1)",
			URIRaw:  "http://www.example.org/hello/world.txt/?id=5&pa{}rt=three#there-you-go",
			Err:     uri.ErrInvalidQuery,
		},
		{
			Comment: "invalid query (invalid character) (2)",
			URIRaw:  "http://www.example.org/hello/world.txt/?id=5&p|art=three#there-you-go",
			Err:     uri.ErrInvalidQuery,
		},
		{
			Comment: "invalid query (invalid character) (3)",
			URIRaw:  "http://httpbin.org/get?utf8=\xe2\x98\x83",
			Err:     uri.ErrInvalidQuery,
		},
		{
			Comment: "query is not correctly escaped.",
			URIRaw:  "http://www.contoso.com/path???/file name",
			Err:     uri.ErrInvalidQuery,
		},
		{
			Comment: "check percent encoding with query, incomplete escape sequence",
			URIRaw:  "https://user:passwd@ex%C3ample.com:8080/a?query=value%#fragment",
			Err:     uri.ErrInvalidQuery,
		},
	}
}

func rawParseFragmentTests() []URITest {
	return []URITest{
		{
			Comment: "empty fragment",
			URIRaw:  "mailto://u:p@host.domain.com#",
		},
		{
			Comment: "empty query and fragment",
			URIRaw:  "mailto://u:p@host.domain.com?#",
		},
		{
			Comment: "invalid char in fragment",
			URIRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go{}",
			Err:     uri.ErrInvalidFragment,
		},
		{
			Comment: "invalid fragment",
			URIRaw:  "http://example.w3.org/legit#ill[egal",
			Err:     uri.ErrInvalidFragment,
		},
		{
			Comment: "check percent encoding with fragment, incomplete escape sequence",
			URIRaw:  "https://user:passwd@ex%C3ample.com:8080/a?query=value#fragment%",
			Err:     uri.ErrInvalidFragment,
		},
	}
}

func rawParsePassTests() []URITest {
	// TODO: regroup themes, verify redundant testing
	return []URITest{
		{
			URIRaw: "foo://example.com:8042/over/there?name=ferret#nose",
		},
		{
			URIRaw: "http://httpbin.org/get?utf8=%e2%98%83",
		},
		{
			URIRaw: "mailto://user@domain.com",
		},
		{
			URIRaw: "ssh://user@git.openstack.org:29418/openstack/keystone.git",
		},
		{
			URIRaw: "https://willo.io/#yolo",
		},
		{
			Comment: "simple host and path",
			URIRaw:  "http://localhost/",
		},
		{
			Comment: "(redundant)",
			URIRaw:  "http://www.richardsonnen.com/",
		},
		// from https://github.com/python-hyper/rfc3986/blob/master/tests/test_validators.py
		{
			Comment: "complete authority",
			URIRaw:  "ssh://ssh@git.openstack.org:22/sigmavirus24",
		},
		{
			Comment: "(redundant)",
			URIRaw:  "https://git.openstack.org:443/sigmavirus24",
		},
		{
			Comment: "query + fragment",
			URIRaw:  "ssh://git.openstack.org:22/sigmavirus24?foo=bar#fragment",
		},
		{
			Comment: "(redundant)",
			URIRaw:  "git://github.com",
		},
		{
			Comment: "complete",
			URIRaw:  "https://user:passwd@http-bin.org:8080/a?query=value#fragment",
		},
		// from github.com/scalatra/rl: URI parser in scala
		{
			Comment: "port",
			URIRaw:  "http://www.example.org:8080",
		},
		{
			Comment: "(redundant)",
			URIRaw:  "http://www.example.org/",
		},
		{
			Comment: "UTF-8 host",
			URIRaw:  "http://www.詹姆斯.org/",
		},
		{
			Comment: "path",
			URIRaw:  "http://www.example.org/hello/world.txt",
		},
		{
			Comment: "query",
			URIRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three",
		},
		{
			Comment: "query+fragment",
			URIRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go",
		},
		{
			Comment: "fragment only",
			URIRaw:  "http://www.example.org/hello/world.txt/#here-we-are",
		},
		{
			Comment: "trailing empty fragment: legit",
			URIRaw:  "http://example.w3.org/legit#",
		},
		{
			Comment: "should detect a path starting with a /",
			URIRaw:  "file:///etc/hosts",
		},
		{
			Comment: `if a URI contains an authority component,
			then the path component must either be empty or begin with a slash ("/") character`,
			URIRaw: "https://host:8080?query=value#fragment",
		},
		{
			Comment: "path must begin with / (2)",
			URIRaw:  "https://host:8080/a?query=value#fragment",
		},
		{
			Comment: "double //, legit with escape",
			URIRaw:  "http+unix://%2Fvar%2Frun%2Fsocket/path?key=value",
		},
		{
			Comment: "double leading slash, legit context",
			URIRaw:  "http://host:8080//foo.html",
		},
		{
			URIRaw: "http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go",
		},
		{
			URIRaw: "http://www.example.org/hélloô/mötor/world.txt/?id=5&part=three#there-you-go",
		},
		{
			URIRaw: "http://www.example.org/hello/yzx;=1.1/world.txt/?id=5&part=three#there-you-go",
		},
		{
			URIRaw: "file://c:/directory/filename",
		},
		{
			URIRaw: "ldap://[2001:db8::7]/c=GB?objectClass?one",
		},
		{
			URIRaw: "ldap://[2001:db8::7]:8080/c=GB?objectClass?one",
		},
		{
			URIRaw: "http+unix:/%2Fvar%2Frun%2Fsocket/path?key=value",
		},
		{
			URIRaw: "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080/a?query=value#fragment",
		},
		{
			Comment: "should assert path and fragment",
			URIRaw:  "https://example-bin.org/path#frag?withQuestionMark",
		},
		{
			Comment: "should assert path and fragment (2)",
			URIRaw:  "mailto://u:p@host.domain.com?#ahahah",
		},
		{
			Comment: "should assert path and query",
			URIRaw:  "ldap://[2001:db8::7]/c=GB?objectClass?one",
		},
		{
			Comment: "should assert path and query",
			URIRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three",
		},
		{
			Comment: "should assert path and query",
			URIRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three?another#abc?efg",
		},
		{
			Comment: "should assert path and query",
			URIRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0]:8080/a?query=value#fragment",
		},
		{
			Comment: "should parse user/password, IPv6 percent-encoded host with zone",
			URIRaw:  "https://user:passwd@[::1%25lo]:8080/a?query=value#fragment",
		},
		// This is an invalid UTF8 sequence that SHOULD error, at least in the context of
		// Ref: https://url.spec.whatwg.org/#percent-encoded-bytes
		{
			Comment: "check percent encoding with DNS hostname, invalid escape sequence in host segment",
			URIRaw:  "https://user:passwd@ex%C3ample.com:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidDNSName,
		},
		{
			Comment: "check percent encoding with registered hostname, invalid escape sequence in host segment",
			URIRaw:  "tel://user:passwd@ex%C3ample.com:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidHost,
		},
		{
			Comment: "check percent encoding with registered hostname, incomplete escape sequence in host segment",
			URIRaw:  "https://user:passwd@ex%C3ample.com%:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidDNSName,
		},
	}
}

func rawParseUserInfoTests() []URITest {
	return []URITest{
		{
			Comment: "userinfo contains invalid character '{'",
			URIRaw:  "mailto://{}:{}@host.domain.com",
			Err:     uri.ErrInvalidUserInfo,
		},
		{
			Comment: "invalid user",
			URIRaw:  "https://user{}:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment",
			Err:     uri.ErrInvalidUserInfo,
		},
	}
}

func rawParseFailTests() []URITest {
	// other failures not already caught by the other test cases
	// (atm empty)
	return nil
}
