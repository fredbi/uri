package uri

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allGenerators = []testGenerator{
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

func rawParseReferenceTests() []uriTest {
	return []uriTest{
		{
			comment:     "valid missing scheme for an URI reference",
			uriRaw:      "//foo.bar/?baz=qux#quux",
			isReference: true,
		},
		{
			comment:     "valid URI reference (not a valid URI)",
			uriRaw:      "//host.domain.com/a/b",
			isReference: true,
		},
		{
			comment:     "valid URI reference with port (not a valid URI)",
			uriRaw:      "//host.domain.com:8080/a/b",
			isReference: true,
		},
		{
			comment:     "absolute reference with port",
			uriRaw:      "//host.domain.com:8080/a/b",
			isReference: true,
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "host.domain.com", u.Authority().Host())
				assert.Equal(t, "8080", u.Authority().Port())
				assert.Equal(t, "/a/b", u.Authority().Path())
			},
		},
		{
			comment:     "absolute reference with query params",
			uriRaw:      "//host.domain.com:8080?query=x/a/b",
			isReference: true,
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "host.domain.com", u.Authority().Host())
				assert.Equal(t, "8080", u.Authority().Port())
				assert.Equal(t, "", u.Authority().Path())
				assert.Equal(t, "x/a/b", u.Query().Get("query"))
			},
		},
		{
			comment:     "absolute reference with query params (reversed)",
			uriRaw:      "//host.domain.com:8080/a/b?query=x",
			isReference: true,
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "host.domain.com", u.Authority().Host())
				assert.Equal(t, "8080", u.Authority().Port())
				assert.Equal(t, "/a/b", u.Authority().Path())
				assert.Equal(t, "x", u.Query().Get("query"))
			},
		},
		{
			comment:     "invalid URI which is a valid reference",
			uriRaw:      "//not.a.user@not.a.host/just/a/path",
			isReference: true,
		},
		{
			comment:     "not an URI but a valid reference",
			uriRaw:      "/",
			isReference: true,
		},
		{
			comment:     "URL is an URI reference",
			uriRaw:      "//not.a.user@not.a.host/just/a/path",
			isReference: true,
		},
		{
			comment:     "URL is an URI reference, with escaped host",
			uriRaw:      "//not.a.user@%66%6f%6f.com/just/a/path/also",
			isReference: true,
		},
		{
			comment:     "non letter is an URI reference",
			uriRaw:      "*",
			isReference: true,
		},
		{
			comment:     "file name is an URI reference",
			uriRaw:      "foo.html",
			isReference: true,
		},
		{
			comment:     "directory is an URI reference",
			uriRaw:      "../dir/",
			isReference: true,
		},
		{
			comment:     "empty string is an URI reference",
			uriRaw:      "",
			isReference: true,
		},
	}
}

func rawParseStructureTests() []uriTest {
	return []uriTest{
		{
			comment: "// without // prefix, this is parsed as a path",
			uriRaw:  "mailto:user@domain.com",
			uri: &uri{
				scheme:   "mailto",
				hierPart: "user@domain.com",
				query:    "",
				fragment: "",
				authority: authorityInfo{
					path: "user@domain.com",
				},
			},
		},
		{
			comment: "with // prefix, this parsed as a user + host",
			uriRaw:  "mailto://user@domain.com",
			uri: &uri{
				scheme:   "mailto",
				hierPart: "//user@domain.com",
				query:    "",
				fragment: "",
				authority: authorityInfo{
					prefix:   "//",
					userinfo: "user",
					host:     "domain.com",
				},
			},
		},
		{
			comment: "pathological input (1)",
			uriRaw:  "?//x",
			err:     ErrInvalidURI,
		},
		{
			comment: "pathological input (2)",
			uriRaw:  "#//x",
			err:     ErrInvalidURI,
		},
		{
			comment: "pathological input (3)",
			uriRaw:  "://x",
			err:     ErrInvalidURI,
		},
		{
			comment: "pathological input (4)",
			uriRaw:  ".?:",
			err:     ErrInvalidURI,
		},
		{
			comment: "pathological input (5)",
			uriRaw:  ".#:",
			err:     ErrInvalidURI,
		},
		{
			comment: "pathological input (6)",
			uriRaw:  "?",
			err:     ErrInvalidURI,
		},
		{
			comment: "pathological input (7)",
			uriRaw:  "#",
			err:     ErrInvalidURI,
		},
		{
			comment: "pathological input (8)",
			uriRaw:  "?#",
			err:     ErrInvalidURI,
		},
		{
			comment: "invalid empty URI",
			uriRaw:  "",
			err:     ErrNoSchemeFound,
		},
		{
			comment: "invalid with blank",
			uriRaw:  " ",
			err:     ErrNoSchemeFound,
		},
		{
			comment: "no separator",
			uriRaw:  "foo",
			err:     ErrNoSchemeFound,
		},
		{
			comment: "no ':' separator",
			uriRaw:  "foo@bar",
			err:     ErrNoSchemeFound,
		},
	}
}

func rawParseSchemeTests() []uriTest {
	return []uriTest{
		{
			comment: "urn scheme",
			uriRaw:  "urn://example-bin.org/path",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "urn", u.Scheme())
			},
		},
		{
			comment: "only scheme (DNS host), valid!",
			uriRaw:  "http:",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "http", u.Scheme())
			},
		},
		{
			comment: "only scheme (registered name empty), valid!",
			uriRaw:  "foo:",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "foo", u.Scheme())
			},
		},
		{
			comment: "scheme without prefix (urn)",
			uriRaw:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "urn", u.Scheme())
				assert.Equal(t, "oasis:names:specification:docbook:dtd:xml:4.1.2", u.Authority().Path())
			},
		},
		{
			comment: "scheme without prefix (urn-like)",
			uriRaw:  "news:comp.infosystems.www.servers.unix",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "news", u.Scheme())
				assert.Equal(t, "comp.infosystems.www.servers.unix", u.Authority().Path())
			},
		},
		{
			comment: "+ and - in scheme (e.g. tel resource)",
			uriRaw:  "tel:+1-816-555-1212",
		},
		{
			comment: "should assert scheme",
			uriRaw:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "urn", u.Scheme())
			},
		},
		{
			comment: "legit separator in scheme",
			uriRaw:  "http+unix://%2Fvar%2Frun%2Fsocket/path?key=value",
		},
		{
			comment: "with scheme only",
			uriRaw:  "https:",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "https", u.Scheme())
			},
		},
		{
			comment: "empty scheme",
			uriRaw:  "://bob/",
			err:     ErrInvalidURI,
		},
		{
			comment: "invalid scheme (should start with a letter) (2)",
			uriRaw:  "?invalidscheme://www.example.com",
			err:     ErrInvalidURI,
		},
		{
			comment: "invalid scheme (invalid character) (2)",
			uriRaw:  "ht?tps:",
			err:     ErrInvalidURI,
		},
		{
			comment: "relative URIs with a colon (':') in their first segment are not considered well-formed",
			uriRaw:  "2013.05.29_14:33:41",
			err:     ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (should start with a letter) (1)",
			uriRaw:  "1http://bob",
			err:     ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (too short)",
			uriRaw:  "x://bob",
			err:     ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (invalid character) (1)",
			uriRaw:  "x{}y://bob",
			err:     ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (invalid character) (3)",
			uriRaw:  "inv;alidscheme://www.example.com",
			err:     ErrInvalidScheme,
		},
		{
			comment: "absolute URI that represents an implicit file URI.",
			uriRaw:  "c:\\directory\filename",
			err:     ErrInvalidScheme,
		},
		{
			comment: "represents a hierarchical absolute URI and does not contain '://'",
			uriRaw:  "www.contoso.com/path/file",
			err:     ErrNoSchemeFound,
		},
	}
}

func rawParsePathTests() []uriTest {
	return []uriTest{
		{
			comment: "legitimate use of several starting /'s in path'",
			uriRaw:  "file://hostname//etc/hosts",
			asserter: func(t testing.TB, u URI) {
				auth := u.Authority()
				require.Equal(t, "//etc/hosts", auth.Path())
				require.Equal(t, "hostname", auth.Host())
			},
		},
		{
			comment: "authority is not empty: valid path with double '/' (see issue#3)",
			uriRaw:  "http://host:8080//foo.html",
		},
		{
			comment: "path",
			uriRaw:  "https://example-bin.org/path?",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "/path", u.Authority().Path())
			},
		},
		{
			comment: "empty path, query and fragment",
			uriRaw:  "mailto://u:p@host.domain.com?#",
			asserter: func(t testing.TB, u URI) {
				assert.Empty(t, u.Authority().Path())
			},
		},
		{
			comment: "empty path",
			uriRaw:  "http://foo.com",
		},
		{
			comment: "path only, no query, no fragmeny",
			uriRaw:  "http://foo.com/path",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "/path", u.Authority().Path())
			},
		},
		{
			comment: "path with escaped spaces",
			uriRaw:  "http://example.w3.org/path%20with%20spaces.html",
			asserter: func(t testing.TB, u URI) {
				// path is stored unescaped
				assert.Equal(t, "/path%20with%20spaces.html", u.Authority().Path())
			},
		},
		{
			comment: "path is just an escape space",
			uriRaw:  "http://example.w3.org/%20",
		},
		{
			comment: "dots in path",
			uriRaw:  "ftp://ftp.is.co.za/../../../rfc/rfc1808.txt",
		},
		{
			comment: "= in path",
			uriRaw:  "ldap://[2001:db8::7]/c=GB?objectClass?one",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "/c=GB", u.Authority().Path())
			},
		},
		{
			comment: "path with drive letter (e.g. windows) (1)",
			// this one is dubious: Microsoft (.Net) recognizes the C:/... string as a path and
			// states this as incorrect uri -- all other validators state a host "c" and state this uri as a valid one
			uriRaw: "file://c:/directory/filename",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "c", u.Authority().Host())
				assert.Equal(t, "/directory/filename", u.Authority().Path())
			},
		},
		{
			comment: "path with drive letter (e.g. windows) (2)",
			// The unambiguous correct URI notation is  file:///c:/directory/filename
			uriRaw: "file:///c:/directory/filename",
			asserter: func(t testing.TB, u URI) {
				assert.Empty(t, u.Authority().Host())
				assert.Equal(t, "/c:/directory/filename", u.Authority().Path())
			},
		},
		{
			comment: `if a URI does not contain an authority component,
		then the path cannot begin with two slash characters ("//").`,
			uriRaw: "https:////a?query=value#fragment",
			err:    ErrInvalidPath,
		},
		{
			comment: "contains unescaped backslashes even if they will be treated as forward slashes",
			uriRaw:  "http:\\host/path/file",
			err:     ErrInvalidPath,
		},
		{
			comment: "invalid path (invalid characters)",
			uriRaw:  "http://www.example.org/hello/{}yzx;=1.1/world.txt/?id=5&part=three#there-you-go",
			err:     ErrInvalidPath,
		},
		{
			comment: "should detect a path starting with several /'s",
			uriRaw:  "file:////etc/hosts",
			err:     ErrInvalidPath,
		},
		{
			comment: "empty host  => double '/' invalid in this context",
			uriRaw:  "http:////foo.html",
			err:     ErrInvalidPath,
		},
		{
			comment: "trailing empty fragment, invalid path",
			uriRaw:  "http://example.w3.org/%legit#",
			err:     ErrInvalidPath,
		},
		{
			comment: "partial escape (1)",
			uriRaw:  "http://example.w3.org/%a",
			err:     ErrInvalidPath,
		},
		{
			comment: "partial escape (2)",
			uriRaw:  "http://example.w3.org/%a/foo",
			err:     ErrInvalidPath,
		},
		{
			comment: "partial escape (3)",
			uriRaw:  "http://example.w3.org/%illegal",
			err:     ErrInvalidPath,
		},
	}
}

func rawParseHostTests() []uriTest {
	return []uriTest{
		{
			comment: "authorized dash '-' in host",
			uriRaw:  "https://example-bin.org/path",
		},
		{
			comment: "host with many segments",
			uriRaw:  "ftp://ftp.is.co.za/rfc/rfc1808.txt",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "ftp.is.co.za", u.Authority().Host())
			},
		},
		{
			comment: "should error on empty host",
			uriRaw:  "https://user:passwd@:8080/a?query=value#fragment",
			err:     ErrMissingHost,
		},
		{
			comment: "unicode host",
			uriRaw:  "http://www.詹姆斯.org/",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "www.詹姆斯.org", u.Authority().Host())
			},
		},
		{
			comment: "illegal characters",
			uriRaw:  "http://<foo>",
			err:     ErrInvalidDNSName,
		},
		{
			comment: "should detect an invalid host (DNS rule) (1)",
			uriRaw:  "https://user:passwd@286;0.0.1:8080/a?query=value#fragment",
			err:     ErrInvalidHost,
		},
		{
			comment: "should detect an invalid host (DNS rule) (2)",
			uriRaw:  "https://user:passwd@256.256.256.256:8080/a?query=value#fragment",
			err:     ErrInvalidHost,
		},
		{
			comment: "registered name containing unallowed character",
			uriRaw:  "bob://x|y/",
			err:     ErrInvalidHost,
		},
		{
			comment: "invalid host (contains blank space)",
			uriRaw:  "http://www.exa mple.org",
			err:     ErrInvalidHost,
		},
		{
			comment: "DNS hostname is too long",
			uriRaw:  fmt.Sprintf("https://%s/", strings.Repeat("x", 256)),
			err:     ErrInvalidDNSName,
		},
		{
			comment: "DNS segment in hostname is empty",
			uriRaw:  "https://seg..com/",
			err:     ErrInvalidDNSName,
		},
		{
			comment: "DNS segment ends with unallowed character",
			uriRaw:  "https://x-.y.com/",
			err:     ErrInvalidDNSName,
		},
		{
			comment: "DNS segment in hostname too long",
			uriRaw:  fmt.Sprintf("https://%s.%s.com/", strings.Repeat("x", 63), strings.Repeat("y", 64)),
			err:     ErrInvalidDNSName,
		},
	}
}

func rawParseIPHostTests() []uriTest {
	return []uriTest{
		{
			comment: "IPv6 host",
			uriRaw:  "mailto://user@[fe80::1]",
			uri: &uri{
				scheme:   "mailto",
				hierPart: "//user@[fe80::1]",
				query:    "",
				fragment: "",
				authority: authorityInfo{
					prefix:   "//",
					userinfo: "user",
					host:     "fe80::1",
					isIPv6:   true,
				},
			},
		},
		{
			comment: "IPv6 host with zone",
			uriRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en1]:8080/a?query=value#fragment",
		},
		{
			comment: "ipv4 host",
			uriRaw:  "https://user:passwd@127.0.0.1:8080/a?query=value#fragment",
		},
		{
			comment: "IPv4 host",
			uriRaw:  "http://192.168.0.1/",
		},
		{
			comment: "IPv4 host with port",
			uriRaw:  "http://192.168.0.1:8080/",
		},
		{
			comment: "IPv6 host",
			uriRaw:  "http://[fe80::1]/",
		},
		{
			comment: "IPv6 host with port",
			uriRaw:  "http://[fe80::1]:8080/",
		},
		{
			comment: "IPv6 host with zone",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25lo]:8080/a?query=value#fragment",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25lo", u.Authority().Host())
			},
		},
		// Tests exercising RFC 6874 compliance:
		{
			comment: "IPv6 host with (escaped) zone identifier",
			uriRaw:  "http://[fe80::1%25en0]/",
		},
		{
			comment: "IPv6 host with zone identifier and port",
			uriRaw:  "http://[fe80::1%25en0]:8080/",
		},
		{
			comment: "IPv6 host with percent-encoded+unreserved zone identifier",
			uriRaw:  "http://[fe80::1%25%65%6e%301-._~]/",
		},
		{
			comment: "IPv6 host with percent-encoded+unreserved zone identifier",
			uriRaw:  "http://[fe80::1%25%65%6e%301-._~]:8080/"},
		// percent encoding in IP addresses
		{
			comment: "IP v4 host (escaped) %31 is percent-encoded for '1'",
			uriRaw:  "http://192.168.0.%31/",
		},
		{
			comment: "URL is an URI",
			uriRaw:  "http://192.168.0.%31:8080/",
		},
		{
			comment: "URL is an URI",
			uriRaw:  "http://[fe80::%31]/",
		},
		{
			comment: "URL is an URI",
			uriRaw:  "http://[fe80::%31]:8080/",
		},
		{
			comment: "URL is an URI",
			uriRaw:  "http://[fe80::%31%25en0]/",
		},
		{
			comment: "URL is an URI",
			uriRaw:  "http://[fe80::%31%25en0]:8080/",
		},
		{
			uriRaw: "https://user:passwd@127.0.0.1:8080/a?query=value#fragment",
		},
		{
			uriRaw: "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment",
		},
		{
			uriRaw: "https://user:passwd@[FF02:30:0:0:0:0:0:5%25lo]:8080/a?query=value#fragment",
		},
		{
			comment: "IPv6 with wrong percent encoding",
			uriRaw:  "http://[fe80::%%31]:8080/",
			err:     ErrInvalidHost,
		},
		// These two cases are valid as textual representations as
		// described in RFC 4007, but are not valid as address
		// literals with IPv6 zone identifiers in URIs as described in
		// RFC 6874.
		{
			comment: "invalid IPv4: part>255",
			uriRaw:  "https://user:passwd@256.256.256.256:8080/a?query=value#fragment",
			err:     ErrInvalidHost,
		},
		{
			comment: "invalid IPv6 (double empty ::)",
			uriRaw:  "https://user:passwd@[FF02::3::5]:8080/a?query=value#fragment",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "invalid IPv6 host with empty (escaped) zone",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25]:8080/a?query=value#fragment",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "invalid IPv6 with unescaped zone (bad percent-encoding)",
			uriRaw:  "https://user:passwd@[FADF:01%en0]:8080/a?query=value#fragment",
			err:     ErrInvalidHost,
		},
		{
			comment: "should not parse IPv6 host with empty zone (bad percent encoding)",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%]:8080/a?query=value#fragment",
			err:     ErrInvalidHost,
		},
		{
			comment: "empty IPv6",
			uriRaw:  "scheme://user:passwd@[]/valid",
			err:     ErrInvalidURI,
		},
		{
			comment: "zero IPv6",
			// this is valid according to
			uriRaw: "scheme://user:passwd@[::]/valid",
			//err:     ErrInvalidURI,
		},
		{
			comment: "invalid IPv6 (lack closing bracket) (1)",
			uriRaw:  "http://[fe80::1/",
			err:     ErrInvalidURI,
		},
		{
			comment: "invalid IPv6 (lack closing bracket) (2)",
			uriRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0:8080/a?query=value#fragment",
			err:     ErrInvalidURI,
		},
		{
			comment: "invalid IPv6 (lack closing bracket) (3)",
			uriRaw:  "https://user:passwd@[FADF:01%en0:8080/a?query=value#fragment",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (1)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (2)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080/?#",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (3)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080#",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (4)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080#abc",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 empty zone",
			uriRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25]:8080/a?query=value#fragment",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 addresses not between square brackets are invalid hosts (1)",
			uriRaw:  "https://0%3A0%3A0%3A0%3A0%3A0%3A0%3A1/a",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 addresses not between square brackets are invalid hosts (2)",
			uriRaw:  "https://FF02:30:0:0:0:0:0:5%25/a",
			err:     ErrInvalidPort, // ':30' parses as a port
		},
		{
			comment: "IP addresses between square brackets should not be ipv4 addresses",
			uriRaw:  "https://[192.169.224.1]/a",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "percent encoded host is valid, with encoded character not being valid",
			uriRaw:  "urn://user:passwd@ex%7Cample.com:8080/a?query=value#fragment",
		},
		{
			comment: "valid percent-encoded host (dash character is allowed in registered name)",
			uriRaw:  "urn://user:passwd@ex%2Dample.com:8080/a?query=value#fragment",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "ex%2Dample.com", u.Authority().Host())
			},
		},
		{
			comment: "check percent encoding with DNS hostname, dash allowed in DNS name",
			uriRaw:  "https://user:passwd@ex%2Dample.com:8080/a?query=value#fragment",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "ex%2Dample.com", u.Authority().Host())
			},
		},
		// Just for fun: IPvFuture...
		{
			comment: "IPvFuture address",
			uriRaw:  "http://[v6.fe80::a_en1]",
		},
		{
			comment: "IPvFuture address",
			uriRaw:  "http://[vFFF.fe80::a_en1]",
		},
		{
			comment: "IPvFuture address (invalid version)",
			uriRaw:  "http://[vZ.fe80::a_en1]",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "IPvFuture address (invalid version)",
			uriRaw:  "http://[v]",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "IPvFuture address (empty address)",
			uriRaw:  "http://[vB.]",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "IPvFuture address (invalid characters)",
			uriRaw:  "http://[vAF.{}]",
			err:     ErrInvalidHostAddress,
		},
	}
}

func rawParsePortTests() []uriTest {
	return []uriTest{
		{
			comment: "multiple ports",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment",
			err:     ErrInvalidPort,
		},
		{
			comment: "should detect an invalid port",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment",
			err:     ErrInvalidPort,
		},
		{
			comment: "path must begin with / or it collides with port",
			uriRaw:  "https://host:8080a?query=value#fragment",
			err:     ErrInvalidPort,
		},
	}
}

func rawParseQueryTests() []uriTest {
	return []uriTest{
		{
			comment: "valid empty query after '?'",
			uriRaw:  "https://example-bin.org/path?",
		},
		{
			comment: "valid query (separator character)",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5@part=three#there-you-go",
		},
		{
			comment: "query contains invalid characters",
			uriRaw:  "http://httpbin.org/get?utf8=\xe2\x98\x83",
			err:     ErrInvalidQuery,
		},
		{
			comment: "invalid query (invalid character) (1)",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&pa{}rt=three#there-you-go",
			err:     ErrInvalidQuery,
		},
		{
			comment: "invalid query (invalid character) (2)",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&p|art=three#there-you-go",
			err:     ErrInvalidQuery,
		},
		{
			comment: "invalid query (invalid character) (3)",
			uriRaw:  "http://httpbin.org/get?utf8=\xe2\x98\x83",
			err:     ErrInvalidQuery,
		},
		{
			comment: "query is not correctly escaped.",
			uriRaw:  "http://www.contoso.com/path???/file name",
			err:     ErrInvalidQuery,
		},
	}
}

func rawParseFragmentTests() []uriTest {
	return []uriTest{
		{
			comment: "empty fragment",
			uriRaw:  "mailto://u:p@host.domain.com#",
		},
		{
			comment: "empty query and fragment",
			uriRaw:  "mailto://u:p@host.domain.com?#",
		},
		{
			comment: "invalid char in fragment",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go{}",
			err:     ErrInvalidFragment,
		},
		{
			comment: "invalid fragment",
			uriRaw:  "http://example.w3.org/legit#ill[egal",
			err:     ErrInvalidFragment,
		},
	}
}

func rawParsePassTests() []uriTest {
	// TODO: regroup themes, verify redundant testing
	return []uriTest{
		{
			uriRaw: "foo://example.com:8042/over/there?name=ferret#nose",
			uri: &uri{
				scheme:   "foo",
				hierPart: "//example.com:8042/over/there",
				query:    "name=ferret",
				fragment: "nose",
				authority: authorityInfo{
					prefix:   "//",
					userinfo: "",
					host:     "example.com",
					port:     "8042",
					path:     "/over/there",
					isIPv6:   false,
				},
			},
		},
		{
			uriRaw: "http://httpbin.org/get?utf8=%e2%98%83",
			uri: &uri{
				scheme:   "http",
				hierPart: "//httpbin.org/get",
				query:    "utf8=%e2%98%83",
				fragment: "",
				authority: authorityInfo{
					prefix:   "//",
					userinfo: "",
					host:     "httpbin.org",
					port:     "",
					path:     "/get",
					isIPv6:   false,
				},
			},
		},
		{
			uriRaw: "mailto://user@domain.com",
			uri: &uri{
				scheme:   "mailto",
				hierPart: "//user@domain.com",
				query:    "",
				fragment: "",
				authority: authorityInfo{
					prefix:   "//",
					userinfo: "user",
					host:     "domain.com",
					port:     "",
					path:     "",
					isIPv6:   false,
				},
			},
		},
		{
			uriRaw: "ssh://user@git.openstack.org:29418/openstack/keystone.git",
			uri: &uri{
				scheme:   "ssh",
				hierPart: "//user@git.openstack.org:29418/openstack/keystone.git",
				query:    "",
				fragment: "",
				authority: authorityInfo{
					prefix:   "//",
					userinfo: "user",
					host:     "git.openstack.org",
					port:     "29418",
					path:     "/openstack/keystone.git",
					isIPv6:   false,
				},
			},
		},
		{
			uriRaw: "https://willo.io/#yolo",
			uri: &uri{
				scheme:   "https",
				hierPart: "//willo.io/",
				query:    "",
				fragment: "yolo",
				authority: authorityInfo{
					prefix:   "//",
					userinfo: "",
					host:     "willo.io",
					port:     "",
					path:     "/",
					isIPv6:   false,
				},
			},
		},
		{
			comment: "simple host and path",
			uriRaw:  "http://localhost/",
		},
		{
			comment: "(redundant)",
			uriRaw:  "http://www.richardsonnen.com/",
		},
		// from https://github.com/python-hyper/rfc3986/blob/master/tests/test_validators.py
		{
			comment: "complete authority",
			uriRaw:  "ssh://ssh@git.openstack.org:22/sigmavirus24",
		},
		{
			comment: "(redundant)",
			uriRaw:  "https://git.openstack.org:443/sigmavirus24",
		},
		{
			comment: "query + fragment",
			uriRaw:  "ssh://git.openstack.org:22/sigmavirus24?foo=bar#fragment",
		},
		{
			comment: "(redundant)",
			uriRaw:  "git://github.com",
		},
		{
			comment: "complete",
			uriRaw:  "https://user:passwd@http-bin.org:8080/a?query=value#fragment",
		},
		// from github.com/scalatra/rl: URI parser in scala
		{
			comment: "port",
			uriRaw:  "http://www.example.org:8080",
		},
		{
			comment: "(redundant)",
			uriRaw:  "http://www.example.org/",
		},
		{
			comment: "UTF-8 host",
			uriRaw:  "http://www.詹姆斯.org/",
		},
		{
			comment: "path",
			uriRaw:  "http://www.example.org/hello/world.txt",
		},
		{
			comment: "query",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three",
		},
		{
			comment: "query+fragment",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go",
		},
		{
			comment: "fragment only",
			uriRaw:  "http://www.example.org/hello/world.txt/#here-we-are",
		},
		{
			comment: "trailing empty fragment: legit",
			uriRaw:  "http://example.w3.org/legit#",
		},
		{
			comment: "should detect a path starting with a /",
			uriRaw:  "file:///etc/hosts",
			asserter: func(t testing.TB, u URI) {
				auth := u.Authority()
				require.Equal(t, "/etc/hosts", auth.Path())
				require.Empty(t, auth.Host())
			},
		},
		{
			comment: `if a URI contains an authority component,
			then the path component must either be empty or begin with a slash ("/") character`,
			uriRaw: "https://host:8080?query=value#fragment",
		},
		{
			comment: "path must begin with / (2)",
			uriRaw:  "https://host:8080/a?query=value#fragment",
		},
		{
			comment: "double //, legit with escape",
			uriRaw:  "http+unix://%2Fvar%2Frun%2Fsocket/path?key=value",
		},
		{
			comment: "double leading slash, legit context",
			uriRaw:  "http://host:8080//foo.html",
		},
		{
			uriRaw: "http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go",
		},
		{
			uriRaw: "http://www.example.org/hélloô/mötor/world.txt/?id=5&part=three#there-you-go",
		},
		{
			uriRaw: "http://www.example.org/hello/yzx;=1.1/world.txt/?id=5&part=three#there-you-go",
		},
		{
			uriRaw: "file://c:/directory/filename",
		},
		{
			uriRaw: "ldap://[2001:db8::7]/c=GB?objectClass?one",
		},
		{
			uriRaw: "ldap://[2001:db8::7]:8080/c=GB?objectClass?one",
		},
		{
			uriRaw: "http+unix:/%2Fvar%2Frun%2Fsocket/path?key=value",
		},
		{
			uriRaw: "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080/a?query=value#fragment",
		},
		{
			comment: "should assert path and fragment",
			uriRaw:  "https://example-bin.org/path#frag?withQuestionMark",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "/path", u.Authority().Path())
				assert.Empty(t, u.Query())
				assert.Equal(t, "frag?withQuestionMark", u.Fragment())
			},
		},
		{
			comment: "should assert path and fragment (2)",
			uriRaw:  "mailto://u:p@host.domain.com?#ahahah",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "", u.Authority().Path())
				assert.Empty(t, u.Query())
				assert.Equal(t, "ahahah", u.Fragment())
			},
		},
		{
			comment: "should assert path and query",
			uriRaw:  "ldap://[2001:db8::7]/c=GB?objectClass?one",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "/c=GB", u.Authority().Path())
				nuri := u.(*uri)
				assert.Equal(t, "objectClass?one", nuri.query) // TODO(fred): use Query() and url.Values
				assert.Equal(t, "", u.Fragment())
			},
		},
		{
			comment: "should assert path and query",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "/hello/world.txt/", u.Authority().Path())
				nuri := u.(*uri)
				assert.Equal(t, "id=5&part=three", nuri.query)
				assert.Equal(t, "", u.Fragment())
			},
		},
		{
			comment: "should assert path and query",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three?another#abc?efg",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "/hello/world.txt/", u.Authority().Path())
				nuri := u.(*uri)
				assert.Equal(t, "id=5&part=three?another", nuri.query)
				assert.Equal(t, "abc?efg", u.Fragment())
				assert.Equal(t, url.Values{"id": []string{"5"}, "part": []string{"three?another"}}, u.Query())
			},
		},
		{
			comment: "should assert path and query",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0]:8080/a?query=value#fragment",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0", u.Authority().Host())
				assert.Equal(t, "//user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0]:8080/a", u.Authority().String())
				assert.Equal(t, "https", u.Scheme())
				assert.Equal(t, url.Values{"query": []string{"value"}}, u.Query())
			},
		},
		{
			comment: "should parse user/password, IPv6 percent-encoded host with zone",
			uriRaw:  "https://user:passwd@[::1%25lo]:8080/a?query=value#fragment",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "https", u.Scheme())
				assert.Equal(t, "8080", u.Authority().Port())
				assert.Equal(t, "user:passwd", u.Authority().UserInfo())
			},
		},
	}
}

func rawParseUserInfoTests() []uriTest {
	return []uriTest{
		{
			comment: "userinfo contains invalid character '{'",
			uriRaw:  "mailto://{}:{}@host.domain.com",
			err:     ErrInvalidUserInfo,
		},
		{
			comment: "invalid user",
			uriRaw:  "https://user{}:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment",
			err:     ErrInvalidUserInfo,
		},
	}
}

func rawParseFailTests() []uriTest {
	// other failures not already caught by the other test cases
	// (atm empty)
	return nil
}
