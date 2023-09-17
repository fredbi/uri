package uri

import "errors"

// Error from the github.com/fredbi/uri module.
type Error interface {
	error
}

// Validation errors.
var (
	ErrNoSchemeFound         = Error(errors.New("no scheme found in URI"))
	ErrInvalidURI            = Error(errors.New("not a valid URI"))
	ErrInvalidCharacter      = Error(errors.New("invalid character in URI"))
	ErrInvalidScheme         = Error(errors.New("invalid scheme in URI"))
	ErrInvalidQuery          = Error(errors.New("invalid query string in URI"))
	ErrInvalidFragment       = Error(errors.New("invalid fragment in URI"))
	ErrInvalidPath           = Error(errors.New("invalid path in URI"))
	ErrInvalidHost           = Error(errors.New("invalid host in URI"))
	ErrInvalidPort           = Error(errors.New("invalid port in URI"))
	ErrInvalidUserInfo       = Error(errors.New("invalid userinfo in URI"))
	ErrMissingHost           = Error(errors.New("missing host in URI"))
	ErrInvalidHostAddress    = Error(errors.New("invalid address for host"))
	ErrInvalidRegisteredName = Error(errors.New("invalid host (registered name)"))
	ErrInvalidDNSName        = Error(errors.New("invalid host (DNS name)"))
)
