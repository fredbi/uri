package uri

import (
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

type defautPortTest struct {
	uriRaw          string
	comment         string
	expectIsDefault bool
	expectedDefault int
}

func TestDefaultPorts(t *testing.T) {
	for _, toPin := range []defautPortTest{
		{
			uriRaw:          "http://host:80",
			comment:         "default http port",
			expectIsDefault: true,
			expectedDefault: 80,
		},
		{
			uriRaw:          "http://host:8080",
			comment:         "non-default http port",
			expectIsDefault: false,
			expectedDefault: 80,
		},
		{
			uriRaw:          "postgresql://host:6532",
			comment:         "non-default posgresql port",
			expectIsDefault: false,
			expectedDefault: 5432,
		},
	} {
		test := toPin
		u, err := Parse(test.uriRaw)
		require.NoError(t, err)

		t.Run(test.comment, func(t *testing.T) {
			t.Run("with IsDefaultPort", func(t *testing.T) {
				require.Equal(t, test.expectIsDefault, u.IsDefaultPort())
			})
			t.Run("with DefaultPort", func(t *testing.T) {
				require.Equal(t, test.expectedDefault, u.DefaultPort())
			})
		})
	}
}
