package uri

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzParse(f *testing.F) {
	allTests := rawParsePassTests()
	allTests = append(allTests, rawParseFailTests()...)
	for _, testCase := range allTests {
		f.Add(testCase.uriRaw)
	}

	f.Fuzz(func(t *testing.T, input string) {
		require.NotPanics(t, func() {
			_, _ = Parse(input)
		})
	})
}
