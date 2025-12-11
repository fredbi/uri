// go: !go1.20

package uri

import (
	"errors"
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

func TestErrUri(t *testing.T) {
	e := errorsJoin(ErrInvalidURI, errSentinelTest, errors.New("cause")) //nolint: err113 // it is okay for a test error

	require.ErrorIs(t, e, ErrInvalidURI)
	require.ErrorIs(t, e, errSentinelTest)
}
