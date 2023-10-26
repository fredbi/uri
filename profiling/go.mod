module github.com/fredbi/uri/profiling

go 1.21.3

// replace github.com/fredbi/uri => /home/fred/src/github.com/fredbi/uri
replace github.com/fredbi/uri => github.com/fredbi/uri v1.1.1-0.20231026093253-d1ef1b61c7b4

require (
	github.com/fredbi/uri v1.1.0
	github.com/pkg/profile v1.7.0
)

require (
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
)
