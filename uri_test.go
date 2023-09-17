package uri

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type uriTest struct {
	uriRaw      string
	uri         *uri
	err         error
	comment     string
	isReference bool
	isNotURI    bool
	asserter    func(testing.TB, URI)
}

// errSentinelTest is used in test cases for when we want to assert an error
// but do not want to check specifically which error was returned.
var errSentinelTest = Error(errors.New("test"))

// TestParse exercises the parser that decomposes an URI string in parts.
func TestParse(t *testing.T) {
	t.Parallel()

	// TODO: organize test tables into themes, e.g. structure, scheme, path, etc
	t.Run("with valid URIs",
		testLoop(rawParsePassTests()),
	)

	t.Run("with invalid URIs",
		testLoop(rawParseFailTests()),
	)
}

// TestString asserts the string representation of a parsed URI.
func TestString(t *testing.T) {
	t.Parallel()

	t.Run("with happy path URIs", func(t *testing.T) {
		tests := []string{
			"foo://example.com:8042/over/there?name=ferret#nose",
			"http://httpbin.org/get?utf8=yödeléï",
			"http://httpbin.org/get?utf8=%e2%98%83",
			"mailto://user@domain.com",
			"ssh://user@git.openstack.org:29418/openstack/keystone.git",
			"https://willo.io/#yolo",
		}

		for _, toPin := range tests {
			test := toPin

			t.Run(fmt.Sprintf("should parse %q", test), func(t *testing.T) {
				t.Parallel()

				uri, err := Parse(test)
				require.NoErrorf(t, err,
					"failed to parse URI %q, err: %v", test, err,
				)

				assert.Equalf(t, test, uri.String(),
					"uri.String() != test: %v != %v", uri.String(), test,
				)
			})
		}
	})
}

func TestValidateScheme(t *testing.T) {
	t.Run("scheme should not be shorter than 2 characters", func(t *testing.T) {
		u := &uri{}
		require.Error(t, u.validateScheme("x"))
	})
}

func testLoop(testCases []uriTest) func(t *testing.T) {
	// table-driven tests for IsURI, IsURIReference, Parse and ParseReference.
	return func(t *testing.T) {
		for _, toPin := range testCases {
			test := toPin

			var notOrEmpty string
			if test.err != nil {
				notOrEmpty = "NOT"
			}
			testName := fmt.Sprintf("%s should %s parse %q", test.comment, notOrEmpty, test.uriRaw)

			t.Run(testName, func(t *testing.T) {
				t.Parallel()

				assertIsURI(t, test.uriRaw, test.err != nil, test.isReference)

				// parse string as a pure URI
				actual, err := Parse(test.uriRaw)

				// parse string as a URI reference
				actualReference, errReference := ParseReference(test.uriRaw)

				if test.isReference {
					if test.isNotURI {
						require.Errorf(t, err, "expected URI not to be a valid URI reference (only edge cases)")
					} else {
						// usually, all URI's are also URI references,
						// but there are some edge cases (e.g. empty string), marked as "isNotURI"
						require.NoErrorf(t, errReference, "expected a URI to be a valid URI reference (exluding edge cases)")
					}

					// tests marked "isReference" are pure references that fail URI (i.e. not scheme)
					require.Errorf(t, err, "expected a relative URI reference NOT to be a valid URI")

					actual = actualReference
					err = errReference
				}

				if err == nil && errReference == nil {
					assert.Equalf(t, actual, actualReference,
						"expected Parse and ParseReference to yield the same result",
					)
				}

				if test.err != nil {
					assertError(t, test.err, err)

					return
				}

				require.NoError(t, err)
				if test.asserter != nil {
					test.asserter(t, actual)
				}

				if test.uri != nil {
					// we want to assert struct in-depth, otherwise no error is good enough
					assertURI(t, test.uriRaw, test.uri, actual)
				}
			})
		}
	}
}

func assertURI(t *testing.T, raw string, expected, actual interface{}) {
	require.Equalf(t, expected, actual,
		"got unexpected result (raw: %s), uri: %v != %v",
		raw,
		fmt.Sprintf("%#v", actual),
		fmt.Sprintf("%#v", expected),
	)
}

func assertError(t *testing.T, expected, err error) {
	require.Errorf(t, err,
		"expected Parse to return an error",
	)

	// check error type
	_, isURIError := err.(Error)
	require.Truef(t, isURIError,
		"expected any error returned to be of type uri.Error, but got %T", err,
	)

	if errors.Is(expected, errSentinelTest) {
		// we got an error, and that is good enough
		t.Logf("by the way, this test returned: %v", err)

		return
	}

	// inquire about which error value was returned, specifically
	require.ErrorIsf(t, err, expected,
		"got unexpected err: %v, expected: %v",
		err,
		expected,
	)
}

func assertIsURI(t *testing.T, raw string, expectError, isReference bool) {
	if isReference {
		if expectError {
			require.False(t, IsURIReference(raw),
				"expected %q to be an invalid URI reference", raw,
			)

			return
		}

		require.True(t, IsURIReference(raw),
			"expected %q to be a valid URI reference", raw,
		)

		return
	}

	if expectError {
		require.False(t, IsURI(raw),
			"expected %q to be an invalid URI", raw,
		)

		return
	}

	require.Truef(t, IsURI(raw),
		"expected %q to be a valid URI", raw,
	)
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
				authority: &authorityInfo{
					prefix:   "//",
					userinfo: "",
					host:     "example.com",
					port:     "8042",
					path:     "/over/there",
					isIPv6:   false,
				},
			},
			err: nil,
		},
		{
			uriRaw: "http://httpbin.org/get?utf8=%e2%98%83",
			uri: &uri{
				scheme:   "http",
				hierPart: "//httpbin.org/get",
				query:    "utf8=%e2%98%83",
				fragment: "",
				authority: &authorityInfo{
					prefix:   "//",
					userinfo: "",
					host:     "httpbin.org",
					port:     "",
					path:     "/get",
					isIPv6:   false,
				},
			},
			err: nil,
		},
		{
			uriRaw: "mailto://user@domain.com",
			uri: &uri{
				scheme:   "mailto",
				hierPart: "//user@domain.com",
				query:    "",
				fragment: "",
				authority: &authorityInfo{
					prefix:   "//",
					userinfo: "user",
					host:     "domain.com",
					port:     "",
					path:     "",
					isIPv6:   false,
				},
			},
			err: nil,
		},
		{
			uriRaw: "ssh://user@git.openstack.org:29418/openstack/keystone.git",
			uri: &uri{
				scheme:   "ssh",
				hierPart: "//user@git.openstack.org:29418/openstack/keystone.git",
				query:    "",
				fragment: "",
				authority: &authorityInfo{
					prefix:   "//",
					userinfo: "user",
					host:     "git.openstack.org",
					port:     "29418",
					path:     "/openstack/keystone.git",
					isIPv6:   false,
				},
			},
			err: nil,
		},
		{
			uriRaw: "https://willo.io/#yolo",
			uri: &uri{
				scheme:   "https",
				hierPart: "//willo.io/",
				query:    "",
				fragment: "yolo",
				authority: &authorityInfo{
					prefix:   "//",
					userinfo: "",
					host:     "willo.io",
					port:     "",
					path:     "/",
					isIPv6:   false,
				},
			},
			err: nil,
		},
		{
			comment: "// without // prefix, this parsed as a path!",
			uriRaw:  "mailto:user@domain.com",
			uri: &uri{
				scheme:   "mailto",
				hierPart: "user@domain.com",
				query:    "",
				fragment: "",
				authority: &authorityInfo{
					path: "user@domain.com",
				},
			},
			err: nil,
		},
		{
			comment: "with // prefix, this parsed as a user + host",
			uriRaw:  "mailto://user@domain.com",
			uri: &uri{
				scheme:   "mailto",
				hierPart: "//user@domain.com",
				query:    "",
				fragment: "",
				authority: &authorityInfo{
					prefix:   "//",
					userinfo: "user",
					host:     "domain.com",
				},
			},
			err: nil,
		},
		{
			comment: "with IPv6 host",
			uriRaw:  "mailto://user@[fe80::1]",
			uri: &uri{
				scheme:   "mailto",
				hierPart: "//user@[fe80::1]",
				query:    "",
				fragment: "",
				authority: &authorityInfo{
					prefix:   "//",
					userinfo: "user",
					host:     "fe80::1",
					isIPv6:   true,
				},
			},
			err: nil,
		},
		{
			comment: "legit use of several starting /'s in path'",
			uriRaw:  "file://hostname//etc/hosts",
			asserter: func(t testing.TB, u URI) {
				auth := u.Authority()
				require.Equal(t, "//etc/hosts", auth.Path())
				require.Equal(t, "hostname", auth.Host())
			},
		},
		{
			comment: "authority not empty, valid path with double '/' (see issue#3)",
			uriRaw:  "http://host:8080//foo.html",
		},
		{
			comment: "urn scheme",
			uriRaw:  "urn://example-bin.org/path",
		},
		{
			comment: "authorized dash '-' in host",
			uriRaw:  "https://example-bin.org/path",
		},
		{
			comment: "empty query afer ?",
			uriRaw:  "https://example-bin.org/path?",
		},
		{
			comment: "empty fragment",
			uriRaw:  "mailto://u:p@host.domain.com#",
		},
		{
			comment: "empty query and fragment",
			uriRaw:  "mailto://u:p@host.domain.com?#",
		},
		{
			comment: "only scheme (DNS host), valid!",
			uriRaw:  "http:",
		},
		{
			comment: "only scheme (registered name empty), valid!",
			uriRaw:  "foo:",
		},
		{
			comment: "path with drive letter (e.g. windows)",
			// this one is dubious: Microsoft (.Net) recognizes the C:/... string as a path and
			// states this as incorrect uri -- all other validators state a host "c" and state this uri as a valid one
			uriRaw: "file://c:/directory/filename",
			// TODO: add in-depth assert
		},
		// The unambiguous correct URI notation is  file:///c:/directory/filename TODO with in-depth assert
		// from https://metacpan.org/source/SONNEN/Data-Validate-URI-0.07/t/is_uri.t
		// (many of those come from the rfc3986 examples)
		{
			comment: "simple host and path",
			uriRaw:  "http://localhost/",
		},
		{
			comment: "path with escaped spaces",
			uriRaw:  "http://example.w3.org/path%20with%20spaces.html",
			// TODO: add in-depth assert
		},
		{
			comment: "path is just an escape space",
			uriRaw:  "http://example.w3.org/%20",
		},
		{
			comment: "host with many segments",
			uriRaw:  "ftp://ftp.is.co.za/rfc/rfc1808.txt",
			// TODO: add in-depth assert
		},
		{
			comment: "dots in path",
			uriRaw:  "ftp://ftp.is.co.za/../../../rfc/rfc1808.txt",
		},
		{
			comment: "(redundant)",
			uriRaw:  "http://www.ietf.org/rfc/rfc2396.txt",
		},
		// redundant: "mailto://John.Doe@example.com", // this is the right way
		{
			comment: "= in path",
			uriRaw:  "ldap://[2001:db8::7]/c=GB?objectClass?one",
		},
		{
			comment: "no prefix, everything in scheme (urn-like)",
			uriRaw:  "news:comp.infosystems.www.servers.unix",
		},
		{
			comment: "+ and - in scheme (e.g. tel resource)",
			uriRaw:  "tel:+1-816-555-1212",
		},
		{
			comment: "IPv4 address in host",
			uriRaw:  "telnet://192.0.2.16:80/",
		},
		{
			comment: "(redundant)",
			uriRaw:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
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
			comment: "complete with IPv6",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080/a?query=value#fragment",
		},
		{
			comment: "ipv6 host with zone",
			uriRaw:  "https://user:passwd@[::1%25lo]:8080/a?query=value#fragment",
		},
		{
			comment: "ipv6 host with zone",
			uriRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en1]:8080/a?query=value#fragment",
		},
		{
			comment: "ipv4 host",
			uriRaw:  "https://user:passwd@127.0.0.1:8080/a?query=value#fragment",
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
			comment: `If a URI contains an authority component,
			then the path component must either be empty or begin with a slash ("/") character`,
			uriRaw: "https://host:8080?query=value#fragment",
		},
		{
			comment: "path must begin with / (2)",
			uriRaw:  "https://host:8080/a?query=value#fragment",
		},
		{comment: "ipv4 address", uriRaw: "http://192.168.0.1/"},
		{comment: "ipv4 address with port", uriRaw: "http://192.168.0.1:8080/"},
		{comment: "ipv6 address", uriRaw: "http://[fe80::1]/"},
		{comment: "ipv6 address with port", uriRaw: "http://[fe80::1]:8080/"},
		{comment: "ipv6 with zone identifier", uriRaw: "http://[fe80::1%25en0]/"},
		{comment: "ipv6 with zone and port", uriRaw: "http://[fe80::1%25en0]:8080/"},
		{comment: "ipv6 with percent-encoded and zone", uriRaw: "http://[fe80::1%25%65%6e%301-._~]/"},
		{comment: "ipv6 with percent-encoded and zone", uriRaw: "http://[fe80::1%25%65%6e%301-._~]:8080/"},
		{comment: "ipv4 with percent-encoded", uriRaw: "http://192.168.0.%31/"},
		{comment: "ipv6 with percent-encoded", uriRaw: "http://[fe80::%31]/"},
		{
			comment: "double //, legit with escape",
			uriRaw:  "http+unix://%2Fvar%2Frun%2Fsocket/path?key=value",
		},
		{
			comment: "double leading slash, legit context",
			uriRaw:  "http://host:8080//foo.html",
		},
		{uriRaw: "http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go"},
		{uriRaw: "http://www.example.org/hélloô/mötor/world.txt/?id=5&part=three#there-you-go"},
		{uriRaw: "http://www.example.org/hello/yzx;=1.1/world.txt/?id=5&part=three#there-you-go"},
		{uriRaw: "https://user:passwd@127.0.0.1:8080/a?query=value#fragment"},
		{uriRaw: "http://www.詹姆斯.org/"},
		{uriRaw: "file://c:/directory/filename"},
		{uriRaw: "ldap://[2001:db8::7]/c=GB?objectClass?one"},
		{uriRaw: "ldap://[2001:db8::7]:8080/c=GB?objectClass?one"},
		{uriRaw: "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment"},
		{uriRaw: "https://user:passwd@[FF02:30:0:0:0:0:0:5%25lo]:8080/a?query=value#fragment"},
		{uriRaw: "tel:+1-816-555-1212"},
		{uriRaw: "http+unix:/%2Fvar%2Frun%2Fsocket/path?key=value"},
		{uriRaw: "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080/a?query=value#fragment"},
		{
			comment: "should assert scheme",
			uriRaw:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "urn", u.Scheme())
			},
		},
		{
			comment: "should assert path",
			uriRaw:  "https://example-bin.org/path?",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "/path", u.Authority().Path())
			},
		},
		{
			comment: "should assert path",
			uriRaw:  "mailto://u:p@host.domain.com?#",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "", u.Authority().Path())
			},
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
			comment: "should detect invalid scheme",
			uriRaw:  "ht?tps:",
			err:     errSentinelTest,
		},
		{
			comment: "should parse IPv6 host with zone",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25lo]:8080/a?query=value#fragment",
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25lo", u.Authority().Host())
			},
		},
		{
			comment: "should not parse IPv6 host with empty zone",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%]:8080/a?query=value#fragment",
			err:     errSentinelTest,
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
		{comment: "should error on empty host",
			uriRaw: "https://user:passwd@:8080/a?query=value#fragment",
			err:    ErrMissingHost,
		},
		{
			comment: "percent encoded host, with encoded character not being valid",
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
		{comment: "DNS hostname is too long",
			uriRaw: fmt.Sprintf("https://%s/", strings.Repeat("x", 256)),
			err:    ErrInvalidDNSName,
		},
		{comment: "DNS segment in hostname is empty",
			uriRaw: "https://seg..com/",
			err:    ErrInvalidDNSName,
		},
		{comment: "DNS segment ends with unallowed character",
			uriRaw: "https://x-.y.com/",
			err:    ErrInvalidDNSName,
		},
		{comment: "DNS segment in hostname too long",
			uriRaw: fmt.Sprintf("https://%s.%s.com/", strings.Repeat("x", 63), strings.Repeat("y", 64)),
			err:    ErrInvalidDNSName,
		},
		{comment: "registered name containing unallowed character",
			uriRaw: "bob://x|y/",
			err:    ErrInvalidHost,
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
			comment:     "should parse absolute reference with port",
			uriRaw:      "//host.domain.com:8080/a/b",
			isReference: true,
			asserter: func(t testing.TB, u URI) {
				assert.Equal(t, "host.domain.com", u.Authority().Host())
				assert.Equal(t, "8080", u.Authority().Port())
				assert.Equal(t, "/a/b", u.Authority().Path())
			},
		},
		{
			comment:     "should parse absolute reference with query params",
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
			comment:     "should parse absolute reference with query params (reversed)",
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
			comment:     "with invalid URI which is a valid reference",
			uriRaw:      "//not.a.user@not.a.host/just/a/path",
			isReference: true,
		},
		// from https://golang.org/src/net/url/url_test.go
		{comment: "URL is an URI", uriRaw: "http://foo.com"},
		{comment: "URL is an URI", uriRaw: "http://foo.com/"},
		{comment: "URL is an URI", uriRaw: "http://foo.com/path"},
		{comment: "not an URI but a valid reference", uriRaw: "/", isReference: true},
		{comment: "URL is an URI reference", uriRaw: "//not.a.user@not.a.host/just/a/path", isReference: true},
		{comment: "URL is an URI reference", uriRaw: "//not.a.user@%66%6f%6f.com/just/a/path/also", isReference: true},
		{comment: "URL is an URI reference", uriRaw: "*", isReference: true},
		{comment: "URL is an URI reference", uriRaw: "foo.html", isReference: true},
		{comment: "URL is an URI reference", uriRaw: "../dir/", isReference: true},
		{comment: "URL is an URI - host is an IP", uriRaw: "http://192.168.0.1/"},
		{comment: "URL is an URI", uriRaw: "http://192.168.0.1:8080/"},
		{comment: "URL is an URI", uriRaw: "http://[fe80::1]/"},
		{comment: "URL is an URI", uriRaw: "http://[fe80::1]:8080/"},
		// Tests exercising RFC 6874 compliance:
		{comment: "URL is an URI with alphanum zone identifier", uriRaw: "http://[fe80::1%25en0]/"},
		{comment: "URL is an URI with alphanum zone identifier", uriRaw: "http://[fe80::1%25en0]:8080/"},
		{comment: "URL is an URI with percent-encoded+unreserved zone identifier", uriRaw: "http://[fe80::1%25%65%6e%301-._~]/"},
		{comment: "URL is an URI with percent-encoded+unreserved zone identifier", uriRaw: "http://[fe80::1%25%65%6e%301-._~]:8080/"},
		// percent encoding in IP addresses
		{comment: "URL is an URI should be valid: %31 is percent-encoded for '1'", uriRaw: "http://192.168.0.%31/"},
		{comment: "URL is an URI", uriRaw: "http://192.168.0.%31:8080/"},
		{comment: "URL is an URI", uriRaw: "http://[fe80::%31]/"},
		{comment: "URL is an URI", uriRaw: "http://[fe80::%31]:8080/"},
		{comment: "URL is an URI", uriRaw: "http://[fe80::%31%25en0]/"},
		{comment: "URL is an URI", uriRaw: "http://[fe80::%31%25en0]:8080/"},
		{comment: "URL is an URI (redundant empty IPv6)", uriRaw: "http://[]/", err: errSentinelTest},
		{comment: "URL is an URI (redundant Iv6 )", uriRaw: "http://[fe80::1/", err: errSentinelTest},
		{comment: "URL is an URI wrong percent encoding", uriRaw: "http://[fe80::%%31]:8080/", err: errSentinelTest},
		// These two cases are valid as textual representations as
		// described in RFC 4007, but are not valid as address
		// literals with IPv6 zone identifiers in URIs as described in
		// RFC 6874.
		{comment: "URL is an URI", uriRaw: "http://[fe80::1%en0]/", err: errSentinelTest},
		{comment: "URL is an URI", uriRaw: "http://[fe80::1%en0]:8080/", err: errSentinelTest},
		{comment: "URL is an URI (redundant empty string)", uriRaw: "", isReference: true, isNotURI: true},
	}
}

func rawParseFailTests() []uriTest {
	return []uriTest{
		{comment: "ipv6 empty zone", uriRaw: "https://user:passwd@[FF02:30:0:0:0:0:0:5%25]:8080/a?query=value#fragment", err: errSentinelTest},
		{
			comment: ` If a URI does not contain an authority component,
		then the path cannot begin with two slash characters ("//").`,
			uriRaw: "https:////a?query=value#fragment",
			err:    errSentinelTest,
		},
		{
			comment: "ipv6 addresses not between square brackets are invalid hosts",
			uriRaw:  "https://0%3A0%3A0%3A0%3A0%3A0%3A0%3A1/a",
			err:     ErrInvalidHostAddress,
		},
		{
			comment: "ip addresses between square brackets should not be ipv4 addresses",
			uriRaw:  "https://[192.169.224.1]/a",
			err:     ErrInvalidHostAddress,
		},
		{comment: "empty ipv6 (redundant)", uriRaw: "http://[]/", err: ErrInvalidURI}, // TODO: should get a better error
		{
			comment: "path must begin with / (invalid)",
			uriRaw:  "https://host:8080a?query=value#fragment",
			err:     ErrInvalidPort,
		},
		{
			comment: "should detect a path starting with several /'s",
			uriRaw:  "file:////etc/hosts",
			err:     ErrInvalidPath,
		},
		{
			comment: "query contains invalid characters",
			uriRaw:  "http://httpbin.org/get?utf8=\xe2\x98\x83",
			uri: &uri{
				scheme:   "http",
				hierPart: "//httpbin.org/get",
				query:    "utf8=\xe2\x98\x83",
				fragment: "",
				authority: &authorityInfo{
					prefix:   "//",
					userinfo: "",
					host:     "httpbin.org",
					port:     "",
					path:     "/get",
					isIPv6:   false,
				},
			},
			err: ErrInvalidQuery,
		},
		{
			comment: "empty host  => double '/' invalid in this context",
			uriRaw:  "http:////foo.html",
			err:     ErrInvalidPath,
		},
		{
			comment: "should not parse invalid URI",
			uriRaw:  "http://httpbin.org/get?utf8=\xe2\x98\x83",
			err:     ErrInvalidQuery,
		},
		{
			comment: "userinfo contains invalid character '{'",
			uriRaw:  "mailto://{}:{}@host.domain.com",
			err:     errSentinelTest,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (1)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080",
			err:     errSentinelTest,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (2)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080/?#",
			err:     errSentinelTest,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (3)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080#",
			err:     errSentinelTest,
		},
		{
			comment: "missing closing bracket for IPv6 litteral (4)",
			uriRaw:  "https://user:passwd@[FF02::3::5:8080#abc",
			err:     errSentinelTest,
		},
		{
			comment: "should not parse IPv6 host with empty zone",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25]:8080/a?query=value#fragment",
			err:     ErrInvalidHostAddress,
		},
		// from the format test in JSONSchema-test suite
		{
			comment: "missing scheme",
			uriRaw:  "//foo.bar/?baz=qux#quux",
			err:     errSentinelTest,
		},
		// from https://docs.microsoft.com/en-gb/dotnet/api/system.uri.iswellformeduristring?view=netframework-4.7.2#System_Uri_IsWellFormedUriString_System_String_System_UriKind_
		{
			comment: "query is not correctly escaped.",
			uriRaw:  "http://www.contoso.com/path???/file name", //
			err:     errSentinelTest,
		},
		{
			comment: "absolute URI that represents an implicit file URI.",
			uriRaw:  "c:\\directory\filename",
			err:     errSentinelTest,
		},
		{
			comment: "contains unescaped backslashes even if they will be treated as forward slashes",
			uriRaw:  "http:\\host/path/file",
			err:     errSentinelTest,
		},
		{
			comment: "represents a hierarchical absolute URI and does not contain '://'",
			uriRaw:  "www.contoso.com/path/file",
			err:     errSentinelTest,
		},
		{
			comment: "relative URIs with a colon (':') in their first segment are not considered well-formed",
			uriRaw:  "2013.05.29_14:33:41",
			err:     errSentinelTest,
		},
		// from https://metacpan.org/source/SONNEN/Data-Validate-URI-0.07/t/is_uri.t
		{
			comment: "empty URI",
			uriRaw:  "",
			err:     errSentinelTest,
		},
		{
			comment: "no separator",
			uriRaw:  "foo",
			err:     errSentinelTest,
		},
		{
			comment: "no ':' separator",
			uriRaw:  "foo@bar",
			err:     errSentinelTest,
		},
		{
			comment: "illegal characters",
			uriRaw:  "http://<foo>",
			err:     errSentinelTest,
		},
		{
			comment: "empty scheme",
			uriRaw:  "://bob/",
			err:     errSentinelTest,
		},
		{
			comment: "bad scheme: no starting with a letter",
			uriRaw:  "1http://bob",
			err:     errSentinelTest,
		},
		{
			comment: "bad escaping",
			uriRaw:  "http://example.w3.org/%illegal.html",
			err:     errSentinelTest,
		},
		{
			comment: "partial escape (1)",
			uriRaw:  "http://example.w3.org/%a",
			err:     errSentinelTest,
		},
		{
			comment: "partial escape (2)",
			uriRaw:  "http://example.w3.org/%a/foo",
			err:     errSentinelTest,
		},
		{
			comment: "partial escape (3)",
			uriRaw:  "http://example.w3.org/%at",
			err:     errSentinelTest,
		},
		// from https://github.com/python-hyper/rfc3986/blob/master/tests/test_validators.py
		{
			comment: "multiple ports",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment",
			err:     errSentinelTest,
		},
		{
			comment: "invalid IPv6 (double empty ::)",
			uriRaw:  "https://user:passwd@[FF02::3::5]:8080/a?query=value#fragment",
			err:     errSentinelTest,
		},
		{
			comment: "invalid IPv6 (not enough parts)",
			uriRaw:  "https://user:passwd@[FADF:01%en0]:8080/a?query=value#fragment",
			err:     errSentinelTest,
		},
		{
			comment: "invalid IPv4: part>255",
			uriRaw:  "https://user:passwd@256.256.256.256:8080/a?query=value#fragment",
			err:     errSentinelTest,
		},
		{
			comment: "invalid IPv6 (missing bracket)",
			uriRaw:  "https://user:passwd@[FADF:01%en0:8080/a?query=value#fragment",
			err:     errSentinelTest,
		},
		// from github.com/scalatra/rl: URI parser in scala
		{
			comment: "invalid host (contains blank space)",
			uriRaw:  "http://www.exa mple.org",
			err:     errSentinelTest,
		},
		// and others..
		{
			comment: "invalid scheme",
			uriRaw:  "?invalidscheme://www.example.com",
			err:     errSentinelTest,
		},
		{
			comment: "invaid scheme (2)",
			uriRaw:  "inv;alidscheme://www.example.com",
			err:     errSentinelTest,
		},
		{
			comment: "invalid char in query",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&pa{}rt=three#there-you-go",
			err:     errSentinelTest,
		},
		{
			comment: "invalid char in fragment",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go{}",
			err:     errSentinelTest,
		},
		{
			comment: "empty IPV6",
			uriRaw:  "scheme://user:passwd@[]/invalid",
			err:     errSentinelTest,
		},
		{
			comment: "invalid fragment",
			uriRaw:  "http://example.w3.org/legit#ill[egal",
			err:     errSentinelTest,
		},
		{
			comment: "pathological input (1)",
			uriRaw:  "?//x",
			err:     errSentinelTest,
		},
		{
			comment: "pathological input (2)",
			uriRaw:  "#//x",
			err:     errSentinelTest,
		},
		{
			comment: "pathological input (3)",
			uriRaw:  "://x",
			err:     errSentinelTest,
		},
		{
			comment: "pathological input (4)",
			uriRaw:  ".?:",
			err:     errSentinelTest,
		},
		{
			comment: "pathological input (5)",
			uriRaw:  ".#:",
			err:     errSentinelTest,
		},
		{
			comment: "trailing empty fragment, invalid path",
			uriRaw:  "http://example.w3.org/%legit#",
			err:     errSentinelTest,
		},
		{
			comment: "invalid scheme",
			uriRaw:  "1http://bob",
			err:     ErrInvalidScheme,
		},
		{
			comment: "invalid scheme (too short)",
			uriRaw:  "x://bob",
			err:     ErrInvalidScheme,
		},
		{
			comment: "should exhibit a parsing error",
			uriRaw:  "?",
			err:     errSentinelTest,
		},
		{
			comment: "should exhibit a parsing error",
			uriRaw:  "#",
			err:     errSentinelTest,
		},
		{
			comment: "should exhibit a parsing error",
			uriRaw:  "?#",
			err:     errSentinelTest,
		},
		{
			comment:  "should exhibit a parsing error (redundant)",
			uriRaw:   "",
			isNotURI: true,
			err:      errSentinelTest,
		},
		{
			comment: "should exhibit a parsing error",
			uriRaw:  " ",
			err:     errSentinelTest,
		},
		{
			comment: "should exhibit a parsing error",
			uriRaw:  "ht?tp://host",
			err:     errSentinelTest,
		},
		{
			comment: "should detect an invalid query",
			uriRaw:  "http://www.example.org/hello/world.txt/?id=5&pa{}rt=three#there-you-go",
			err:     ErrInvalidQuery,
		},
		{
			comment: "should detect an invalid path",
			uriRaw:  "http://www.example.org/hello/{}yzx;=1.1/world.txt/?id=5&part=three#there-you-go",
			err:     ErrInvalidPath,
		},
		{
			comment: "should detect an invalid host (DNS rule)",
			uriRaw:  "https://user:passwd@286;0.0.1:8080/a?query=value#fragment",
			err:     ErrInvalidHost,
		},
		{
			comment: "should detect an invalid host (DNS rule) (2)",
			uriRaw:  "https://user:passwd@256.256.256.256:8080/a?query=value#fragment",
			err:     ErrInvalidHost,
		},
		{
			comment: "should detect an invalid URI (lack closing bracket)",
			uriRaw:  "https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0:8080/a?query=value#fragment",
			err:     ErrInvalidURI,
		},
		{
			comment: "should detect an invalid port",
			uriRaw:  "https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment",
			err:     ErrInvalidPort,
		},
		{
			comment: "should detect an invalid user",
			uriRaw:  "https://user{}:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment",
			err:     ErrInvalidUserInfo,
		},
	}
}
