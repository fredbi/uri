package uri

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	uriTest struct {
		uriRaw string
		uri    *uri
		err    error
	}

	urlTest struct {
		url                    string
		expectedValid          bool
		expectedValidReference bool
	}
)

func Test_rawURIParse(t *testing.T) {
	t.Parallel()

	allTests := rawParseTests()
	allTests = append(allTests, rawParseFailedTests()...)

	for _, toPin := range allTests {
		test := toPin

		t.Run(fmt.Sprintf("should parse %q", test.uriRaw), func(t *testing.T) {
			t.Parallel()

			got, err := Parse(test.uriRaw)
			if test.err != nil {
				require.ErrorIsf(t, err, test.err,
					"got unexpected err: %v != %v",
					err,
					test.err,
				)

				return
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(got, test.uri),
				"got unexpected result (raw: %s), uri: %v != %v",
				test.uriRaw,
				fmt.Sprintf("%#v", got),
				fmt.Sprintf("%#v", test.uri),
			)
		})
	}
}

func rawParseTests() []uriTest {
	return []uriTest{
		{
			"foo://example.com:8042/over/there?name=ferret#nose",
			&uri{
				"foo", "//example.com:8042/over/there", "name=ferret", "nose",
				&authorityInfo{
					"//",
					"",
					"example.com",
					"8042",
					"/over/there",
				},
			},
			nil,
		},
		{
			"http://httpbin.org/get?utf8=%e2%98%83",
			&uri{
				"http", "//httpbin.org/get", "utf8=%e2%98%83", "",
				&authorityInfo{
					"//",
					"",
					"httpbin.org",
					"",
					"/get",
				},
			},
			nil,
		},
		{
			"mailto://user@domain.com",
			&uri{
				"mailto", "//user@domain.com", "", "",
				&authorityInfo{
					"//",
					"user",
					"domain.com",
					"",
					"",
				},
			},
			nil,
		},
		{
			"ssh://user@git.openstack.org:29418/openstack/keystone.git",
			&uri{
				"ssh", "//user@git.openstack.org:29418/openstack/keystone.git", "", "",
				&authorityInfo{
					"//",
					"user",
					"git.openstack.org",
					"29418",
					"/openstack/keystone.git",
				},
			},
			nil,
		},
		{
			"https://willo.io/#yolo",
			&uri{
				"https", "//willo.io/", "", "yolo",
				&authorityInfo{"//", "", "willo.io", "", "/"},
			},
			nil,
		},
	}
}

func rawParseFailedTests() []uriTest {
	return []uriTest{
		{
			"http://httpbin.org/get?utf8=\xe2\x98\x83",
			&uri{
				"http", "//httpbin.org/get", "utf8=\xe2\x98\x83", "",
				&authorityInfo{
					"//",
					"",
					"httpbin.org",
					"",
					"/get",
				},
			},
			ErrInvalidQuery,
		},
		{
			// without // prefix, this is a path!
			"mailto:user@domain.com",
			&uri{
				"mailto", "user@domain.com", "", "",
				&authorityInfo{
					path: "user@domain.com",
				},
			},
			nil,
		},
	}
}

func Test_ParseThenString(t *testing.T) {
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

	t.Run("should not parse invalid URI", func(t *testing.T) {
		_, err := Parse("http://httpbin.org/get?utf8=\xe2\x98\x83")
		// no normalization takes place at the moment
		assert.Error(t, err)
	})
}

// TestMoreURI borrows from other URI validators to exercise strict RFC3986
// conformance (taken from .Net, perl, python, ).
func TestMoreURI(t *testing.T) {
	t.Parallel()

	invalidURIs := []string{
		"mailto://{}:{}@host.domain.com",
		"https://user:passwd@[FF02::3::5:8080",
		"https://user:passwd@[FF02::3::5:8080/?#",
		"https://user:passwd@[FF02::3::5:8080#",
		"https://user:passwd@[FF02::3::5:8080#abc",

		// this test comes from the format test in JSONSchema-test suite
		"//foo.bar/?baz=qux#quux", // missing scheme and //

		// from https://docs.microsoft.com/en-gb/dotnet/api/system.uri.iswellformeduristring?view=netframework-4.7.2#System_Uri_IsWellFormedUriString_System_String_System_UriKind_
		"http://www.contoso.com/path???/file name", // The string is not correctly escaped.
		"c:\\directory\filename",                   // The string is an absolute Uri that represents an implicit file Uri.
		"http:\\host/path/file",                    // The string contains unescaped backslashes even if they will be treated as forward slashes
		"www.contoso.com/path/file",                // The string represents a hierarchical absolute Uri and does not contain "://"
		"2013.05.29_14:33:41",                      // relative URIs with a colon (':') in their first segment are not considered well-formed.

		// from https://metacpan.org/source/SONNEN/Data-Validate-URI-0.07/t/is_uri.t
		"",
		"foo",
		"foo@bar",
		"http://<foo>", // illegal characters
		"://bob/",      // empty scheme
		"1http://bob",  // bad scheme
		"http://example.w3.org/%illegal.html",
		"http://example.w3.org/%a",     // partial escape
		"http://example.w3.org/%a/foo", // partial escape
		"http://example.w3.org/%at",    // partial escape

		// from https://github.com/python-hyper/rfc3986/blob/master/tests/test_validators.py
		"https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment", // multiple ports
		"https://user:passwd@[FF02::3::5]:8080/a?query=value#fragment",                                   // invalid IPv6
		"https://user:passwd@[FADF:01%en0]:8080/a?query=value#fragment",                                  // invalid IPv6
		"https://user:passwd@256.256.256.256:8080/a?query=value#fragment",                                // invalid IPv4
		"https://user:passwd@[FADF:01%en0:8080/a?query=value#fragment",                                   // invalid IPv6 (missing bracket)

		// from github.com/scalatra/rl: URI parser in scala
		"http://www.exa mple.org",

		// and others..
		"?invalidscheme://www.example.com",
		"inv;alidscheme://www.example.com",
		"http://www.example.org/hello/world.txt/?id=5&pa{}rt=three#there-you-go", // invalid char in query
		"http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go{}", // invalid char in fragment
		"scheme://user:passwd@[]/invalid",                                        // empty IPV6

		// invalid fragment
		"http://example.w3.org/legit#ill[egal",

		// pathological input
		"?//x",
		"#//x",
		"://x",

		// trailing empty fragment, invalid path
		"http://example.w3.org/%legit#",
	}

	validURIs := []string{
		"http:////foo.html", // empty host, correct path (see issue#3)
		"urn://example-bin.org/path",
		"https://example-bin.org/path",
		"https://example-bin.org/path?",
		"mailto://u:p@host.domain.com#",  // empty fragment
		"mailto://u:p@host.domain.com?#", // empty query + fragment
		"http:",
		"foo:",

		// this one is dubious: Microsoft (.Net) recognize the C:/... string as a path and
		// states this as incorrect uri -- all other validators state a host "c" and state this uri as a valid one
		"file://c:/directory/filename",

		// from https://metacpan.org/source/SONNEN/Data-Validate-URI-0.07/t/is_uri.t
		// (many of those come from the rfc3986 examples)
		"http://localhost/",
		"http://example.w3.org/path%20with%20spaces.html",
		"http://example.w3.org/%20",
		"ftp://ftp.is.co.za/rfc/rfc1808.txt",
		"ftp://ftp.is.co.za/../../../rfc/rfc1808.txt",
		"http://www.ietf.org/rfc/rfc2396.txt",
		"ldap://[2001:db8::7]/c=GB?objectClass?one",
		"mailto:John.Doe@example.com",   // valid but counter-intuitive: userinfo is actually a path
		"mailto://John.Doe@example.com", // this is the right way
		"news:comp.infosystems.www.servers.unix",
		"tel:+1-816-555-1212",
		"telnet://192.0.2.16:80/",
		"urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
		"http://www.richardsonnen.com/",

		// from https://github.com/python-hyper/rfc3986/blob/master/tests/test_validators.py
		"ssh://ssh@git.openstack.org:22/sigmavirus24",
		"https://git.openstack.org:443/sigmavirus24",
		"ssh://git.openstack.org:22/sigmavirus24?foo=bar#fragment",
		"git://github.com",
		"https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080/a?query=value#fragment",
		"https://user:passwd@[::1%25lo]:8080/a?query=value#fragment",
		"https://user:passwd@[FF02:30:0:0:0:0:0:5%25en1]:8080/a?query=value#fragment",
		"https://user:passwd@127.0.0.1:8080/a?query=value#fragment",
		"https://user:passwd@http-bin.org:8080/a?query=value#fragment",

		// from github.com/scalatra/rl: URI parser in scala
		"http://www.example.org:8080",
		"http://www.example.org/",
		"http://www.詹姆斯.org/",
		"http://www.example.org/hello/world.txt",
		"http://www.example.org/hello/world.txt/?id=5&part=three",
		"http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go",
		"http://www.example.org/hello/world.txt/#here-we-are",

		// trailing empty fragment: legit
		"http://example.w3.org/legit#",
	}

	t.Run("with invalid URIs", func(t *testing.T) {
		t.Parallel()

		for _, toPin := range invalidURIs {
			invURI := toPin

			t.Run(fmt.Sprintf("should be invalid: %q", invURI), func(t *testing.T) {
				t.Parallel()

				require.Falsef(t, IsURI(invURI),
					"expected %q to be an invalid URI", invURI,
				)
			})
		}
	})

	t.Run("with valid URIs", func(t *testing.T) {
		t.Parallel()

		for _, toPin := range validURIs {
			validURI := toPin

			t.Run(fmt.Sprintf("should be valid: %q", validURI), func(t *testing.T) {
				t.Parallel()

				require.Truef(t, IsURI(validURI),
					"expected %q to be a valid URI", validURI,
				)

				_, err := Parse(validURI)
				require.NoError(t, err)
			})
		}
	})
}

func Test_MoreParse(t *testing.T) {
	t.Parallel()

	t.Run("should detect an invalid scheme", func(t *testing.T) {
		_, err := Parse("1http://bob")
		assert.Equal(t, ErrInvalidScheme, err)
	})

	t.Run("should parse without error", func(t *testing.T) {
		_, err := Parse("http://www.example.org/hello/world.txt/?id=5&part=three#there-you-go")
		assert.NoError(t, err)

		_, err = Parse("http://www.example.org/hélloô/mötor/world.txt/?id=5&part=three#there-you-go")
		assert.NoError(t, err)

		_, err = Parse("http://www.example.org/hello/yzx;=1.1/world.txt/?id=5&part=three#there-you-go")
		assert.NoError(t, err)

		_, err = Parse("https://user:passwd@127.0.0.1:8080/a?query=value#fragment")
		assert.NoError(t, err)

		_, err = Parse("http://www.詹姆斯.org/")
		assert.NoError(t, err)

		_, err = Parse("file://c:/directory/filename")
		assert.NoError(t, err)

		_, err = Parse("ldap://[2001:db8::7]/c=GB?objectClass?one")
		assert.NoError(t, err)

		_, err = Parse("ldap://[2001:db8::7]:8080/c=GB?objectClass?one")
		assert.NoError(t, err)

		_, err = Parse("https://user:passwd@[FF02:30:0:0:0:0:0:5%25]:8080/a?query=value#fragment")
		assert.NoError(t, err)

		_, err = Parse("https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment")
		assert.NoError(t, err)

		_, err = Parse("https://user:passwd@[FF02:30:0:0:0:0:0:5%25lo]:8080/a?query=value#fragment")
		assert.NoError(t, err)

		_, err = Parse("tel:+1-816-555-1212")
		assert.NoError(t, err)

		_, err = Parse("http+unix:/%2Fvar%2Frun%2Fsocket/path?key=value")
		assert.NoError(t, err)

		_, err = Parse("https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080/a?query=value#fragment")
		assert.NoError(t, err)
	})

	t.Run("should detect an invalid query", func(t *testing.T) {
		_, err := Parse("http://www.example.org/hello/world.txt/?id=5&pa{}rt=three#there-you-go")
		assert.Equal(t, ErrInvalidQuery, err)
	})

	t.Run("should detect an invalid path", func(t *testing.T) {
		_, err := Parse("http://www.example.org/hello/{}yzx;=1.1/world.txt/?id=5&part=three#there-you-go")
		assert.Equal(t, ErrInvalidPath, err)
	})

	t.Run("should detect an invalid host", func(t *testing.T) {
		_, err := Parse("https://user:passwd@286;0.0.1:8080/a?query=value#fragment")
		assert.Equal(t, ErrInvalidHost, err)

		_, err = Parse("https://user:passwd@256.256.256.256:8080/a?query=value#fragment")
		assert.Equal(t, ErrInvalidHost, err)

		_, err = Parse("http+unix://%2Fvar%2Frun%2Fsocket/path?key=value") // no authority => no "//"
		assert.Equal(t, ErrInvalidHost, err)
	})

	t.Run("should detect an invalid URI (lack closing bracket)", func(t *testing.T) {
		_, err := Parse("https://user:passwd@[FF02:30:0:0:0:0:0:5%25en0:8080/a?query=value#fragment")
		assert.Equal(t, ErrInvalidURI, err)
	})

	t.Run("should detect an invalid port", func(t *testing.T) {
		_, err := Parse("https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A]:8080:8090/a?query=value#fragment")
		assert.Equal(t, ErrInvalidPort, err)
	})

	t.Run("should detect an invalid user", func(t *testing.T) {
		_, err := Parse("https://user{}:passwd@[FF02:30:0:0:0:0:0:5%25en0]:8080/a?query=value#fragment")
		assert.Equal(t, ErrInvalidUserInfo, err)
	})

	t.Run("should parse and verify properties", func(t *testing.T) {
		t.Run("should assert scheme", func(t *testing.T) {
			u, err := Parse("urn:oasis:names:specification:docbook:dtd:xml:4.1.2")
			assert.NoError(t, err)
			assert.Equal(t, "urn", u.Scheme())
		})

		t.Run("should assert path", func(t *testing.T) {
			u, err := Parse("https://example-bin.org/path?")
			assert.NoError(t, err)
			assert.Equal(t, "/path", u.Authority().Path())

			u, err = Parse("mailto://u:p@host.domain.com?#")
			assert.NoError(t, err)
			assert.Equal(t, "", u.Authority().Path())
		})

		t.Run("should assert path and fragment", func(t *testing.T) {
			u, err := Parse("https://example-bin.org/path#frag?withQuestionMark")
			assert.NoError(t, err)
			assert.Equal(t, "/path", u.Authority().Path())
			nuri := u.(*uri)
			assert.Equal(t, "", nuri.query)
			assert.Equal(t, "frag?withQuestionMark", u.Fragment())

			u, err = Parse("mailto://u:p@host.domain.com?#ahahah")
			assert.NoError(t, err)
			assert.Equal(t, "", u.Authority().Path())
			nuri = u.(*uri)
			assert.Equal(t, "", nuri.query)
			assert.Equal(t, "ahahah", u.Fragment())
		})

		t.Run("should assert path and query", func(t *testing.T) {
			u, err := Parse("ldap://[2001:db8::7]/c=GB?objectClass?one")
			assert.NoError(t, err)
			assert.Equal(t, "/c=GB", u.Authority().Path())
			nuri := u.(*uri)
			assert.Equal(t, "objectClass?one", nuri.query)
			assert.Equal(t, "", u.Fragment())

			u, err = Parse("http://www.example.org/hello/world.txt/?id=5&part=three")
			assert.NoError(t, err)
			assert.Equal(t, "/hello/world.txt/", u.Authority().Path())
			nuri = u.(*uri)
			assert.Equal(t, "id=5&part=three", nuri.query)
			assert.Equal(t, "", u.Fragment())

			u, err = Parse("http://www.example.org/hello/world.txt/?id=5&part=three?another#abc?efg")
			assert.NoError(t, err)
			assert.Equal(t, "/hello/world.txt/", u.Authority().Path())
			nuri = u.(*uri)
			assert.Equal(t, "id=5&part=three?another", nuri.query)
			assert.Equal(t, "abc?efg", u.Fragment())
			assert.Equal(t, url.Values{"id": []string{"5"}, "part": []string{"three?another"}}, u.Query())

			u, err = Parse("https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0]:8080/a?query=value#fragment")
			assert.NoError(t, err)
			assert.Equal(t, "21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0", u.Authority().Host())
			assert.Equal(t, "//user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25en0]:8080/a", u.Authority().String())
			assert.Equal(t, "https", u.Scheme())
			assert.Equal(t, url.Values{"query": []string{"value"}}, u.Query())
		})
	})

	t.Run("should exhibit a parsing error", func(t *testing.T) {
		_, err := Parse("?")
		assert.Error(t, err)

		_, err = Parse("#")
		assert.Error(t, err)

		_, err = Parse("?#")
		assert.Error(t, err)

		_, err = Parse("")
		assert.Error(t, err)

		_, err = Parse(" ")
		assert.Error(t, err)

		_, err = Parse("ht?tp://host")
		assert.Error(t, err)
	})
}

func Test_Edge(t *testing.T) {
	t.Parallel()

	t.Run("with scheme only", func(t *testing.T) {
		u, err := Parse("https:")
		require.NoError(t, err)
		assert.Equal(t, "https", u.Scheme())
	})

	t.Run("should detect invalid scheme", func(t *testing.T) {
		_, err := Parse("ht?tps:")
		require.Error(t, err)
	})

	t.Run("should parse IPv6 host", func(t *testing.T) {
		u, err := Parse("https://user:passwd@[21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25]:8080/a?query=value#fragment")
		require.NoError(t, err)

		assert.Equal(t, "21DA:00D3:0000:2F3B:02AA:00FF:FE28:9C5A%25", u.Authority().Host())
	})

	t.Run("should parse user/password, IPv6 percent-encoded host", func(t *testing.T) {
		u, err := Parse("https://user:passwd@[::1%25lo]:8080/a?query=value#fragment")
		require.NoError(t, err)

		assert.Equal(t, "https", u.Scheme())
		assert.Equal(t, "8080", u.Authority().Port())
		assert.Equal(t, "user:passwd", u.Authority().UserInfo())
	})

	t.Run("should error on empty host", func(t *testing.T) {
		_, err := Parse("https://user:passwd@:8080/a?query=value#fragment")
		require.Equal(t, ErrMissingHost, err)
	})

	t.Run("percent encoded host", func(t *testing.T) {
		_, err := Parse("urn://user:passwd@ex%7Cample.com:8080/a?query=value#fragment")
		require.Errorf(t, err,
			"expected uri with percent-encoded host to be invalid",
		)

		u, err := Parse("urn://user:passwd@ex%2Dample.com:8080/a?query=value#fragment")
		require.NoErrorf(t, err,
			"expected uri with percent-encoded host to be valid",
		)
		assert.Equal(t, "ex%2Dample.com", u.Authority().Host())
	})

	t.Run("check percent encoding with DNS hostname", func(t *testing.T) {
		u, err := Parse("https://user:passwd@ex%2Dample.com:8080/a?query=value#fragment")
		require.NoErrorf(t, err, "expected uri with percent-encoded host to be valid")
		assert.Equal(t, "ex%2Dample.com", u.Authority().Host())
	})
}

// Test_Relative asserts that relative uris are invalid (e.g. missing scheme).
func Test_Relative(t *testing.T) {
	t.Parallel()

	invalidURIrefs := []string{
		"//host.domain.com/a/b",
		"//host.domain.com:8080/a/b",
	}

	for _, toPin := range invalidURIrefs {
		invalidURIref := toPin

		t.Run(fmt.Sprintf("with invalid URI %q", invalidURIref), func(t *testing.T) {
			t.Parallel()

			t.Run("should error on invalid URI", func(t *testing.T) {

				_, err := Parse(invalidURIref)
				assert.Error(t, err)
			})

			t.Run("should not error on invalid URI but valid reference", func(t *testing.T) {
				_, err := ParseReference(invalidURIref)
				assert.NoError(t, err)
			})
		})
	}
}

func Test_AbsoluteReference(t *testing.T) {
	t.Parallel()

	t.Run("should parse absolute reference with port", func(t *testing.T) {
		v, _ := ParseReference("//host.domain.com:8080/a/b")
		require.NotNil(t, v)

		assert.Equal(t, "host.domain.com", v.Authority().Host())
		assert.Equal(t, "8080", v.Authority().Port())
		assert.Equal(t, "/a/b", v.Authority().Path())
	})

	t.Run("should parse absolute reference with query params", func(t *testing.T) {
		v, _ := ParseReference("//host.domain.com:8080?query=x/a/b")
		require.NotNil(t, v)

		assert.Equal(t, "host.domain.com", v.Authority().Host())
		assert.Equal(t, "8080", v.Authority().Port())
		assert.Equal(t, "", v.Authority().Path())
		assert.Equal(t, "x/a/b", v.Query().Get("query"))
	})

	t.Run("should parse absolute reference with query params (reversed)", func(t *testing.T) {
		v, _ := ParseReference("//host.domain.com:8080/a/b?query=x")
		require.NotNil(t, v)

		assert.Equal(t, "host.domain.com", v.Authority().Host())
		assert.Equal(t, "8080", v.Authority().Port())
		assert.Equal(t, "/a/b", v.Authority().Path())
		assert.Equal(t, "x", v.Query().Get("query"))
	})
}

const pathThatLooksSchemeRelative = "//not.a.user@not.a.host/just/a/path"

// Test_URL verifies that go all url stdlib tests pass as uri with this package.
// valid URLs are valid URI or valid URI references
//
// See https://golang.org/src/net/url/url_test.go
//
// NOTE: this package makes a strict distinction between uri and uri-reference.
func Test_URL(t *testing.T) {
	t.Parallel()

	t.Run("with URL input", func(t *testing.T) {
		for _, toPin := range parseRequestURLTests() {
			test := toPin

			t.Run(fmt.Sprintf("should parse %q", test.url), func(t *testing.T) {
				_, err := Parse(test.url)
				if !test.expectedValid {
					require.Errorf(t, err,
						"parse(%q) gave nil error; want some error", test.url,
					)

					return
				}

				require.NoErrorf(t, err,
					"parse(%q) gave err %v; want no error", test.url, err,
				)

				isRef := IsURIReference(test.url)
				assert.Equalf(t, test.expectedValidReference, isRef,
					"IsURIReference(%q) returned %t; want %t", test.url, isRef, test.expectedValidReference,
				)
			})
		}
	})

	t.Run("with invalid URI which is a valid reference", func(t *testing.T) {
		_, err := Parse(pathThatLooksSchemeRelative)
		assert.Error(t, err)

		_, err = ParseReference(pathThatLooksSchemeRelative)
		assert.NoError(t, err)
	})
}

func parseRequestURLTests() []urlTest {
	return []urlTest{
		{"http://foo.com", true, true},
		{"http://foo.com/", true, true},
		{"http://foo.com/path", true, true},
		{"/", false, true},
		{pathThatLooksSchemeRelative, false, true},
		{"//not.a.user@%66%6f%6f.com/just/a/path/also", false, true},
		{"*", false, true}, // ???
		{"http://192.168.0.1/", true, true},
		{"http://192.168.0.1:8080/", true, true},
		{"http://[fe80::1]/", true, true},
		{"http://[fe80::1]:8080/", true, true},
		// Tests exercising RFC 6874 compliance:
		{"http://[fe80::1%25en0]/", true, true},                 // with alphanum zone identifier
		{"http://[fe80::1%25en0]:8080/", true, true},            // with alphanum zone identifier
		{"http://[fe80::1%25%65%6e%301-._~]/", true, true},      // with percent-encoded+unreserved zone identifier
		{"http://[fe80::1%25%65%6e%301-._~]:8080/", true, true}, // with percent-encoded+unreserved zone identifier
		{"foo.html", false, true},
		{"../dir/", false, true},
		{"http://192.168.0.%31/", false, false},
		{"http://192.168.0.%31:8080/", false, false},
		{"http://[fe80::%31]/", false, false},
		{"http://[fe80::%31]:8080/", false, false},
		{"http://[fe80::%31%25en0]/", false, false},
		{"http://[fe80::%31%25en0]:8080/", false, false},
		// These two cases are valid as textual representations as
		// described in RFC 4007, but are not valid as address
		// literals with IPv6 zone identifiers in URIs as described in
		// RFC 6874.
		{"http://[fe80::1%en0]/", false, false},
		{"http://[fe80::1%en0]:8080/", false, false},
		// Added this
		{"", false, true},
	}
}

func Test_Issue3(t *testing.T) {
	t.Run("should detect a path starting with a /", func(t *testing.T) {
		u, err := Parse("file:///etc/hosts")
		require.NoError(t, err)
		auth := u.Authority()
		require.Equal(t, "/etc/hosts", auth.Path())
		require.Empty(t, auth.Host())
	})

	t.Run("should detect a path starting with several /'s", func(t *testing.T) {
		u, err := Parse("file:////etc/hosts")
		require.NoError(t, err)
		auth := u.Authority()
		require.Equal(t, "//etc/hosts", auth.Path())
		require.Empty(t, auth.Host())
	})
}
