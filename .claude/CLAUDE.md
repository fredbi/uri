# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an RFC 3986 compliant URI builder, parser, and validator for Go. It provides stricter conformance than the standard library's `net/url` package, supporting strict RFC validation for URIs and URI relative references.

**Key differentiators from `net/url`:**
- Stricter RFC 3986/3987 compliance
- Explicit validation errors with typed error handling
- Support for URI references (relative URIs without schemes)
- DNS hostname validation for well-known schemes
- Strict percent-encoding validation (must decode to valid UTF-8)
- IPv6 zone identifier support per RFC 6874
- IPvFuture address support

## Development Commands

### Testing
```bash
# Run all tests with race detection and coverage
go test -v -race -cover -coverprofile=coverage.out -covermode=atomic ./...

# Run a single test
go test -v -run TestName

# Run tests with specific Go version
go test -v ./...

# Run fuzzing (requires Go 1.18+)
go test -fuzz=FuzzParse -fuzztime=30s
```

### Linting
```bash
# Run golangci-lint (uses .golangci.yml config)
golangci-lint run

# Auto-fix issues where possible
golangci-lint run --fix
```

### Benchmarking
```bash
# Run benchmarks
go test -bench=. -benchmem

# Profile CPU usage
go test -bench=. -cpuprofile=cpu.prof

# Profile memory allocation
go test -bench=. -memprofile=mem.prof
```

### Vulnerability Scanning
```bash
# Check for known vulnerabilities
govulncheck ./...
```

## Architecture

### Core Types

**`URI` interface** (uri.go:24-54): The main public interface representing a parsed URI with methods to access components (Scheme, Authority, Query, Fragment) and transform it via a Builder.

**`uri` struct** (uri.go:309-319): Internal implementation holding raw components (scheme, hierPart, query, fragment) and parsed authority information.

**`Authority` interface** (uri.go:56-72): Represents authority information (userinfo, host, port, path) with validation and IP detection capabilities.

**`authorityInfo` struct** (uri.go:483-492): Internal implementation of Authority containing parsed components and IP type metadata.

**`Builder` interface** (builder.go:4-16): Fluent API for constructing/modifying URIs. The `uri` type implements this interface, allowing parsed URIs to be modified via `URI.Builder()`.

### Parsing Strategy

The parser (uri.go:135-307) uses a single-pass approach that:
1. Locates component boundaries by finding delimiter positions (':', '?', '#')
2. Validates delimiter ordering to reject malformed URIs early
3. Extracts raw components without regex
4. Parses authority separately (handles userinfo, host, port, path)
5. Validates all components before returning

**Key insight:** The parser differentiates between URI (`Parse`) and URI reference (`ParseReference`). References allow omitting the scheme (e.g., `//example.com/path`).

### Validation Architecture

Validation is separated by component with specialized validators:

- **Scheme validation** (uri.go:420-452): Must start with ASCII letter, followed by letters/digits/+/-/.
- **Host validation** (uri.go:626-683): Branches into IP address validation or DNS/registered-name validation based on scheme.
- **DNS validation** (dns.go:133-318): RFC 1035 compliant with segment length limits (63 bytes), total length limits (255 bytes), and proper character restrictions.
- **IPv4 validation** (ip.go:31-101): Custom parser ensuring dotted-decimal format without leading zeros or percent-encoding.
- **IPv6 validation** (ip.go:103-179): Leverages `netip.ParseAddr` with additional checks for zone identifiers (%25-prefixed).
- **Path/Query/Fragment validation** (uri.go:460-481): Uses `validateUnreservedWithExtra` to check character sets with appropriate extra allowed characters.

**Percent-encoding philosophy:** All percent-encoded sequences MUST decode to valid UTF-8 runes (stricter than RFC which says "should"). This prevents invalid UTF-8 propagation.

### DNS vs Registered Name

The `UsesDNSHostValidation` function (dns.go:18-131) determines whether a scheme requires strict DNS hostname validation or allows the more permissive RFC 3986 "registered name" format. This is a package-level variable that can be overridden.

**Well-known schemes** (http, https, ftp, etc.) use DNS validation. Others fall back to registered-name rules allowing percent-encoding and sub-delimiters.

### Error Handling

Typed errors (errors.go) use sentinel errors wrapped with context via `errorsJoin`. The module provides:
- `ErrURI`: Base sentinel error
- Component-specific errors: `ErrInvalidScheme`, `ErrInvalidHost`, `ErrInvalidPort`, etc.
- IP-specific errors: Custom `ipError` type for IPv4 validation failures

Errors support Go 1.20+ `errors.Join` with fallback for Go 1.19 (pre_go20.go).

### Character Set Validation

The `decode.go` file contains core character validation:
- `isUnreserved`: Letters, digits, and unreserved marks (-._~)
- `isSubDelim`: RFC 3986 sub-delimiters (!$&'()*+,;=)
- `validateUnreservedWithExtra`: Generic validator accepting base unreserved + extra runes
- `unescapePercentEncoding`: Validates and decodes %XX sequences ensuring valid UTF-8

### Performance Considerations

V1.1.0 achieved ~20x performance improvement by:
- Eliminating regex usage
- Reducing allocations via `strings.Builder` with pre-calculated capacity
- Single-pass parsing with early validation
- Avoiding repeated string operations

Current performance: 8-25% slower than `net/url.Parse` depending on workload, but significantly stricter validation.

## Testing Strategy

**Test organization:**
- `fixtures_test.go`: Massive test corpus from Perl, Python, Scala, .NET URI validators
- Component-specific tests: `dns_test.go`, `ip_test.go`, `decode_test.go`
- `fuzz_test.go`: Fuzzing support for discovering edge cases
- `benchmark_test.go`: Performance regression tracking
- `example_test.go`: Executable documentation

**When adding features:**
1. Add test cases to appropriate `*_test.go` file
2. Consider adding fuzz targets for new parsing code
3. Run benchmarks to ensure no performance regression
4. Add examples if the feature affects public API

## Code Style Notes

- The codebase uses explicit error wrapping with context messages
- Prefer early returns for error cases
- Use `strings.Builder` with `Grow()` for string concatenation
- Comments reference RFC sections extensively
- Constants are defined for magic characters (colonMark, slashMark, etc.)
- Avoid allocations in hot paths (validation loops)

## Known Limitations (from README)

- IRI validation lacks full strictness (unicode support is partial)
- IPv6 validation relies on standard library (not fully RFC compliant)
- Empty fragment/query handling needs refinement (e.g., `https://host?` vs `https://host`)
- URI normalization not yet implemented

## Go Version Support

- Minimum: Go 1.19
- Conditional compilation for Go 1.20+ features (errors.Join) via pre_go20.go/post_go20.go
