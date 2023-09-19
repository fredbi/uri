package uri

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzParse(f *testing.F) {
	for _, generator := range allGenerators {
		for _, testCase := range generator() {
			f.Add(testCase.uriRaw)
		}
	}

	f.Fuzz(func(t *testing.T, input string) {
		require.NotPanics(t, func() {
			_, _ = Parse(input)
		})
	})
}
