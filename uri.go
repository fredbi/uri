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
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// URI represents a general RFC3986 URI.
type URI interface {
	// Scheme the URI conforms to.
	Scheme() string

	// Authority information for the URI, including the "//" prefix.
	Authority() Authority

	// Query returns a map of key/value pairs of all parameters
	// in the query string of the URI.
	Query() url.Values

	// Fragment returns the fragment (component preceded by '#') in the
	// URI if there is one.
	Fragment() string

	// Builder returns a Builder that can be used to modify the URI.
	Builder() Builder

	// String representation of the URI
	String() string

	// Validate the different components of the URI
	Validate() error
}

// Authority information that a URI contains
// as specified by RFC3986.
//
// Username and password are given by UserInfo().
type Authority interface {
	UserInfo() string
	Host() string
	Port() string
	Path() string
	String() string
	Validate(...string) error
}

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
	authorityPrefix    = "//"
)

// IsURI tells if a URI is valid according to RFC3986/RFC397.
func IsURI(raw string) bool {
	_, err := Parse(raw)
	return err == nil
}

// IsURIReference tells if a URI reference is valid according to RFC3986/RFC397
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-4.1 and
// https://www.rfc-editor.org/rfc/rfc3986#section-4.2
func IsURIReference(raw string) bool {
	_, err := ParseReference(raw)
	return err == nil
}

// Parse attempts to parse a URI.
// It returns an error if the URI is not RFC3986-compliant.
func Parse(raw string) (URI, error) {
	return parse(raw, false)
}

// ParseReference attempts to parse a URI relative reference.
//
// It returns an error if the URI is not RFC3986-compliant.
func ParseReference(raw string) (URI, error) {
	return parse(raw, true)
}

func parse(raw string, withURIReference bool) (URI, error) {
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
		return nil, ErrInvalidURI
	}

	if schemeEnd == 1 {
		return nil, errors.Join(
			ErrInvalidScheme,
			fmt.Errorf("scheme has a minimum length of 2 characters"),
		)
	}

	if hierPartEnd == 1 || queryEnd == 1 {
		// ".:", ".?", ".#"
		return nil, ErrInvalidURI
	}

	if hierPartEnd > 0 && hierPartEnd < schemeEnd || queryEnd > 0 && queryEnd < schemeEnd {
		// e.g. htt?p: ; h#ttp: ..
		return nil, ErrInvalidURI
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
			u := &uri{
				scheme: scheme,
			}

			return u, u.Validate()
		}
	case !withURIReference:
		// scheme is required for URI
		return nil, errors.Join(
			ErrNoSchemeFound,
			fmt.Errorf("for URI (not URI reference), the scheme is required"),
		)
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

		authorityInfo, err := parseAuthority(raw[curr:hierPartEnd])
		if err != nil {
			return nil, errors.Join(ErrInvalidURI, err)
		}

		u := &uri{
			scheme:    scheme,
			hierPart:  raw[curr:hierPartEnd],
			authority: authorityInfo,
		}

		return u, u.Validate()
	}

	var (
		hierPart, query, fragment string
		authorityInfo             *authorityInfo
		err                       error
	)

	if hierPartEnd > 0 {
		hierPart = raw[curr:hierPartEnd]
		authorityInfo, err = parseAuthority(hierPart)
		if err != nil {
			return nil, errors.Join(ErrInvalidURI, err)
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
		authorityInfo, err = parseAuthority(hierPart)
		if err != nil {
			return nil, errors.Join(ErrInvalidURI, err)
		}

		u := &uri{
			scheme:    scheme,
			hierPart:  hierPart,
			authority: authorityInfo,
			query:     query,
		}

		return u, u.Validate()
	}

	if queryEnd > 0 {
		// there is a fragment
		if hierPartEnd < 0 {
			// no query
			hierPart = raw[curr:queryEnd]
			authorityInfo, err = parseAuthority(hierPart)
			if err != nil {
				return nil, errors.Join(ErrInvalidURI, err)
			}
		}

		if queryEnd+1 < len(raw) {
			fragment = raw[queryEnd+1:]
		}
	}

	u := &uri{
		scheme:    scheme,
		hierPart:  hierPart,
		query:     query,
		fragment:  fragment,
		authority: authorityInfo,
	}

	return u, u.Validate()
}

type uri struct {
	// raw components
	scheme   string
	hierPart string
	query    string
	fragment string

	// parsed components
	authority *authorityInfo
}

func (u *uri) URI() URI {
	return u
}

func (u *uri) Scheme() string {
	return u.scheme
}

func (u *uri) Authority() Authority {
	u.ensureAuthorityExists()
	return u.authority
}

// Query returns parsed query parameters like standard lib URL.Query().
func (u *uri) Query() url.Values {
	v, _ := url.ParseQuery(u.query)
	return v
}

func (u *uri) Fragment() string {
	return u.fragment
}

var (
	rexScheme   = regexp.MustCompile(`^[\p{L}][\p{L}\d\+-\.]+$`)
	rexFragment = regexp.MustCompile(`^([\p{L}\d\-\._~\:@!\$\&'\(\)\*\+,;=\?/]|(%[[:xdigit:]]{2})+)+$`)
	rexQuery    = rexFragment
	rexSegment  = regexp.MustCompile(`^([\p{L}\d\-\._~\:@!\$\&'\(\)\*\+,;=]|(%[[:xdigit:]]{2})+)+$`)
	rexHostname = regexp.MustCompile(`^[a-zA-Z0-9\p{L}]((-?[a-zA-Z0-9\p{L}]+)?|(([a-zA-Z0-9-\p{L}]{0,63})(\.)){1,6}([a-zA-Z\p{L}]){2,})$`)

	// unreserved | pct-encoded | sub-delims.
	rexRegname = regexp.MustCompile(`^([\p{L}\d\-\._~!\$\&'\(\)\*\+,;=]|(%[[:xdigit:]]{2})+)+$`)
	// unreserved | pct-encoded | sub-delims | ":".
	rexUserInfo = regexp.MustCompile(`^([\p{L}\d\-\._~\:!\$\&'\(\)\*\+,;=\?/]|(%[[:xdigit:]]{2})+)+$`)

	rexIPv6Zone = regexp.MustCompile(`:[^%:]+%25(([\p{L}\d\-\._~\:@!\$\&'\(\)\*\+,;=]|(%[[:xdigit:]]{2}))+)?$`)
)

func isNumerical(input string) bool {
	return strings.IndexFunc(input,
		func(r rune) bool { return r < '0' || r > '9' },
	) == -1
}

// Validate checks that all parts of a URI abide by allowed characters.
func (u *uri) Validate() error {
	if u.scheme != "" {
		if err := u.validateScheme(u.scheme); err != nil {
			return err
		}
	}

	if u.query != "" {
		if err := u.validateQuery(u.query); err != nil {
			return err
		}
	}

	if u.fragment != "" {
		if err := u.validateFragment(u.fragment); err != nil {
			return err
		}
	}

	if u.hierPart != "" {
		if u.authority != nil {
			return u.Authority().Validate(u.scheme)
		}
	}

	// empty hierpart case
	return nil
}

// validateScheme verifies the correctness of the scheme part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.1
//
//	scheme = ALPHA *( ALPHA / DIGIT / "+" / "-" / "." )
//
// NOTE: scheme is not supposed to contain any percent-encoded sequence.
// TODO(fredbi): verify the IRI RFC to check if unicode is allowed in scheme.
func (u *uri) validateScheme(scheme string) error {
	if ok := rexScheme.MatchString(scheme); !ok {
		return ErrInvalidScheme
	}

	return nil
}

// validateQuery validates the query part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.4
//
//	pchar = unreserved / pct-encoded / sub-delims / ":" / "@"
//	query = *( pchar / "/" / "?" )
func (u *uri) validateQuery(query string) error {
	if ok := rexQuery.MatchString(query); !ok {
		return ErrInvalidQuery
	}

	return nil
}

// validateFragment validatesthe fragment part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.5
//
//	pchar = unreserved / pct-encoded / sub-delims / ":" / "@"
//
// fragment    = *( pchar / "/" / "?" )
func (u *uri) validateFragment(fragment string) error {
	if ok := rexFragment.MatchString(fragment); !ok {
		return ErrInvalidFragment
	}

	return nil
}

type authorityInfo struct {
	prefix   string
	userinfo string
	host     string
	port     string
	path     string
}

func (a authorityInfo) UserInfo() string { return a.userinfo }
func (a authorityInfo) Host() string     { return a.host }
func (a authorityInfo) Port() string     { return a.port }
func (a authorityInfo) Path() string     { return a.path }
func (a authorityInfo) String() string {
	buf := strings.Builder{}
	buf.WriteString(a.prefix)
	buf.WriteString(a.userinfo)

	if len(a.userinfo) > 0 {
		buf.WriteByte(atHost)
	}

	if strings.IndexByte(a.host, colonMark) > 0 {
		// ipv6 address host
		buf.WriteString("[" + a.host + "]")
	} else {
		buf.WriteString(a.host)
	}

	if len(a.port) > 0 {
		buf.WriteByte(colonMark)
	}

	buf.WriteString(a.port)
	buf.WriteString(a.path)
	return buf.String()
}

// Validate the Authority part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.2
func (a authorityInfo) Validate(schemes ...string) error {
	if a.path != "" {
		if err := a.validatePath(a.path); err != nil {
			return err
		}
	}

	if a.host != "" {
		if err := a.validateHost(a.host, schemes...); err != nil {
			return err
		}
	}

	if a.port != "" {
		if err := a.validatePort(a.port, a.host); err != nil {
			return err
		}
	}

	if a.userinfo != "" {
		if err := a.validateUserInfo(a.userinfo); err != nil {
			return err
		}
	}

	return nil
}

// validatePath validates the path part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.3
func (a authorityInfo) validatePath(path string) error {
	for _, segment := range strings.Split(path, "/") {
		if segment == "" {
			continue
		}
		if ok := rexSegment.MatchString(segment); !ok {
			return ErrInvalidPath
		}
	}

	return nil
}

// validateHost validates the host part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2
func (a authorityInfo) validateHost(host string, schemes ...string) error {
	var isIP bool
	if ok := rexIPv6Zone.MatchString(host); ok {
		z := strings.IndexByte(a.host, percentMark)
		isIP = net.ParseIP(host[0:z]) != nil
	} else {
		isIP = net.ParseIP(host) != nil
	}

	if !isIP {
		// this is not an IP, check for host DNS or registered name
		if err := validateHostForScheme(a.host, schemes...); err != nil {
			return err
		}
	}

	return nil
}

// validateHostForScheme validates the host according to 2 different sets of rules:
//   - if the scheme is a scheme well-known for using DNS host names, the DNS host validation applies (RFC)
//     (applies to schemes at: https://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml)
//   - otherwise, applies the "registered-name" validation stated by RFC 3986:
//
// dns-name see: https://www.rfc-editor.org/rfc/rfc1034, https://www.rfc-editor.org/info/rfc5890
// reg-name    = *( unreserved / pct-encoded / sub-delims )
func validateHostForScheme(host string, schemes ...string) error {
	var isHost bool
	unescapedHost, err := url.PathUnescape(host)
	if err != nil {
		return ErrInvalidHost
	}

	for _, scheme := range schemes {
		if UsesDNSHostValidation(scheme) {
			// DNS name
			isHost = rexHostname.MatchString(unescapedHost)
		} else {
			// standard RFC 3986
			isHost = rexRegname.MatchString(unescapedHost)
		}

		if !isHost {
			return ErrInvalidHost
		}
	}

	return nil
}

// validatePort validates the port part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.2.3
//
// port = *DIGIT
func (a authorityInfo) validatePort(port, host string) error {
	if !isNumerical(port) {
		return ErrInvalidPort
	}

	if host == "" {
		return ErrMissingHost
	}

	return nil
}

// validateUserInfo validates the userinfo part.
//
// Reference: https://www.rfc-editor.org/rfc/rfc3986#section-3.2.1
//
// userinfo    = *( unreserved / pct-encoded / sub-delims / ":" )
func (a authorityInfo) validateUserInfo(userinfo string) error {
	if ok := rexUserInfo.MatchString(userinfo); !ok {
		return ErrInvalidUserInfo
	}

	return nil
}

func parseAuthority(hier string) (*authorityInfo, error) {
	// as per RFC 3986 Section 3.6
	var prefix, userinfo, host, port, path string

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
			// ipv6 addresses: "[" xx:yy:zz "]":port
			rawHost := host
			closingbracket := strings.IndexByte(host, closingBracketMark)
			if closingbracket > bracket+1 {
				host = host[bracket+1 : closingbracket]
				rawHost = rawHost[closingbracket+1:]
			} else {
				return nil, ErrInvalidURI
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

	return &authorityInfo{
		prefix:   prefix,
		userinfo: userinfo,
		host:     host,
		port:     port,
		path:     path,
	}, nil
}

func (u *uri) ensureAuthorityExists() {
	if u.authority == nil {
		u.authority = &authorityInfo{}

		return
	}

	if u.authority.userinfo != "" ||
		u.authority.host != "" ||
		u.authority.port != "" {
		u.authority.prefix = "//"
	}
}

func (u *uri) String() string {
	buf := strings.Builder{}
	if len(u.scheme) > 0 {
		buf.WriteString(u.scheme)
		buf.WriteByte(colonMark)
	}

	buf.WriteString(u.authority.String())

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
