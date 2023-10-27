// Package uri is meant to be an RFC 3986 compliant URI builder and parser.
//
// This is based on the work from ttacon/uri (credits: Trey Tacon).
//
// This fork concentrates on RFC 3986 strictness for URI parsing and validation.
//
// Reference: https://tools.ietf.org/html/rfc3986
//
// Tests have been augmented with test suites of URI validators in other languages:
// perl, python, scala, .Net.
//
// Extra features like MySQL URIs present in the original repo have been removed.
package uri

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	// char and string literals.
	colonMark          = ':'
	questionMark       = '?'
	fragmentMark       = '#'
	percentMark        = '%'
	atHost             = '@'
	slashMark          = '/'
	openingBracketMark = '['
	closingBracketMark = ']'
	dotSeparator       = '.'
	authorityPrefix    = "//"
)

const (
	// DNS name constants
	maxSegmentLength = 63
	maxDomainLength  = 255
)

var (
	// predefined sets of accecpted runes beyond the "unreserved" character set
	pcharExtraRunes           = []rune{colonMark, atHost} // pchar = unreserved | ':' | '@'
	queryOrFragmentExtraRunes = append(pcharExtraRunes, slashMark, questionMark)
	userInfoExtraRunes        = append(pcharExtraRunes, colonMark)
)

type (
	// URI represents a general RFC3986 URI.
	URI struct {
		// raw components
		err      error
		scheme   string
		hierPart string
		query    string
		fragment string

		// parsed components
		authority Authority
	}

	// Authority information that a URI contains
	// as specified by RFC3986.
	//
	// Username and password are given by UserInfo().
	Authority struct {
		err      error
		prefix   string
		userinfo string
		host     string
		port     string
		path     string
		ipType   // after host validation, the IP type is more precisely identified
	}
)

// IsURI tells if a URI is valid according to RFC3986/RFC397.
func IsURI(raw string, opts ...Option) bool {
	_, err := Parse(raw, opts...)
	return err == nil
}

// IsURIReference tells if a URI reference is valid according to RFC3986/RFC397
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-4.1 and
// https://www.rfc-editor.org/rfc/rfc3986#section-4.2
func IsURIReference(raw string, opts ...Option) bool {
	_, err := ParseReference(raw, opts...)
	return err == nil
}

// Parse attempts to parse a URI.
//
// It returns an error if the URI is not RFC3986-compliant.
func Parse(raw string, opts ...Option) (URI, error) {
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	return parse(raw, o)
}

// ParseReference attempts to parse a URI relative reference.
//
// It returns an error if the URI is not RFC3986-compliant.
//
// Notice that this call is syntactically equivalent to Parse(raw, WithURIReference(true)),
// but slightly more efficient.
func ParseReference(raw string, opts ...Option) (URI, error) {
	o, redeem := applyURIReferenceOptions(opts)
	defer func() { redeem(o) }()

	return parse(raw, o)
}

func parse(raw string, o *options) (URI, error) {
	var (
		scheme string
		curr   int
	)

	schemeEnd := strings.IndexByte(raw, colonMark)      // position of a ":"
	hierPartEnd := strings.IndexByte(raw, questionMark) // position of a "?"
	queryEnd := strings.IndexByte(raw, fragmentMark)    // position of a "#"

	// exclude pathological input
	if schemeEnd == 0 || hierPartEnd == 0 || queryEnd == 0 {
		// ":", "?", "#"
		err := errorsJoin(
			ErrInvalidURI,
			fmt.Errorf("URI cannot start by a '%q', '%q' or '%q' mark", colonMark, questionMark, fragmentMark),
		)
		return URI{err: err}, err
	}

	if schemeEnd == 1 {
		err := errorsJoin(
			ErrInvalidScheme,
			fmt.Errorf("scheme has a minimum length of 2 characters"),
		)
		return URI{err: err}, err
	}

	if hierPartEnd == 1 || queryEnd == 1 {
		// ".:", ".?", ".#"
		err := errorsJoin(
			ErrInvalidURI,
			fmt.Errorf("invalid combination of start markers, near: %q", raw[:2]),
		)
		return URI{err: err}, err
	}

	if hierPartEnd > 0 && hierPartEnd < schemeEnd || queryEnd > 0 && queryEnd < schemeEnd {
		// e.g. htt?p: ; h#ttp: ..
		mini, maxi := miniMaxi(hierPartEnd, schemeEnd, queryEnd, schemeEnd)
		err := errorsJoin(
			ErrInvalidURI,
			fmt.Errorf("URI part markers %q,%q,%q are in an incorrect order, near: %q", colonMark, questionMark, fragmentMark, raw[mini:maxi]),
		)
		return URI{err: err}, err
	}

	if queryEnd > 0 && queryEnd < hierPartEnd {
		// e.g.  https://abc#a?b
		hierPartEnd = queryEnd
	}

	isRelative := strings.HasPrefix(raw, authorityPrefix)
	switch {
	case schemeEnd > 0 && !isRelative:
		scheme = raw[curr:schemeEnd]
		if schemeEnd+1 == len(raw) {
			// trailing ':' (e.g. http:)
			u := URI{
				scheme: scheme,
			}

			u.authority.ipType, u.err = u.validate(o)

			return u, u.err
		}
	case !o.withURIReference:
		// scheme is required for URI
		err := errorsJoin(
			ErrNoSchemeFound,
			fmt.Errorf("for URI (not URI reference), the scheme is required"),
		)
		return URI{err: err}, err
	case isRelative:
		// scheme is optional for URI references.
		//
		// start with // and a ':' is following... e.g //example.com:8080/path
		schemeEnd = -1
	}

	curr = schemeEnd + 1

	if hierPartEnd == len(raw)-1 || (hierPartEnd < 0 && queryEnd < 0) {
		// trailing ? or (no query & no fragment)
		if hierPartEnd < 0 {
			hierPartEnd = len(raw)
		}

		authority, err := parseAuthority(raw[curr:hierPartEnd], o)
		if err != nil {
			err = errorsJoin(ErrInvalidURI, err)
			return URI{err: err}, err
		}

		u := URI{
			scheme:    scheme,
			hierPart:  raw[curr:hierPartEnd],
			authority: authority,
		}

		u.authority.ipType, u.err = u.validate(o)
		u.authority.err = u.err

		return u, u.err
	}

	var (
		hierPart, query, fragment string
		authority                 Authority
		err                       error
	)

	if hierPartEnd > 0 {
		hierPart = raw[curr:hierPartEnd]
		authority, err = parseAuthority(hierPart, o)
		if err != nil {
			err = errorsJoin(ErrInvalidURI, err)
			return URI{err: err}, err
		}

		if hierPartEnd+1 < len(raw) {
			if queryEnd < 0 {
				// query ?, no fragment
				query = raw[hierPartEnd+1:]
			} else if hierPartEnd < queryEnd-1 {
				// query ?, fragment
				query = raw[hierPartEnd+1 : queryEnd]
			}
		}

		curr = hierPartEnd + 1
	}

	if queryEnd == len(raw)-1 && hierPartEnd < 0 {
		// trailing #,  no query "?"
		hierPart = raw[curr:queryEnd]
		authority, err = parseAuthority(hierPart, o)
		if err != nil {
			err = errorsJoin(ErrInvalidURI, err)
			return URI{err: err}, err
		}

		u := URI{
			scheme:    scheme,
			hierPart:  hierPart,
			authority: authority,
			query:     query,
		}

		u.authority.ipType, u.err = u.validate(o)
		u.authority.err = u.err // TODO: should propagate only if this is an authority error

		return u, u.err
	}

	if queryEnd > 0 {
		// there is a fragment
		if hierPartEnd < 0 {
			// no query
			hierPart = raw[curr:queryEnd]
			authority, err = parseAuthority(hierPart, o)
			if err != nil {
				err = errorsJoin(ErrInvalidURI, err)
				return URI{err: err}, err
			}
		}

		if queryEnd+1 < len(raw) {
			fragment = raw[queryEnd+1:]
		}
	}

	u := URI{
		scheme:    scheme,
		hierPart:  hierPart,
		query:     query,
		fragment:  fragment,
		authority: authority,
	}

	u.authority.ipType, u.err = u.validate(o)
	u.authority.err = u.err

	return u, u.err
}

// Scheme for this URI.
func (u URI) Scheme() string {
	return u.scheme
}

// Authority information for the URI, including the "//" prefix.
func (u URI) Authority() Authority {
	return u.authority.withEnsuredAuthority()
}

// Query returns a map of key/value pairs of all parameters
// in the query string of the URI.
//
//	This map contains the parsed query parameters like standard lib URL.Query().
func (u URI) Query() url.Values {
	v, _ := url.ParseQuery(u.query)
	return v
}

// Fragment returns the fragment (component preceded by '#') in the
// URI if there is one.
func (u URI) Fragment() string {
	return u.fragment
}

// String representation of an URI.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-6.2.2.1 and later
func (u URI) String() string {
	buf := strings.Builder{}
	buf.Grow(len(u.scheme) + 1 + len(u.query) + 1 + len(u.fragment) + 1 + u.authority.builderSize())

	if len(u.scheme) > 0 {
		buf.WriteString(u.scheme)
		buf.WriteByte(colonMark)
	}

	u.authority.buildString(&buf)

	if len(u.query) > 0 {
		buf.WriteByte(questionMark)
		buf.WriteString(u.query)
	}

	if len(u.fragment) > 0 {
		buf.WriteByte(fragmentMark)
		buf.WriteString(u.fragment)
	}

	return buf.String()
}

// Err is the inner error state of the URI parsing.
func (u URI) Err() error {
	return u.err
}

// validate checks that all parts of a URI abide by allowed characters.
func (u URI) validate(o *options) (ipType, error) {
	if u.scheme != "" && o.validationFlags&flagValidateScheme > 0 {
		if err := u.validateScheme(u.scheme, o); err != nil {
			return ipType{}, err
		}
	}

	if u.query != "" && o.validationFlags&flagValidateQuery > 0 {
		if err := u.validateQuery(u.query, o); err != nil {
			return ipType{}, err
		}
	}

	if u.fragment != "" && o.validationFlags&flagValidateFragment > 0 {
		if err := u.validateFragment(u.fragment, o); err != nil {
			return ipType{}, err
		}
	}

	if u.hierPart != "" {
		return u.authority.validateForScheme(u.scheme, o)
	}

	// empty hierpart case
	return ipType{}, nil
}

// validateScheme verifies the correctness of the scheme part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.1
// scheme = ALPHA *( ALPHA / DIGIT / "+" / "-" / "." )
//
// NOTE: the scheme is not supposed to contain any percent-encoded sequence.
func (u URI) validateScheme(scheme string, _ *options) error {
	if len(scheme) < 2 {
		return ErrInvalidScheme
	}

	c := scheme[0]
	if !isASCIILetter(c) {
		return errorsJoin(
			ErrInvalidScheme,
			fmt.Errorf("an URI scheme must start with an ASCII letter"),
		)
	}

	for i := 1; i < len(scheme); i++ {
		c := scheme[i]
		switch {
		case isDigit(c):
			// ok
		case isASCIILetter(c):
		// ok
		case c == '+' || c == '-' || c == '.':
		// ok
		default:
			return errorsJoin(
				ErrInvalidScheme,
				fmt.Errorf("invalid character %q found in scheme", c),
			)
		}
	}

	return nil
}

// validateQuery validates the query part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.4
//
//	pchar = unreserved / pct-encoded / sub-delims / ":" / "@"
//	fragment    = *( pchar / "/" / "?" )
func (u URI) validateQuery(query string, _ *options) error {
	if err := validateUnreservedWithExtra(query, queryOrFragmentExtraRunes); err != nil {
		return errorsJoin(ErrInvalidQuery, err)
	}

	return nil
}

// validateFragment validatesthe fragment part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.5
//
//	pchar = unreserved / pct-encoded / sub-delims / ":" / "@"
//	fragment    = *( pchar / "/" / "?" )
func (u URI) validateFragment(fragment string, _ *options) error {
	if err := validateUnreservedWithExtra(fragment, queryOrFragmentExtraRunes); err != nil {
		return errorsJoin(ErrInvalidFragment, err)
	}

	return nil
}

func (a Authority) UserInfo() string { return a.userinfo }
func (a Authority) Host() string     { return a.host }
func (a Authority) Port() string     { return a.port }
func (a Authority) Path() string     { return a.path }
func (a Authority) String() string {
	buf := strings.Builder{}
	buf.Grow(a.builderSize())
	a.buildString(&buf)

	return buf.String()
}

func (a Authority) builderSize() int {
	return len(a.prefix) + len(a.userinfo) + 1 + len(a.host) + 2 + len(a.port) + 1 + len(a.path)
}

func (a Authority) buildString(buf *strings.Builder) {
	buf.WriteString(a.prefix)
	buf.WriteString(a.userinfo)

	if len(a.userinfo) > 0 {
		buf.WriteByte(atHost)
	}

	if a.isIPv6 {
		buf.WriteString("[" + a.host + "]")
	} else {
		buf.WriteString(a.host)
	}

	if len(a.port) > 0 {
		buf.WriteByte(colonMark)
	}

	buf.WriteString(a.port)
	buf.WriteString(a.path)
}

// validate the Authority part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.2
func (a Authority) validateForScheme(scheme string, o *options) (ipType, error) {
	var ip ipType

	if a.path != "" && o.validationFlags&flagValidatePath > 0 {
		if err := a.validatePath(a.path, o); err != nil {
			return ip, err
		}
	}

	if a.host != "" && o.validationFlags&flagValidateHost > 0 {
		var err error
		ip, err = a.validateHost(a.host, a.isIPv6, scheme, o)
		if err != nil {
			return ip, err
		}
	}

	if a.port != "" && o.validationFlags&flagValidatePort > 0 {
		if err := a.validatePort(a.port, a.host, o); err != nil {
			return ip, err
		}
	}

	if a.userinfo != "" && o.validationFlags&flagValidateUserInfo > 0 {
		if err := a.validateUserInfo(a.userinfo, o); err != nil {
			return ip, err
		}
	}

	return ip, nil
}

// validatePath validates the path part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.3
func (a Authority) validatePath(path string, _ *options) error {
	if a.host == "" && a.port == "" && len(path) >= 2 && path[0] == slashMark && path[1] == slashMark {
		return errorsJoin(
			ErrInvalidPath,
			fmt.Errorf(
				`if a URI does not contain an authority component, then the path cannot begin with two slash characters ("//"): %q`,
				a.path,
			))
	}

	var previousPos int
	for pos, char := range path {
		if char != slashMark {
			continue
		}

		if pos > previousPos {
			if err := validateUnreservedWithExtra(path[previousPos:pos], pcharExtraRunes); err != nil {
				return errorsJoin(
					ErrInvalidPath,
					err,
				)
			}
		}

		previousPos = pos + 1
	}

	if previousPos < len(path) { // don't care if the last char was a separator
		if err := validateUnreservedWithExtra(path[previousPos:], pcharExtraRunes); err != nil {
			return errorsJoin(
				ErrInvalidPath,
				err,
			)
		}
	}

	return nil
}

// validateHost validates the host part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2
func (a Authority) validateHost(host string, isIPv6 bool, scheme string, o *options) (ipType, error) {
	// check for IP addresses
	// * IPv6 are required to be enclosed within '[]' (isIPv6=true), if an IPv6 zone is present,
	// there is a trailing escaped sequence, but the heading IPv6 literal must not be escaped.
	// * IPv4 are not percent-escaped: strict addresses never contain parts starting a zero (e.g. 012 should be 12).
	// * address the provision made in the RFC for a "IPvFuture"
	if isIPv6 {
		if host[0] == 'v' || host[0] == 'V' {
			if err := validateIPvFuture(host[1:]); err != nil {
				return ipType{}, errorsJoin(
					ErrInvalidHostAddress,
					err,
				)
			}

			return ipType{isIPv6: true, isIPvFuture: true}, nil
		}

		return ipType{isIPv6: true}, validateIPv6(host)
	}

	if err := validateIPv4(host); err == nil {
		return ipType{isIPv4: true}, nil
	}

	// This is not an IP: check for host DNS or registered name
	if err := validateHostForScheme(host, scheme, o); err != nil {
		return ipType{}, errorsJoin(
			ErrInvalidHost,
			err,
		)
	}

	return ipType{}, nil
}

// validateHostForScheme validates the host according to 2 different sets of rules:
//   - if the scheme is a scheme well-known for using DNS host names, the DNS host validation applies (RFC)
//     (applies to schemes at: https://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml)
//   - otherwise, applies the "registered-name" validation stated by RFC 3986:
//
// dns-name see: https://www.rfc-editor.org/rfc/rfc1034, https://www.rfc-editor.org/info/rfc5890
// reg-name    = *( unreserved / pct-encoded / sub-delims )
func validateHostForScheme(host string, scheme string, o *options) error {
	if UsesDNSHostValidation(scheme) {
		if err := validateDNSHostForScheme(host); err != nil {
			return err
		}
	}

	return validateRegisteredHostForScheme(host, o)
}

func validateRegisteredHostForScheme(host string, _ *options) error {
	// RFC 3986 registered name
	if err := validateUnreservedWithExtra(host, nil); err != nil {
		return errorsJoin(
			ErrInvalidRegisteredName,
			err,
		)
	}

	return nil
}

// validatePort validates the port part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.2.3
//
// port = *DIGIT
func (a Authority) validatePort(port, host string, _ *options) error {
	const maxPort uint64 = 65535

	if !isNumerical(port) {
		return ErrInvalidPort
	}

	if host == "" {
		return errorsJoin(
			ErrMissingHost,
			fmt.Errorf("whenever a port is specified, a host part must be present"),
		)
	}

	portNum, _ := strconv.ParseUint(port, 10, 64)
	if portNum > maxPort {
		return errorsJoin(
			ErrInvalidPort,
			fmt.Errorf("a valid port lies in the range (0-%d)", maxPort),
		)
	}

	return nil
}

// validateUserInfo validates the userinfo part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.2.1
//
// userinfo    = *( unreserved / pct-encoded / sub-delims / ":" )
func (a Authority) validateUserInfo(userinfo string, _ *options) error {
	if err := validateUnreservedWithExtra(userinfo, userInfoExtraRunes); err != nil {
		return errorsJoin(
			ErrInvalidUserInfo,
			err,
		)
	}

	return nil
}

func parseAuthority(hier string, _ *options) (Authority, error) {
	// as per RFC 3986 Section 3.6
	var (
		prefix, userinfo, host, port, path string
		isIPv6                             bool
	)

	// authority sections MUST begin with a '//'
	if strings.HasPrefix(hier, authorityPrefix) {
		prefix = authorityPrefix
		hier = strings.TrimPrefix(hier, authorityPrefix)
	}

	if prefix == "" {
		path = hier
	} else {
		// authority   = [ userinfo "@" ] host [ ":" port ]
		slashEnd := strings.IndexByte(hier, slashMark)
		if slashEnd > -1 {
			if slashEnd < len(hier) {
				path = hier[slashEnd:]
			}
			hier = hier[:slashEnd]
		}

		host = hier
		if at := strings.IndexByte(host, atHost); at > 0 {
			userinfo = host[:at]
			if at+1 < len(host) {
				host = host[at+1:]
			}
		}

		if bracket := strings.IndexByte(host, openingBracketMark); bracket >= 0 {
			// ipv6 addresses: "["xx:yy:zz"]":port
			rawHost := host
			closingbracket := strings.IndexByte(host, closingBracketMark)
			switch {
			case closingbracket > bracket+1:
				host = host[bracket+1 : closingbracket]
				rawHost = rawHost[closingbracket+1:]
				isIPv6 = true
			case closingbracket > bracket:
				return Authority{}, errorsJoin(
					ErrInvalidHostAddress,
					fmt.Errorf("empty IPv6 address"),
				)
			default:
				return Authority{}, errorsJoin(
					ErrInvalidHostAddress,
					fmt.Errorf("mismatched square brackets"),
				)
			}

			if colon := strings.IndexByte(rawHost, colonMark); colon >= 0 {
				if colon+1 < len(rawHost) {
					port = rawHost[colon+1:]
				}
			}
		} else {
			if colon := strings.IndexByte(host, colonMark); colon >= 0 {
				if colon+1 < len(host) {
					port = host[colon+1:]
				}
				host = host[:colon]
			}
		}
	}

	return Authority{
		prefix:   prefix,
		userinfo: userinfo,
		host:     host,
		port:     port,
		path:     path,
		ipType:   ipType{isIPv6: isIPv6}, // provisional flag
	}, nil
}

func (a Authority) Err() error {
	return a.err
}

func miniMaxi(vals ...int) (int, int) {
	var mini, maxi int
	if len(vals) == 0 {
		return mini, maxi
	}

	mini, maxi = vals[0], vals[0]

	for _, val := range vals[1:] {
		if val < mini {
			mini = val
		}
		if val > maxi {
			maxi = val
		}
	}

	if mini < 0 {
		mini = 0
	}
	if maxi < 0 {
		maxi = 0
	}

	return mini, maxi
}
