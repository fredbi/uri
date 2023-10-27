package uri

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func validateUnreservedWithExtra(s string, acceptedRunes []rune) error {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError {
			return errorsJoin(ErrInvalidEscaping,
				fmt.Errorf("invalid UTF8 rune near: %q", s[i:]),
			)
		}
		i += size

		// accepts percent-encoded sequences, but only if they correspond to a valid UTF-8 encoding
		if r == percentMark {
			if i >= len(s) {
				return errorsJoin(ErrInvalidEscaping,
					fmt.Errorf("incomplete escape sequence"),
				)
			}

			_, offset, err := unescapePercentEncoding(s[i:])
			if err != nil {
				return errorsJoin(ErrInvalidEscaping, err)
			}

			i += offset

			continue
		}

		// RFC grammar definitions:
		// sub-delims  = "!" / "$" / "&" / "'" / "(" / ")"
		//               / "*" / "+" / "," / ";" / "="
		// gen-delims  = ":" / "/" / "?" / "#" / "[" / "]" / "@"
		// unreserved    = ALPHA / DIGIT / "-" / "." / "_" / "~"
		// pchar         = unreserved / pct-encoded / sub-delims / ":" / "@"
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
			// unreserved
			r != '-' && r != '.' && r != '_' && r != '~' &&
			// sub-delims
			r != '!' && r != '$' && r != '&' && r != '\'' && r != '(' && r != ')' &&
			r != '*' && r != '+' && r != ',' && r != ';' && r != '=' {
			runeFound := false
			for _, acceptedRune := range acceptedRunes {
				if r == acceptedRune {
					runeFound = true
					break
				}
			}

			if !runeFound {
				return fmt.Errorf("contains an invalid character: '%U' (%q) near %q", r, r, s[i:])
			}
		}
	}

	return nil
}

func unescapePercentEncoding(s string) (rune, int, error) {
	var (
		offset          int
		codePoint       [utf8.UTFMax]byte
		codePointLength int
		err             error
	)

	if codePoint[0], err = unescapeSequence(s); err != nil {
		return utf8.RuneError, 0, err
	}

	codePointLength++
	offset += 2

	// escaped utf8 sequence
	if codePoint[0] >= 0b11000000 {
		// expect another escaped sequence
		if offset >= len(s) {
			return 0, 0, fmt.Errorf("expected a '%%' escape character, near: %q", s)
		}

		if s[offset] != '%' {
			return 0, 0, fmt.Errorf("expected a '%%' escape character, near: %q", s[offset:])
		}
		offset++

		if codePoint[1], err = unescapeSequence(s[offset:]); err != nil {
			return utf8.RuneError, 0, err
		}

		codePointLength++
		offset += 2

		if codePoint[0] >= 0b11100000 {
			// expect yet another escaped sequence
			if offset >= len(s) {
				return 0, 0, fmt.Errorf("expected a '%%' escape character, near: %q", s)
			}

			if s[offset] != '%' {
				return 0, 0, fmt.Errorf("expected a '%%' escape character, near: %q", s[offset:])
			}
			offset++

			if codePoint[2], err = unescapeSequence(s[offset:]); err != nil {
				return utf8.RuneError, 0, err
			}
			codePointLength++
			offset += 2

			if codePoint[0] >= 0b11110000 {
				// expect a fourth escaped sequence
				if offset >= len(s) {
					return 0, 0, fmt.Errorf("expected a '%%' escape character, near: %q", s)
				}

				if s[offset] != '%' {
					return 0, 0, fmt.Errorf("expected a '%%' escape character, near: %q", s[offset:])
				}
				offset++

				if codePoint[3], err = unescapeSequence(s[offset:]); err != nil {
					return utf8.RuneError, 0, err
				}
				codePointLength++
				offset += 2
			}
		}
	}

	unescapedRune, _ := utf8.DecodeRune(codePoint[:codePointLength])
	if unescapedRune == utf8.RuneError {
		return utf8.RuneError, 0, fmt.Errorf("the escaped code points do not add up to a valid rune")
	}

	return unescapedRune, offset, nil
}

func unescapeSequence(escapeSequence string) (byte, error) {
	if len(escapeSequence) < 2 {
		return 0, fmt.Errorf("expected escaping '%%' to be followed by 2 hex digits, near: %q", escapeSequence)
	}

	if !isHex(escapeSequence[0]) || !isHex(escapeSequence[1]) {
		return 0, fmt.Errorf("part contains a malformed percent-encoded hex digit, near: %q", escapeSequence)
	}

	return unhex(escapeSequence[0])<<4 | unhex(escapeSequence[1]), nil
}

func isHex[T byte | rune](c T) bool {
	switch {
	case isDigit(c):
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	default:
		return false
	}
}

func isNotDigit[T rune | byte](r T) bool {
	return r < '0' || r > '9'
}

func isDigit[T rune | byte](r T) bool {
	return r >= '0' && r <= '9'
}

func isASCIILetter[T byte | rune](c T) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isNumerical(input string) bool {
	return strings.IndexFunc(input, isNotDigit[rune]) == -1
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
