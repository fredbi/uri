package uri

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	uriTest struct {
		uriRaw      string
		uri         *uri
		err         error
		comment     string
		isReference bool
		isNotURI    bool
		asserter    func(testing.TB, URI)
	}

	testGenerator func() []uriTest
)

// TestParse exercises the parser that decomposes an URI string in parts.
func TestParse(t *testing.T) {
	t.Parallel()

	t.Run("with valid URIs",
		testLoop(rawParsePassTests),
	)

	t.Run("with URI structure checks",
		testLoop(rawParseStructureTests),
	)

	t.Run("with valid URI references",
		testLoop(rawParseReferenceTests),
	)

	t.Run("with schemes",
		testLoop(rawParseSchemeTests),
	)

	t.Run("with userinfo",
		testLoop(rawParseUserInfoTests),
	)

	t.Run("with hosts",
		testLoop(rawParseHostTests),
	)

	t.Run("with IP hosts",
		testLoop(rawParseIPHostTests),
	)

	t.Run("with ports",
		testLoop(rawParsePortTests),
	)

	t.Run("with paths",
		testLoop(rawParsePathTests),
	)

	t.Run("with queries",
		testLoop(rawParseQueryTests),
	)

	t.Run("with fragments",
		testLoop(rawParseFragmentTests),
	)

	t.Run("with other invalid URIs",
		testLoop(rawParseFailTests),
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

func TestValidatePath(t *testing.T) {
	u := authorityInfo{}
	for _, path := range []string{
		"/a/b/c",
		"a",
		"a/",
		"a/b/",
		"/",
		"",
		"www/詹姆斯/org/",
		"a//b//",
	} {
		require.NoErrorf(t, u.validatePath(path),
			"expected path %q to validate",
			path,
		)
	}

	for _, path := range []string{
		"/a/b{/c",
		"a{",
		"a{/",
		"a{/b/",
		"/{",
		"{",
		"www/詹{姆斯/org/",
	} {
		require.Errorf(t, u.validatePath(path),
			"expected path %q NOT to validate",
			path,
		)
	}
}

func TestValidateHostForScheme(t *testing.T) {
	for _, host := range []string{
		"a.b.c",
		"a",
		"a.b1b",
		"a.b2",
		"a.b.c.d",
		"a-b.c-d",
		"www.詹姆斯.org",
		"www.詹-姆斯.org",
		fmt.Sprintf("a.%s.c", strings.Repeat("b", 63)),
	} {
		require.NoErrorf(t, validateHostForScheme(host, host, "http"),
			"expected host %q to validate",
			host,
		)
	}

	for _, host := range []string{
		"a.b.c|",
		"a.b.c-",
		"a-",
		"a.",
		"a.b.",
		"a.1b",
		"a.2",
		"a.b.c..",
		".",
		"",
		"www.詹姆斯.org/",
		"www.詹{姆}斯.org/",
		fmt.Sprintf("a.%s.c", strings.Repeat("b", 64)),
	} {
		require.Errorf(t, validateHostForScheme(host, host, "http"),
			"expected host %q NOT to validate",
			host,
		)
	}
}

func testLoop(generator testGenerator) func(t *testing.T) {
	// table-driven tests for IsURI, IsURIReference, Parse and ParseReference.
	return func(t *testing.T) {
		for _, toPin := range generator() {
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
