// go: !go1.20

package uri

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrUri(t *testing.T) {
	e := errorsJoin(ErrInvalidURI, errSentinelTest, fmt.Errorf("cause"))

	assert.True(t, errors.Is(ErrInvalidURI, ErrInvalidURI))
	assert.True(t, errors.Is(e, ErrInvalidURI))
	assert.True(t, errors.Is(e, errSentinelTest))
}
