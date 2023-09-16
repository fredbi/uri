package uri

import "errors"

// Validation errors.
var (
	ErrNoSchemeFound    = errors.New("no scheme found in URI")
	ErrInvalidURI       = errors.New("not a valid URI")
	ErrInvalidCharacter = errors.New("invalid character in URI")
	ErrInvalidScheme    = errors.New("invalid scheme in URI")
	ErrInvalidQuery     = errors.New("invalid query string in URI")
	ErrInvalidFragment  = errors.New("invalid fragment in URI")
	ErrInvalidPath      = errors.New("invalid path in URI")
	ErrInvalidHost      = errors.New("invalid host in URI")
	ErrInvalidPort      = errors.New("invalid port in URI")
	ErrInvalidUserInfo  = errors.New("invalid userinfo in URI")
	ErrMissingHost      = errors.New("missing host in URI")
)
