package uri

// MashalText yields an URI as UTF8-encoded bytes
func (u URI) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// MarshalBinary is like MarshalText
func (u URI) MarshalBinary() ([]byte, error) {
	return u.MarshalText()
}

// UnmarshalText unmarshals an URI from UTF8-encoded bytes.
//
// If the original input is not UTF8, consider translating it first from
// the original character set, e.g. using github.com/paulrosania/go-charset.
//
// Notice that:
// * URI references are not accepted by default
// * only package-level default options are applicable
//
// Callers may set package-level defaults to alter the default behavior.
func (u *URI) UnmarshalText(b []byte) error {
	o, redeem := applyURIOptions(nil) // default options
	defer func() { redeem(o) }()

	v, err := parse(string(b), o)
	if err != nil {
		return err
	}

	*u = v

	return nil
}

// UnmarshalBinary is like UnmarshalText
func (u *URI) UnmarshalBinary(b []byte) error {
	return u.UnmarshalText(b)
}
