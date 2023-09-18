//go:build profiling

package uri

import (
	"testing"

	"github.com/pkg/profile"
	"github.com/stretchr/testify/require"
)

func TestParseWithProfile(t *testing.T) {
	const (
		profDir = "prof"
		n       = 1000
	)

	t.Run("collect CPU profile", func(t *testing.T) {
		defer profile.Start(
			profile.CPUProfile,
			profile.ProfilePath(profDir),
			profile.NoShutdownHook,
		).Stop()

		// current: Parse calls total CPU: 100ms -> 70ms
		// validateHost: 30ms -> 10ms
		// validatePath: 20ms -> 20ms (same, less gc work) -> 10ms
		// validatePort: 10ms
		runProfile(t, n)
	})

	t.Run("collect memory profile", func(t *testing.T) {
		defer profile.Start(
			profile.MemProfile,
			profile.ProfilePath(profDir),
			profile.NoShutdownHook,
		).Stop()

		// current: object allocs: 653 746 -> 533 849 -> 505 606
		runProfile(t, n)
	})
}

func runProfile(t *testing.T, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		for _, generator := range allGenerators {
			for _, testCase := range generator() {
				if testCase.isReference || testCase.err != nil {
					// skip URI references and invalid cases
					continue
				}

				u, err := Parse(testCase.uriRaw)
				require.NoErrorf(t, err, "unexpected error for %q", testCase.uriRaw)
				require.NotEmpty(t, u)
			}
		}
	}
}
