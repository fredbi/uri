package uri

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// errSentinelTest is used in test cases for when we want to assert an error
// but do not want to check specifically which error was returned.
var errSentinelTest = Error(errors.New("test"))

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
