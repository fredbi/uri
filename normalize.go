package uri

import (
	"errors"
	"fmt"
	"io"
	"net/netip"
	"net/url"
	"path"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/idna"
	"golang.org/x/text/unicode/norm"
)

type encodingContext uint8

const (
	encodingContextUserInfo encodingContext = iota
	encodingContextHost
	encodingContextIPv6
	encodingContextIPv6Zone
	encodingContextRelativePathFirstSegment
	encodingContextPathSegment
	encodingContextQuery
	encodingContextFragment
)

// Normalize yields a canonicalized representation of the URI.
//
// See https://en.wikipedia.org/wiki/URI_normalization
func (u uri) Normalize(opts ...NormalizeOption) (string, error) {
	n, err := u.Normalized(opts...)
	if err != nil {
		return "", err
	}

	return n.String(), nil
}

// Normalized yields a new URI with normalized content.
//
// Calling String() on that one would produce the same string as calling
// Normalize() on the original URI.
func (u uri) Normalized(opts ...NormalizeOption) (URI, error) { // TODO: include UTF8 percent-encoding check in validation
	o := normalizeOptionsWithDefaults(opts)
	scheme := normalizedScheme(u.scheme, o)
	query, err := normalizedQuery(u.query, o)
	if err != nil {
		return nil, err
	}

	fragment, err := normalizedFragment(u.fragment, o)
	if err != nil {
		return nil, err
	}

	userinfo, err := normalizedUserinfo(u.authority.userinfo, o)
	if err != nil {
		return nil, err
	}

	// TODO: add more info to the uri structure to avoid this
	var host string
	unescapedHost, err := url.PathUnescape(u.authority.host)
	addr, err := netip.ParseAddr(unescapedHost) // NOTE: in validation, we accept percent-encode and we should not
	isIPv4 := err == nil && addr.Is4()
	switch {
	case u.authority.isIPv6 || isIPv4:
		host = addr.String() // is this correct when empty/zero address?
	default:
		host, err = normalizedHost(u.authority.host, o)
		if err != nil {
			return nil, err
		}
	}

	port, err := normalizedPort(u.authority.port, scheme, o)
	if err != nil {
		return nil, err
	}

	pth, err := normalizedPath(u.authority.path, o)
	if err != nil {
		return nil, err
	}

	return &uri{
		scheme: scheme,
		authority: authorityInfo{
			prefix:   authorityPrefix,
			userinfo: userinfo,
			host:     host,
			port:     port,
			path:     pth,
		},
		query:    query,
		fragment: fragment,
	}, nil
}

func normalizedScheme(scheme string, o *normalizeOptions) string {
	// DONE
	if len(scheme) == 0 {
		return ""
	}

	return strings.ToLower(scheme)
}

func normalizedPath(pth string, o *normalizeOptions) (string, error) {
	// TODO: perf
	if len(pth) == 0 {
		return "/", nil
	}

	normalized := path.Clean(pth)

	segments := strings.Split(normalized, "/")
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		normalizedPart, _ := url.PathUnescape(segment)
		normalizedPart = url.PathEscape(normalizedPart)

		parts = append(parts, normalizedPart)
	}

	return strings.Join(parts, "/"), nil
}

func normalizedUserinfo(userinfo string, o *normalizeOptions) (string, error) {
	// TODO: perf
	normalized, _ := url.PathUnescape(userinfo)
	normalized = url.PathEscape(normalized)

	return normalized, nil
}

func normalizedPort(port, scheme string, o *normalizeOptions) (string, error) {
	if len(port) == 0 {
		return "", nil
	}

	portNum, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		return "", err
	}

	if portNum == 0 || defaultPortForScheme(scheme) == portNum { //TODO: leave default port option
		return "", nil
	}

	return strconv.FormatUint(portNum, 10), nil
}

func normalizedHost(host string, o *normalizeOptions) (string, error) {
	normalized := strings.ToLower(host)

	var err error
	normalized, err = normalizedPercentEncoding(normalized, encodingContextHost, o)
	if err != nil {
		return "", err
	}

	// normalized = width.Fold.String(normalized) // redundant?? this is what purell does
	normalized = norm.NFC.String(normalized)
	if o.asciiHost {
		normalized, _ = idna.ToASCII(normalized) // convert to puny code
	}

	return normalized, nil
}

func normalizedQuery(query string, o *normalizeOptions) (string, error) {
	normalized, err := normalizedPercentEncoding(query, encodingContextQuery, o)
	if err != nil {
		return "", err
	}

	return norm.NFC.String(normalized), nil
}

func normalizedFragment(fragment string, o *normalizeOptions) (string, error) {
	// TODO: test http://example.com/index.html#!s3!search terms
	//http://example.com/data.csv#row=4
	normalized, err := normalizedPercentEncoding(fragment, encodingContextFragment, o)
	if err != nil {
		return "", err
	}

	return norm.NFC.String(normalized), nil
}

// normalizedPercentEncoding returns a string with no extraneous % encoded chars.
// In addition, percent encoding ensures that the hex digits are always upper cased.
//
// Notice that the notion of "extraneous" depends on the context for this string.
func normalizedPercentEncoding(s string, uriContext encodingContext, o *normalizeOptions) (string, error) {
	var normalized strings.Builder
	normalized.Grow(len(s))
	skip := 0

	for i, r := range s {
		if skip > 0 {
			skip--

			continue
		}

		if r == '%' {
			// TODO: factorize this
			// percent-encoded sequence
			skip = 2
			offset := i
			if len(s) <= i+skip {
				return "", errors.Join(
					ErrInvalidEscaping, // TODO: this should be ensured by validation
					fmt.Errorf("expected escaping '%%' to be followed by 2 hex digits, near: %q", s[i:]),
				)
			}

			var codePoint [utf8.UTFMax]byte

			// superfluous encoding
			escapeSequence := [2]byte{
				s[offset+1],
				s[offset+2],
			}

			codePoint[0] = unescapeCodePoint(escapeSequence)
			codePointLength := 1

			// escaped utf8 sequence
			if codePoint[0] > 0b11000000 {
				skip += 3
				offset += 3
				codePointLength++
				// expect another escaped sequence

				if len(s) <= offset {
					return "", errors.Join(
						ErrInvalidEscaping, // TODO: this should be ensured by validation
						fmt.Errorf("expected rune (at least 2 bytes) to be encoded with an additional percent-escaped byte at %q", s[i:]),
					)
				}

				if s[offset] != '%' {
					return "", errors.Join(
						ErrInvalidEscaping, // TODO: this should be ensured by validation
						fmt.Errorf("expected rune (at least 2 bytes) to be encoded with an additional percent-escaped byte but got %q at %q", s[offset], s[i:]),
					)
				}

				escapeSequence = [2]byte{
					s[offset+1],
					s[offset+2],
				}

				codePoint[1] = unescapeCodePoint(escapeSequence)

				if codePoint[0] > 0b11100000 {
					skip += 3
					offset += 3
					codePointLength++

					// expect yet another escaped sequence
					if len(s) <= offset {
						return "", errors.Join(
							ErrInvalidEscaping, // TODO: this should be ensured by validation
							fmt.Errorf("expected rune (at least 3 bytes) to be encoded with an additional percent-escaped byte at %q", s[i:]),
						)
					}

					if s[offset] != '%' {
						return "", errors.Join(
							ErrInvalidEscaping, // TODO: this should be ensured by validation
							fmt.Errorf("expected rune (at least 3 bytes) to be encoded with an additional percent-escaped byte but got %q at %q", s[offset], s[i:]),
						)
					}

					escapeSequence = [2]byte{
						s[offset+1],
						s[offset+2],
					}

					codePoint[2] = unescapeCodePoint(escapeSequence)

					if codePoint[0] > 0b11110000 {
						skip += 3
						offset += 3
						codePointLength++

						if len(s) <= offset {
							return "", errors.Join(
								ErrInvalidEscaping, // TODO: this should be ensured by validation
								fmt.Errorf("expected rune (at least 4 bytes) to be encoded with an additional percent-escaped byte at %q", s[i:]),
							)
						}

						if s[offset] != '%' {
							return "", errors.Join(
								ErrInvalidEscaping, // TODO: this should be ensured by validation
								fmt.Errorf("expected rune (at least 4 bytes) to be encoded with an additional percent-escaped byte but got %q at %q", s[offset], s[i:]),
							)
						}

						escapeSequence = [2]byte{
							s[offset+1],
							s[offset+2],
						}

						codePoint[3] = unescapeCodePoint(escapeSequence)
					}
				}
			}

			unescapedRune, _ := utf8.DecodeRune(codePoint[:codePointLength])
			if unescapedRune == utf8.RuneError {
				return "", errors.Join(
					ErrInvalidEscaping,
					fmt.Errorf("the escaped code points do not add up to a valid rune near: %q", s[i:]),
				)
			}

			if uriContext == encodingContextHost {
				unescapedRune = unicode.ToLower(unescapedRune)
			}

			if !shouldEscape(unescapedRune, uriContext, o) {
				// extraneous escape detected
				normalized.WriteRune(unescapedRune)

				continue
			}

			// escape is legit, ensure upper case hex encoding of the canonical UTF-8 representation
			writeEscapedRune(&normalized, unescapedRune)

			continue
		}

		if shouldEscape(r, uriContext, o) {
			writeEscapedRune(&normalized, r)

			continue
		}

		normalized.WriteRune(r)
	}

	return normalized.String(), nil
}

func unescapeCodePoint(escapeSequence [2]byte) byte {
	return unhex(escapeSequence[0])<<4 | unhex(escapeSequence[1])
}

func writeEscapedRune(w io.Writer, r rune) {
	// escape a rune with percent encoding (resulting into up to 4 %-encoded sequences)
	const upperhex = "0123456789ABCDEF"
	var in [utf8.UTFMax]byte
	n := utf8.EncodeRune(in[:], r)

	for byteInRune := 0; byteInRune < n; byteInRune++ {
		c := in[byteInRune]
		var out [3]byte
		out[0] = '%'
		out[1] = upperhex[c>>4]
		out[2] = upperhex[c&15]
		_, _ = w.Write(out[:])
	}
}

func shouldEscape(r rune, uriContext encodingContext, o *normalizeOptions) bool {
	// TODO: make more readable using functions...

	if 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9' { // ALPHA / DIGIT
		return false
	}

	// analyse various separators
	switch r {
	case ' ', '\t', '\n', '\r':
		return true // see below for unicode blanks
	case '%':
		return true
	case '-', '.', '_', '~': // §2.3 Unreserved characters
		return false
	case ':', '/', '?', '#', '[', ']', '@', // gen-delims
		'!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=': // sub-delims
		switch uriContext {
		case encodingContextRelativePathFirstSegment: // segment-nz-nc
			switch r {
			case ':', '/', '?', '#', '[', ']': // escape gen-delims except @
				return true
			default: // don't escape sub-delims
				return false
			}
		case encodingContextPathSegment:
			switch r {
			case '/', '?', '#', '[', ']': // escape gen-delims except :, @
				return true
			default: // don't escape sub-delims
				return false
			}
		case encodingContextQuery, encodingContextFragment:
			switch r {
			case '#', '[', ']': // escape gen-delims except : @ / and ?
				return true
			default:
				return false
			}
		case encodingContextHost:
			switch r {
			case ':', '/', '?', '#', '[', ']', '@': // escape gen-delims
				return true
			default: // don't escape sub-delims
				return false
			}
		case encodingContextUserInfo:
			switch r {
			case '/', '?', '#', '[', ']', '@': // escape gen-delims except :
				return true
			default:
				return false
			}
		case encodingContextIPv6:
			return false
		case encodingContextIPv6Zone:
			return false
		default:
			panic("invalid encodingContext")
			return false
		}
	}

	if unicode.IsSpace(r) {
		return true
	}

	// escape unicode?
	if o.escapeUnicode {
		// this makes an URI RFC-compliant URI, as stated RCF3986 §2.5
		return true
	}

	// this makes an IRI RFC-compliant URI
	if isUcsChar(r) {
		return false
	}

	if uriContext == encodingContextQuery && isIPrivate(r) {
		return false
	}

	return true
}

var (
	ucschar = &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x000A0, 0x0D7FF, 1},
			{0x0F900, 0x0FDCF, 1},
			{0x0FDF0, 0x0FFEF, 1},
		},
		R32: []unicode.Range32{
			{0x10000, 0x1FFFD, 1},
			{0x20000, 0x2FFFD, 1},
			{0x30000, 0x3FFFD, 1},
			{0x40000, 0x4FFFD, 1},
			{0x50000, 0x5FFFD, 1},
			{0x60000, 0x6FFFD, 1},
			{0x70000, 0x7FFFD, 1},
			{0x80000, 0x8FFFD, 1},
			{0x90000, 0x9FFFD, 1},
			{0xA0000, 0xAFFFD, 1},
			{0xB0000, 0xBFFFD, 1},
			{0xC0000, 0xCFFFD, 1},
			{0xD0000, 0xDFFFD, 1},
			{0xE1000, 0xEFFFD, 1},
		},
	}

	iprivate = &unicode.RangeTable{
		R16: []unicode.Range16{
			{0xE000, 0xF8FF, 1},
		},
		R32: []unicode.Range32{
			{0xF0000, 0xFFFFD, 1},
			{0x100000, 0x10FFFD, 1},
		},
	}
)

func isUcsChar(r rune) bool {
	return unicode.In(r, ucschar)
}

func isIPrivate(r rune) bool {
	return unicode.In(r, iprivate)
}
