package main

import (
	"log"

	"github.com/fredbi/uri"
	"github.com/fredbi/uri/profiling/fixtures"
	"github.com/pkg/profile"
)

const (
	profDir = "prof"
)

func main() {
	const (
		n = 100000
	)

	profileCPU(n)
	profileMemory(n)
}

func profileCPU(n int) {
	defer profile.Start(
		profile.CPUProfile,
		profile.ProfilePath(profDir),
		profile.NoShutdownHook,
	).Stop()

	// current: Parse calls total CPU: 100ms -> 70ms
	// validateHost: 30ms -> 10ms
	// validatePath: 20ms -> 20ms (same, less gc work) -> 10ms
	// validatePort: 10ms
	runProfile(n)
}

func profileMemory(n int) {
	defer profile.Start(
		profile.MemProfile,
		profile.ProfilePath(profDir),
		profile.NoShutdownHook,
	).Stop()

	// current: object allocs: 653 746 -> 533 849 -> 505 606
	runProfile(n)
}

func runProfile(n int) {
	for i := 0; i < n; i++ {
		for _, generator := range fixtures.AllGenerators {
			for _, testCase := range generator() {
				if testCase.IsReference || testCase.Err != nil {
					// skip URI references and invalid cases
					continue
				}

				u, err := uri.Parse(testCase.URIRaw)
				if u == nil || err != nil {
					log.Fatalf("unexpected error for %q", testCase.URIRaw)
				}
			}
		}
	}
}
