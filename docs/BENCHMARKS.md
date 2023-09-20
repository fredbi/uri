# Benchmarks

```sh
go test -v -run Bench -benchtime 30s -bench Bench
```

## v0.1

```
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
```

## After removal of regexp's

```
goos: linux
goarch: amd64
pkg: github.com/fredbi/uri
cpu: Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz
Benchmark_Parse
Benchmark_Parse-4            	42297637	       850.0 ns/op	     305 B/op	       6 allocs/op
Benchmark_ParseURLAsURI
Benchmark_ParseURLAsURI-4    	40728069	       892.0 ns/op	     317 B/op	       6 allocs/op
Benchmark_ParseURLStdLib
Benchmark_ParseURLStdLib-4   	80623975	       443.5 ns/op	     160 B/op	       1 allocs/op
Benchmark_String
Benchmark_String-4           	134472675	       270.5 ns/op	     142 B/op	       5 allocs/op
```

## After removal of internal slice allocs

```
goos: linux
goarch: amd64
pkg: github.com/fredbi/uri
cpu: Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz
Benchmark_Parse
Benchmark_Parse-4            	49388762	       722.5 ns/op	     284 B/op	       4 allocs/op
Benchmark_ParseURLAsURI
Benchmark_ParseURLAsURI-4    	46957615	       760.3 ns/op	     298 B/op	       5 allocs/op
Benchmark_ParseURLStdLib
Benchmark_ParseURLStdLib-4   	82265780	       439.2 ns/op	     160 B/op	       1 allocs/op
Benchmark_String
Benchmark_String-4           	140821515	       257.9 ns/op	     142 B/op	       5 allocs/op
```

## After removal of internal pointer

```
goos: linux
goarch: amd64
pkg: github.com/fredbi/uri
cpu: Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz
Benchmark_Parse
Benchmark_Parse-4            	51277591	       708.6 ns/op	     268 B/op	       3 allocs/op
Benchmark_ParseURLAsURI
Benchmark_ParseURLAsURI-4    	47649984	       759.2 ns/op	     282 B/op	       4 allocs/op
Benchmark_ParseURLStdLib
Benchmark_ParseURLStdLib-4   	81916470	       436.0 ns/op	     160 B/op	       1 allocs/op
Benchmark_String
Benchmark_String-4           	141247459	       256.4 ns/op	     142 B/op	       5 allocs/op
```

## After transform of strings.Split()

```
goos: linux
goarch: amd64
pkg: github.com/fredbi/uri
cpu: Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz
Benchmark_Parse
Benchmark_Parse-4            	62720376	       592.6 ns/op	     208 B/op	       2 allocs/op
Benchmark_ParseURLAsURI
Benchmark_ParseURLAsURI-4    	61047117	       654.0 ns/op	     208 B/op	       2 allocs/op
Benchmark_ParseURLStdLib
Benchmark_ParseURLStdLib-4   	67373106	       473.0 ns/op	     160 B/op	       1 allocs/op
Benchmark_String
Benchmark_String-4           	130358996	       269.6 ns/op	     142 B/op	       5 allocs/op
```

## New benchmark with different payloads

```
goos: linux
goarch: amd64
pkg: github.com/fredbi/uri
cpu: Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz
Benchmark_Parse
Benchmark_Parse/with_URI_simple_payload
Benchmark_Parse/with_URI_simple_payload-4         	58577457	       616.1 ns/op	     208 B/op	       2 allocs/op
Benchmark_Parse/with_URL_simple_payload
Benchmark_Parse/with_URL_simple_payload-4         	75833767	       471.7 ns/op	     168 B/op	       1 allocs/op
Benchmark_Parse/with_URI_mixed_payload
Benchmark_Parse/with_URI_mixed_payload-4          	57780166	       625.8 ns/op	     208 B/op	       2 allocs/op
Benchmark_Parse/with_URL_mixed_payload
Benchmark_Parse/with_URL_mixed_payload-4          	80009212	       449.0 ns/op	     163 B/op	       1 allocs/op
Benchmark_Parse/with_URI_payload_with_IPs
Benchmark_Parse/with_URI_payload_with_IPs-4       	56046999	       647.4 ns/op	     197 B/op	       1 allocs/op
Benchmark_Parse/with_URL_payload_with_IPs
Benchmark_Parse/with_URL_payload_with_IPs-4       	64658163	       538.2 ns/op	     176 B/op	       1 allocs/op
Benchmark_String
Benchmark_String-4                                	137858840	       261.5 ns/op	     142 B/op	       5 allocs/op
```

