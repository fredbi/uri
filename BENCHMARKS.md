# Benchmarks

## v0.1

o test -v -run Bench -benchtime 30s -bench Bench
goos: linux
goarch: amd64
pkg: github.com/fredbi/uri
cpu: Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz
Benchmark_Parse
Benchmark_Parse-4            	 8891922	      4266 ns/op	     196 B/op	       2 allocs/op
Benchmark_ParseURLAsURI
Benchmark_ParseURLAsURI-4    	 8295577	      4359 ns/op	     202 B/op	       3 allocs/op
Benchmark_ParseURLStdLib
Benchmark_ParseURLStdLib-4   	81022693	       435.5 ns/op	     160 B/op	       1 allocs/op
Benchmark_String
Benchmark_String-4           	134486211	       264.6 ns/op	     142 B/op	       5 allocs/op
PASS

