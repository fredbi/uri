package uri

import (
	"path"
)

func (u URI) WithJoinPath(elems ...string) URI {
	if u.Err() != nil {
		return u
	}

	o, redeem := applyURIOptions([]Option{withValidationFlags(flagValidatePath)})
	defer func() { redeem(o) }()

	// Ref:
	// x, zee := ur.Parse("ez")
	// x.JoinPath(elem...)
	u.authority = u.authority.withEnsuredAuthority()
	full := append([]string{u.authority.path}, elems...)
	u.authority.path = path.Join(full...)
	u.authority.ipType, u.err = u.validate(o)
	u.authority.err = u.err
	return u
}

func (u URI) WithUserPassword(username, password string) URI { //nolint: unparam,revive
	return URI{}
}

func (u URI) WithRedacted() URI {
	return URI{}
}

func (u URI) RequestURI() string {
	return "" // TODO
}

func (u URI) ResolveReference(ref URI) URI { //nolint: unparam,revive
	return URI{} // TODO
}

// Not builders
func (u URI) EscapedFragment() string {
	// TODO
	return u.fragment
}

func (u URI) IsReference() bool {
	return false // TODO
}

func (a Authority) Redacted() string { // NOTE: net/url.URL mutates
	return "" // TODO
}

func (a Authority) Username() string {
	return ""
}

func (a Authority) User() string {
	return ""
}

func (a Authority) Password() (string, bool) {
	return "", false
}
