package profiling

import (
	"fmt"
	"iter"
	"net/url"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/fredbi/uri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	uriTest struct {
		uriRaw      string
		uri         uri.URI //nolint:unused // reserved for future use
		err         error
		comment     string
		isReference bool
		// isNotURI    bool
		asserter func(testing.TB, uri.URI)
	}

	testGenerator = iter.Seq[uriTest]
)

func allGenerators() iter.Seq[testGenerator] {
	return slices.Values([]testGenerator{
		rawParsePassTests(),
		rawParseFailTests(),
		rawParseReferenceTests(),
		rawParseStructureTests(),
		rawParseSchemeTests(),
		rawParseUserInfoTests(),
		rawParsePathTests(),
		rawParseHostTests(),
		rawParseIPHostTests(),
		rawParsePortTests(),
		rawParseQueryTests(),
		rawParseFragmentTests(),
	})
}

func rawParseReferenceTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
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
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "host.domain.com", u.Authority().Host())
				assert.Equal(tb, "8080", u.Authority().Port())
				assert.Equal(tb, "/a/b", u.Authority().Path())
			},
		},
		{
			comment:     "absolute reference with query params",
			uriRaw:      "//host.domain.com:8080?query=x/a/b",
			isReference: true,
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "host.domain.com", u.Authority().Host())
				assert.Equal(tb, "8080", u.Authority().Port())
				assert.Empty(tb, u.Authority().Path())
				assert.Equal(tb, "x/a/b", u.Query().Get("query"))
			},
		},
		{
			comment:     "absolute reference with query params (reversed)",
			uriRaw:      "//host.domain.com:8080/a/b?query=x",
			isReference: true,
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "host.domain.com", u.Authority().Host())
				assert.Equal(tb, "8080", u.Authority().Port())
				assert.Equal(tb, "/a/b", u.Authority().Path())
				assert.Equal(tb, "x", u.Query().Get("query"))
			},
		},
		{
			comment:     "invalid uri.URI which is a valid reference",
			uriRaw:      "//not.a.user@not.a.host/just/a/path",
			isReference: true,
		},
		{
			comment:     "not an uri.URI but a valid reference",
			uriRaw:      "/",
			isReference: true,
		},
		{
			comment:     "URL is an uri.URI reference",
			uriRaw:      "//not.a.user@not.a.host/just/a/path",
			isReference: true,
		},
		{
			comment:     "URL is an uri.URI reference, with escaped host",
			uriRaw:      "//not.a.user@%66%6f%6f.com/just/a/path/also",
			isReference: true,
		},
		{
			comment:     "non letter is an uri.URI reference",
			uriRaw:      "*",
			isReference: true,
		},
		{
			comment:     "file name is an uri.URI reference",
			uriRaw:      "foo.html",
			isReference: true,
		},
		{
			comment:     "directory is an uri.URI reference",
			uriRaw:      "../dir/",
			isReference: true,
		},
		{
			comment:     "empty string is an uri.URI reference",
			uriRaw:      "",
			isReference: true,
		},
	})
}

func rawParseStructureTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
		{
			comment: "// without // prefix, this is parsed as a path",
			uriRaw:  "mailto:user@domain.com",
		},
		{
			comment: "with // prefix, this parsed as a user + host",
			uriRaw:  "mailto://user@domain.com",
		},
		{
			comment: "pathological input (1)",
			uriRaw:  "?//x",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "pathological input (2)",
			uriRaw:  "#//x",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "pathological input (3)",
			uriRaw:  "://x",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "pathological input (4)",
			uriRaw:  ".?:",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "pathological input (5)",
			uriRaw:  ".#:",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "pathological input (6)",
			uriRaw:  "?",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "pathological input (7)",
			uriRaw:  "#",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "pathological input (8)",
			uriRaw:  "?#",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "invalid empty uri.URI",
			uriRaw:  "",
			err:     uri.ErrNoSchemeFound,
		},
		{
			comment: "invalid with blank",
			uriRaw:  " ",
			err:     uri.ErrNoSchemeFound,
		},
		{
			comment: "no separator",
			uriRaw:  "foo",
			err:     uri.ErrNoSchemeFound,
		},
		{
			comment: "no ':' separator",
			uriRaw:  "foo@bar",
			err:     uri.ErrNoSchemeFound,
		},
	})
}

func rawParseSchemeTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
		{
			comment: "urn scheme",
			uriRaw:  "urn://example-bin.org/path",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "urn", u.Scheme())
			},
		},
		{
			comment: "only scheme (DNS host), valid!",
			uriRaw:  "http:",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "http", u.Scheme())
			},
		},
		{
			comment: "only scheme (registered name empty), valid!",
			uriRaw:  "foo:",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "foo", u.Scheme())
			},
		},
		{
			comment: "scheme without prefix (urn)",
			uriRaw:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "urn", u.Scheme())
				assert.Equal(tb, "oasis:names:specification:docbook:dtd:xml:4.1.2", u.Authority().Path())
			},
		},
		{
			comment: "scheme without prefix (urn-like)",
			uriRaw:  "news:comp.infosystems.www.servers.unix",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "news", u.Scheme())
				assert.Equal(tb, "comp.infosystems.www.servers.unix", u.Authority().Path())
			},
		},
		{
			comment: "+ and - in scheme (e.g. tel resource)",
			uriRaw:  "tel:+1-816-555-1212",
		},
		{
			comment: "should assert scheme",
			uriRaw:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "urn", u.Scheme())
			},
		},
		{
			comment: "legit separator in scheme",
			uriRaw:  "http+unix://%2Fvar%2Frun%2Fsocket/path?key=value",
		},
		{
			comment: "with scheme only",
			uriRaw:  "https:",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "https", u.Scheme())
			},
		},
		{
			comment: "empty scheme",
			uriRaw:  "://bob/",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "invalid scheme (should start with a letter) (2)",
			uriRaw:  "?invalidscheme://www.example.com",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "invalid scheme (invalid character) (2)",
			uriRaw:  "ht?tps:",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "relative uri.URIs with a colon (':') in their first segment are not considered well-formed",
			uriRaw:  "2013.05.29_14:33:41",
			err:     uri.ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (should start with a letter) (1)",
			uriRaw:  "1http://bob",
			err:     uri.ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (too short)",
			uriRaw:  "x://bob",
			err:     uri.ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (invalid character) (1)",
			uriRaw:  "x{}y://bob",
			err:     uri.ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (invalid character) (3)",
			uriRaw:  "inv;alidscheme://www.example.com",
			err:     uri.ErrInvalidScheme,
		},
		{
			comment: "absolute uri.URI that represents an implicit file URI.",
			uriRaw:  "c:\\directory\filename",
			err:     uri.ErrInvalidScheme,
		},
		{
			comment: "represents a hierarchical absolute uri.URI and does not contain '://'",
			uriRaw:  "www.contoso.com/path/file",
			err:     uri.ErrNoSchemeFound,
		},
	})
}

func rawParsePathTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
		{
			comment: "legitimate use of several starting /'s in path'",
			uriRaw:  "file://hostname//etc/hosts",
			asserter: func(tb testing.TB, u uri.URI) {
				auth := u.Authority()
				require.Equal(tb, "//etc/hosts", auth.Path())
				require.Equal(tb, "hostname", auth.Host())
			},
		},
		{
			comment: "authority is not empty: valid path with double '/' (see issue#3)",
			uriRaw:  "http://host:8080//foo.html",
		},
		{
			comment: "path",
			uriRaw:  "https://example-bin.org/path?",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "/path", u.Authority().Path())
			},
		},
		{
			comment: "empty path, query and fragment",
			uriRaw:  "mailto://u:p@host.domain.com?#",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Empty(tb, u.Authority().Path())
			},
		},
		{
			comment: "empty path",
			uriRaw:  "http://foo.com",
		},
		{
			comment: "path only, no query, no fragmeny",
			uriRaw:  "http://foo.com/path",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "/path", u.Authority().Path())
			},
		},
		{
			comment: "path with escaped spaces",
			uriRaw:  "http://example.w3.org/path%20with%20spaces.html",
			asserter: func(tb testing.TB, u uri.URI) {
				// path is stored unescaped
				assert.Equal(tb, "/path%20with%20spaces.html", u.Authority().Path())
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
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "/c=GB", u.Authority().Path())
			},
		},
		{
			comment: "path with drive letter (e.g. windows) (1)",
			// this one is dubious: Microsoft (.Net) recognizes the C:/... string as a path and
			// states this as incorrect uri -- all other validators state a host "c" and state this uri as a valid one
			uriRaw: "file://c:/directory/filename",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "c", u.Authority().Host())
				assert.Equal(tb, "/directory/filename", u.Authority().Path())
			},
		},
		{
			comment: "path with drive letter (e.g. windows) (2)",
			// The unambiguous correct uri.URI notation is  file:///c:/directory/filename
			uriRaw: "file:///c:/directory/filename",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Empty(tb, u.Authority().Host())
				assert.Equal(tb, "/c:/directory/filename", u.Authority().Path())
			},
		},
		{
			comment: `if a uri.URI does not contain an authority component,
		then the path cannot begin with two slash characters ("//").`,
			uriRaw: "https:////a?query=value#fragment",
			err:    uri.ErrInvalidPath,
		},
		{
			comment: "contains unescaped backslashes even if they will be treated as forward slashes",
			uriRaw:  "http:\\host/path/file",
			err:     uri.ErrInvalidPath,
		},
		{
			comment: "invalid path (invalid characters)",
			uriRaw:  "http://www.example.org/hello/{}yzx;=1.1/world.txt/?id=5&part=three#there-you-go",
			err:     uri.ErrInvalidPath,
		},
		{
			comment: "should detect a path starting with several /'s",
			uriRaw:  "file:////etc/hosts",
			err:     uri.ErrInvalidPath,
		},
		{
			comment: "empty host  => double '/' invalid in this context",
			uriRaw:  "http:////foo.html",
			err:     uri.ErrInvalidPath,
		},
		{
			comment: "trailing empty fragment, invalid path",
			uriRaw:  "http://example.w3.org/%legit#",
			err:     uri.ErrInvalidPath,
		},
		{
			comment: "partial escape (1)",
			uriRaw:  "http://example.w3.org/%a",
			err:     uri.ErrInvalidPath,
		},
		{
			comment: "partial escape (2)",
			uriRaw:  "http://example.w3.org/%a/foo",
			err:     uri.ErrInvalidPath,
		},
		{
			comment: "partial escape (3)",
			uriRaw:  "http://example.w3.org/%illegal",
			err:     uri.ErrInvalidPath,
		},
	})
}

func rawParseHostTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
		{
			comment: "authorized dash '-' in host",
			uriRaw:  "https://example-bin.org/path",
		},
		{
			comment: "host with many segments",
			uriRaw:  "ftp://ftp.is.co.za/rfc/rfc1808.txt",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "ftp.is.co.za", u.Authority().Host())
			},
		},
		{
			comment: "percent encoded host is valid, with encoded character not being valid",
			uriRaw:  "urn://user:passwd@ex%7Cample.com:8080/a?query=value#fragment",
		},
		{
			comment: "valid percent-encoded host (dash character is allowed in registered name)",
			uriRaw:  "urn://user:passwd@ex%2Dample.com:8080/a?query=value#fragment",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "ex%2Dample.com", u.Authority().Host())
			},
		},
		{
			comment: "check percent encoding with DNS hostname, dash allowed in DNS name",
			uriRaw:  "https://user:passwd@ex%2Dample.com:8080/a?query=value#fragment",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "ex%2Dample.com", u.Authority().Host())
			},
		},
		{
			comment: "should error on empty host",
			uriRaw:  "https://user:passwd@:8080/a?query=value#fragment",
			err:     uri.ErrMissingHost,
		},
		{
			comment: "unicode host",
			uriRaw:  "http://www.詹姆斯.org/", //nolint:gosmopolitan // legitimate test case for IRI (Internationalized Resource Identifier) support
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "www.詹姆斯.org", u.Authority().Host()) //nolint:gosmopolitan // legitimate test case for IRI (Internationalized Resource Identifier) support
			},
		},
		{
			comment: "illegal characters",
			uriRaw:  "http://<foo>",
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "should detect an invalid host (DNS rule) (1)",
			uriRaw:  "https://user:passwd@286;0.0.1:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "should detect an invalid host (DNS rule) (2)",
			uriRaw:  "https://user:passwd@256.256.256.256:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "registered name containing unallowed character",
			uriRaw:  "bob://x|y/",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "invalid host (contains blank space)",
			uriRaw:  "http://www.exa mple.org",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "DNS hostname is too long",
			uriRaw:  fmt.Sprintf("https://%s/", strings.Repeat("x", 256)),
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "DNS segment in hostname is empty",
			uriRaw:  "https://seg..com/",
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "DNS last segment in hostname is empty",
			uriRaw:  "https://seg.empty.com./",
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "DNS segment ends with unallowed character",
			uriRaw:  "https://x-.y.com/",
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "DNS segment in hostname too long",
			uriRaw:  fmt.Sprintf("https://%s.%s.com/", strings.Repeat("x", 63), strings.Repeat("y", 64)),
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "DNS with all segments empty",
			uriRaw:  "https://........./",
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "DNS segment ends with incomplete escape sequence",
			uriRaw:  "https://x.y.com%/",
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "DNS segment contains an invalid rune",
			uriRaw:  fmt.Sprintf("https://x.y.com%s/", string([]rune{utf8.RuneError})),
			err:     uri.ErrInvalidDNSName,
		},
	})
}

func rawParseIPHostTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
		{
			comment: "IPv6 host",
			uriRaw:  "mailto://user@[fe80::1]",
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
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25lo", u.Authority().Host())
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
			uriRaw:  "http://[fe80::1%25%65%6e%301-._~]:8080/",
		},
		{
			comment: "IPv6 host with invalid percent-encoding in zone identifier",
			uriRaw:  "http://[fe80::1%25%C3~]:8080/",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 host with invalid percent-encoding in zone identifier",
			uriRaw:  "http://[fe80::1%25%F3~]:8080/",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IP v4 host (escaped) %31 is percent-encoded for '1'",
			uriRaw:  "http://192.168.0.%31/",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "IPv4 address with percent-encoding is not allowed",
			uriRaw:  "http://192.168.0.%31:8080/",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "invalid IPv4 with port (2)",
			uriRaw:  "https://user:passwd@127.256.0.1:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "invalid IPv4 with port (3)",
			uriRaw:  "https://user:passwd@127.0127.0.1:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "valid IPv4 with port (1)",
			uriRaw:  "https://user:passwd@127.0.0.1:8080/a?query=value#fragment",
		},
		{
			comment: "invalid IPv4: part>255",
			uriRaw:  "https://user:passwd@256.256.256.256:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "IPv6 percent-encoding is limited to ZoneID specification, mus be %25",
			uriRaw:  "http://[fe80::%31]/",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 percent-encoding is limited to ZoneID specification, mus be %25 (2))",
			uriRaw:  "http://[fe80::%31]:8080/",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 percent-encoding is limited to ZoneID specification, mus be %25 (2))",
			uriRaw:  "http://[fe80::%31%25en0]/",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 percent-encoding is limited to ZoneID specification, mus be %25 (2))",
			uriRaw:  "http://[%310:fe80::%25en0]/",
			err:     uri.ErrInvalidHostAddress,
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
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 with wrong percent encoding",
			uriRaw:  "http://[fe80::%26lo]:8080/",
			err:     uri.ErrInvalidHostAddress,
		},
		// These two cases are valid as textual representations as
		// described in RFC 4007, but are not valid as address
		// literals with IPv6 zone identifiers in uri.URIs as described in
		// RFC 6874.
		{
			comment: "invalid IPv6 (double empty ::)",
			uriRaw:  "https://user:passwd@[FF02::3::5]:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "invalid IPv6 host with empty (escaped) zone",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25]:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "invalid IPv6 with unescaped zone (bad percent-encoding)",
			uriRaw:  "https://user:passwd@[FADF:01%en0]:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "should not parse IPv6 host with empty zone (bad percent encoding)",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%]:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "empty IPv6",
			uriRaw:  "scheme://user:passwd@[]/valid",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "zero IPv6",
			uriRaw:  "scheme://user:passwd@[::]/valid",
		},
		{
			comment: "invalid IPv6 (lack closing bracket) (1)",
			uriRaw:  "http://[fe80::1/",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "invalid IPv6 (lack closing bracket) (2)",
			uriRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0:8080/a?query=value#fragment",
			err:     uri.ErrInvalidURI,
		},
		{
			comment: "invalid IPv6 (lack closing bracket) (3)",
			uriRaw:  "https://user:passwd@[FADF:01%en0:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (1)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (2)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080/?#",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (3)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080#",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (4)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080#abc",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 empty zone",
			uriRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25]:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 unescaped zone with reserved characters",
			uriRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25:lo]:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPv6 addresses not between square brackets are invalid hosts (1)",
			uriRaw:  "https://0%3A0%3A0%3A0%3A0%3A0%3A0%3A1/a",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "IPv6 addresses not between square brackets are invalid hosts (2)",
			uriRaw:  "https://FF02:30:0:0:0:0:0:5%25/a",
			err:     uri.ErrInvalidPort, // ':30' parses as a port
		},
		{
			comment: "IP addresses between square brackets should not be ipv4 addresses",
			uriRaw:  "https://[192.169.224.1]/a",
			err:     uri.ErrInvalidHostAddress,
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
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPvFuture address (invalid version)",
			uriRaw:  "http://[v]",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPvFuture address (empty address)",
			uriRaw:  "http://[vB.]",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPvFuture address (invalid characters)",
			uriRaw:  "http://[vAF.{}]",
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPvFuture address (invalid rune) (1)",
			uriRaw:  fmt.Sprintf("http://[v6.fe80::a_en1%s]", string([]rune{utf8.RuneError})),
			err:     uri.ErrInvalidHostAddress,
		},
		{
			comment: "IPvFuture address (invalid rune) (2)",
			uriRaw:  fmt.Sprintf("http://[v6%s.fe80::a_en1]", string([]rune{utf8.RuneError})),
			err:     uri.ErrInvalidHostAddress,
		},
	})
}

func rawParsePortTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
		{
			comment: "multiple ports",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment",
			err:     uri.ErrInvalidPort,
		},
		{
			comment: "should detect an invalid port",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment",
			err:     uri.ErrInvalidPort,
		},
		{
			comment: "path must begin with / or it collides with port",
			uriRaw:  "https://host:8080a?query=value#fragment",
			err:     uri.ErrInvalidPort,
		},
	})
}

func rawParseQueryTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
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
			err:     uri.ErrInvalidQuery,
		},
		{
			comment: "invalid query (invalid character) (1)",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&pa{}rt=three#there-you-go",
			err:     uri.ErrInvalidQuery,
		},
		{
			comment: "invalid query (invalid character) (2)",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&p|art=three#there-you-go",
			err:     uri.ErrInvalidQuery,
		},
		{
			comment: "invalid query (invalid character) (3)",
			uriRaw:  "http://httpbin.org/get?utf8=\xe2\x98\x83",
			err:     uri.ErrInvalidQuery,
		},
		{
			comment: "query is not correctly escaped.",
			uriRaw:  "http://www.contoso.com/path???/file name",
			err:     uri.ErrInvalidQuery,
		},
		{
			comment: "check percent encoding with query, incomplete escape sequence",
			uriRaw:  "https://user:passwd@ex%C3ample.com:8080/a?query=value%#fragment",
			err:     uri.ErrInvalidQuery,
		},
	})
}

func rawParseFragmentTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
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
			err:     uri.ErrInvalidFragment,
		},
		{
			comment: "invalid fragment",
			uriRaw:  "http://example.w3.org/legit#ill[egal",
			err:     uri.ErrInvalidFragment,
		},
		{
			comment: "check percent encoding with fragment, incomplete escape sequence",
			uriRaw:  "https://user:passwd@ex%C3ample.com:8080/a?query=value#fragment%",
			err:     uri.ErrInvalidFragment,
		},
	})
}

func rawParsePassTests() iter.Seq[uriTest] {
	// TODO: regroup themes, verify redundant testing
	return slices.Values([]uriTest{
		{
			uriRaw: "foo://example.com:8042/over/there?name=ferret#nose",
		},
		{
			uriRaw: "http://httpbin.org/get?utf8=%e2%98%83",
		},
		{
			uriRaw: "mailto://user@domain.com",
		},
		{
			uriRaw: "ssh://user@git.openstack.org:29418/openstack/keystone.git",
		},
		{
			uriRaw: "https://willo.io/#yolo",
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
		// from github.com/scalatra/rl: uri.URI parser in scala
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
			uriRaw:  "http://www.詹姆斯.org/", //nolint:gosmopolitan // legitimate test case for IRI (Internationalized Resource Identifier) support
		},
		{
			comment: "Host with number at the start",
			uriRaw:  "http://www.4chan.com/",
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
			asserter: func(tb testing.TB, u uri.URI) {
				auth := u.Authority()
				require.Equal(tb, "/etc/hosts", auth.Path())
				require.Empty(tb, auth.Host())
			},
		},
		{
			comment: `if a uri.URI contains an authority component,
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
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "/path", u.Authority().Path())
				assert.Empty(tb, u.Query())
				assert.Equal(tb, "frag?withQuestionMark", u.Fragment())
			},
		},
		{
			comment: "should assert path and fragment (2)",
			uriRaw:  "mailto://u:p@host.domain.com?#ahahah",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Empty(tb, u.Authority().Path())
				assert.Empty(tb, u.Query())
				assert.Equal(tb, "ahahah", u.Fragment())
			},
		},
		{
			comment: "should assert path and query",
			uriRaw:  "ldap://[2001:db8::7]/c=GB?objectClass?one",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "/c=GB", u.Authority().Path())
				assert.Equal(tb, "objectClass?one", u.Query())
				assert.Empty(tb, u.Fragment())
			},
		},
		{
			comment: "should assert path and query",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "/hello/world.txt/", u.Authority().Path())
				assert.Equal(tb, "id=5&part=three", u.Query())
				assert.Empty(tb, u.Fragment())
			},
		},
		{
			comment: "should assert path and query",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three?another#abc?efg",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "/hello/world.txt/", u.Authority().Path())
				assert.Equal(tb, "id=5&part=three?another", u.Query())
				assert.Equal(tb, "abc?efg", u.Fragment())
				assert.Equal(tb, url.Values{"id": []string{"5"}, "part": []string{"three?another"}}, u.Query())
			},
		},
		{
			comment: "should assert path and query",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0]:8080/a?query=value#fragment",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0", u.Authority().Host())
				assert.Equal(tb, "//user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0]:8080/a", u.Authority().String())
				assert.Equal(tb, "https", u.Scheme())
				assert.Equal(tb, url.Values{"query": []string{"value"}}, u.Query())
			},
		},
		{
			comment: "should parse user/password, IPv6 percent-encoded host with zone",
			uriRaw:  "https://user:passwd@[::1%25lo]:8080/a?query=value#fragment",
			asserter: func(tb testing.TB, u uri.URI) {
				assert.Equal(tb, "https", u.Scheme())
				assert.Equal(tb, "8080", u.Authority().Port())
				assert.Equal(tb, "user:passwd", u.Authority().UserInfo())
			},
		},
	})
}

func rawParseUserInfoTests() iter.Seq[uriTest] {
	return slices.Values([]uriTest{
		{
			comment: "userinfo contains invalid character '{'",
			uriRaw:  "mailto://{}:{}@host.domain.com",
			err:     uri.ErrInvalidUserInfo,
		},
		{
			comment: "invalid user",
			uriRaw:  "https://user{}:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment",
			err:     uri.ErrInvalidUserInfo,
		},
	})
}

func rawParseFailTests() iter.Seq[uriTest] {
	// other failures not already caught by the other test cases
	// (atm empty)
	return slices.Values([]uriTest{
		{
			comment: "invalid scheme (should not be escaped)",
			uriRaw:  "inv%25alidscheme://www.example.com",
			err:     uri.ErrInvalidScheme,
		},
		// This is an invalid UTF8 sequence that SHOULD error, at least in the context of
		// Ref: https://url.spec.whatwg.org/#percent-encoded-bytes
		{
			comment: "invalid query (invalid escape sequence)",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&pa{}%rt=three#there-you-go",
			err:     uri.ErrInvalidQuery,
		},
		{
			comment: "check percent encoding with DNS hostname, invalid escape sequence in host segment",
			uriRaw:  "https://user:passwd@ex%C3ample.com:8080/a?query=value#fragment",
			err:     uri.ErrInvalidDNSName,
		},
		{
			comment: "check percent encoding with registered hostname, invalid escape sequence in host segment",
			uriRaw:  "tel://user:passwd@ex%C3ample.com:8080/a?query=value#fragment",
			err:     uri.ErrInvalidHost,
		},
		{
			comment: "check percent encoding with registered hostname, incomplete escape sequence in host segment",
			uriRaw:  "https://user:passwd@ex%C3ample.com%:8080/a?query=value#fragment",
			err:     uri.ErrInvalidDNSName,
		},
	})
}
