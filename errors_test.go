package uri

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// errSentinelTest is used in test cases for when we want to assert an error
// but do not want to check specifically which error was returned.
var errSentinelTest = Error(errors.New("test")) //nolint:err113 // it is okay for a test error

func TestIPError(t *testing.T) {
	for _, e := range []error{
		errInvalidCharacter,
		errValueGreater255,
		errAtLeastOneDigit,
		errLeadingZero,
		errTooLong,
		errTooShort,
	} {
		require.NotEmpty(t, e.Error())
	}

	const invalidValue uint8 = 255
	require.Empty(t, ipError(invalidValue).Error())
}

func TestErr(t *testing.T) {
	t.Run("with valid URIs", func(t *testing.T) {
		for _, toPin := range rawParsePassTests() {
			test := toPin
			u, err := Parse(test.uriRaw)
			require.NoErrorf(t, err, "in testcase: %s (%q)", test.comment, test.uriRaw)
			require.Equal(t, err, u.Err(), "in testcase: %s (%q)", test.comment, test.uriRaw)

			if !errors.Is(err, ErrInvalidQuery) && !errors.Is(err, ErrInvalidScheme) && !errors.Is(err, ErrInvalidURI) {
				require.Equal(t, err, u.Authority().Err(), "in testcase: %s (%q)", test.comment, test.uriRaw)
			}
		}
	})
	t.Run("with invalid URIs", func(t *testing.T) {
		for _, toPin := range rawParseFailTests() {
			test := toPin
			u, err := Parse(test.uriRaw)
			require.Errorf(t, err, "in testcase: %s (%q)", test.comment, test.uriRaw)
			require.Equal(t, err, u.Err(), "in testcase: %s (%q)", test.comment, test.uriRaw)

			if !errors.Is(err, ErrInvalidQuery) && !errors.Is(err, ErrInvalidScheme) && !errors.Is(err, ErrInvalidURI) {
				require.Equal(t, err, u.Authority().Err(), "in testcase: %s (%q)", test.comment, test.uriRaw)
			}
		}
	})
}
