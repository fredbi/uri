# uri
![Lint](https://github.com/fredbi/uri/actions/workflows/01-golang-lint.yaml/badge.svg)
![CI](https://github.com/fredbi/uri/actions/workflows/02-test.yaml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/fredbi/uri/badge.svg?branch=master)](https://coveralls.io/github/fredbi/uri?branch=master)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/fredbi/uri)
[![Go Reference](https://pkg.go.dev/badge/github.com/fredbi/uri.svg)](https://pkg.go.dev/github.com/fredbi/uri)
[![license](http://img.shields.io/badge/license/License-MIT-yellow.svg)](https://raw.githubusercontent.com/fredbi/uri/master/LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/fredbi/uri)](https://goreportcard.com/report/github.com/fredbi/uri)

Package uri is meant to be an RFC 3986 compliant URI builder, parser and validator for `golang`.

It supports strict RFC validation for URI and URI relative references.

This allows for stricter conformance than the `net/url` package in the Go standard libary,
which provides a workable but loose implementation of the RFC for URLs.

## Usage

### Parsing

```go
	u, err := Parse("https://example.com:8080/path")
	if err != nil {
		fmt.Printf("Invalid URI")
	} else {
		fmt.Printf("%s", u.Scheme())
	}
	// Output: https
```

```go
	u, err := ParseReference("//example.com/path")
	if err != nil {
		fmt.Printf("Invalid URI reference")
	} else {
		fmt.Printf("%s", u.Authority().Path())
	}
	// Output: /path
```

### Validation

```go
    isValid := IsURI("urn://example.com?query=x#fragment/path") // true

    isValid= IsURI("//example.com?query=x#fragment/path") // false

    isValid= IsURIReference("//example.com?query=x#fragment/path") // true
```

#### Caveats

* Registered name vs DNS name: RFC3986 defines a super-permissive "registered name" for hosts, for URIs
  not specifically related to an Internet name. Our validation performs a stricter host validation according
  to DNS rules whenever the scheme is a well-known IANA-registered scheme (this function is customizable).
* IPv6 validation relies on IPv6 parsing functionality from the standard library. It is not super strict
  on the IPv6 specification.

### Building

The exposed type `URI` can be transformed into a fluent `Builder` to set the parts of an URI.

```go
	aURI, _ := Parse("mailto://user@domain.com")
	newURI := auri.Builder().SetUserInfo(test.name).SetHost("newdomain.com").SetScheme("http").SetPort("443")
```

### Canonicalization

Not supported for now.

For url normalization, see [PuerkitoBio/purell](https://github.com/PuerkitoBio/purell).

## Reference specifications

The librarian's corner (WIP).

|----------------------------------|----------------------------------------------------|----------|
| Uniform Resource Identifier (URI)| [RFC3986](https://www.rfc-editor.org/rfc/rfc3986)  | |
| Uniform Resource Locator (URL)   | [RFC1738](https://www.rfc-editor.org/info/rfc1738) | |
| Relative URL                     | [RFC1808](https://www.rfc-editor.org/info/rfc1808) | |
| Internationalized Resource Identifier (IRI) | [RFC3987](https://tools.ietf.org/html/rfc3987)| |
| IPv6 addressing scheme reference and erratum|||
| Representing IPv6 Zone Identifiers| [RFC6874](https://www.rfc-editor.org/rfc/rfc6874.txt)|||
| - https://tools.ietf.org/html/rfc6874 |||
| - https://www.rfc-editor.org/rfc/rfc3513 |||

## Disclaimer

Not supported:
- provisions for "IPvFuture" are not implemented.

Hostnames vs domain names:
- a list of common schemes triggers the validation of hostname against domain name rules.

> Example:
> * ftp://host, http://host default to validating a proper DNS hostname.

## [FAQ](docs/FAQ.md)

## [Benchmarks](docs/BENCHMARKS.md)

## Credits

Tests have been aggregated from test suites of URI validators from other languages:
Perl, Python, Scala, .Net. and the Go url standard library.

> This package was initially based on the work from ttacon/uri (credits: Trey Tacon).
> Extra features like MySQL URIs present in the original repo have been removed.

A lot of improvements have been brought by the incredible guys at [`fyne-io`](https://github.com/fyne-io). Thanks all.

## TODOs

- [ ] revisit URI vs IRI support & strictness
- [ ] support IRI ucschar as unreserved characters

ucschar        = %xA0-D7FF / %xF900-FDCF / %xFDF0-FFEF
                  / %x10000-1FFFD / %x20000-2FFFD / %x30000-3FFFD
                  / %x40000-4FFFD / %x50000-5FFFD / %x60000-6FFFD
                  / %x70000-7FFFD / %x80000-8FFFD / %x90000-9FFFD
                  / %xA0000-AFFFD / %xB0000-BFFFD / %xC0000-CFFFD
                  / %xD0000-DFFFD / %xE1000-EFFFD
- [ ] support IRI iprivate in query
iprivate       = %xE000-F8FF / %xF0000-FFFFD / %x100000-10FFFD

		// TODO: RFC6874
		//  A <zone_id> SHOULD contain only ASCII characters classified as
   		// "unreserved" for use in URIs [RFC3986].  This excludes characters
   		// such as "]" or even "%" that would complicate parsing.  However, the
   		// syntax described below does allow such characters to be percent-
   		// encoded, for compatibility with existing devices that use them.
