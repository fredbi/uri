package uri

// Builder methods

func (u URI) WithScheme(scheme string, opts ...Option) URI {
	if u.Err() != nil {
		return u
	}

	opts = append(opts, withValidationFlags(flagValidateScheme|flagValidateHost))
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	u.scheme = scheme
	u.authority.ipType, u.err = u.validate(o)

	return u
}

func (u URI) WithAuthority(authority Authority, opts ...Option) URI {
	if u.Err() != nil {
		return u
	}

	opts = append(opts, withValidationFlags(flagValidateHost|flagValidatePort|flagValidateUserInfo|flagValidatePath))
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	u.authority = authority
	u.authority.ipType, u.err = u.validate(o)
	u.authority.err = u.err

	return u
}

func (u URI) WithUserInfo(userinfo string, opts ...Option) URI {
	if u.Err() != nil {
		return u
	}

	opts = append(opts, withValidationFlags(flagValidateUserInfo))
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	u.authority = u.authority.withEnsuredAuthority()
	u.authority.userinfo = userinfo
	u.authority.ipType, u.err = u.validate(o)
	u.authority.err = u.err

	return u
}

func (u URI) WithHost(host string, opts ...Option) URI {
	if u.Err() != nil {
		return u
	}

	opts = append(opts, withValidationFlags(flagValidateHost|flagValidatePort))
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	u.authority = u.authority.withEnsuredAuthority()
	u.authority.host = host
	u.authority.ipType, u.err = u.validate(o)
	u.authority.err = u.err

	return u
}

func (u URI) WithPort(port string, opts ...Option) URI { // TODO: port as int?
	if u.Err() != nil {
		return u
	}

	opts = append(opts, withValidationFlags(flagValidatePort))
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	u.authority = u.authority.withEnsuredAuthority()
	u.authority.port = port
	u.authority.ipType, u.err = u.validate(o)
	u.authority.err = u.err

	return u
}

func (u URI) WithPath(path string, opts ...Option) URI {
	if u.Err() != nil {
		return u
	}

	opts = append(opts, withValidationFlags(flagValidatePath))
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	u.authority = u.authority.withEnsuredAuthority()
	u.authority.path = path
	u.authority.ipType, u.err = u.validate(o)
	u.authority.err = u.err

	return u
}

func (u URI) WithQuery(query string, opts ...Option) URI {
	if u.Err() != nil {
		return u
	}

	opts = append(opts, withValidationFlags(flagValidateQuery))
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	u.query = query
	u.authority.ipType, u.err = u.validate(o)

	return u
}

func (u URI) WithFragment(fragment string, opts ...Option) URI {
	if u.Err() != nil {
		return u
	}

	opts = append(opts, withValidationFlags(flagValidateFragment))
	o, redeem := applyURIOptions(opts)
	defer func() { redeem(o) }()

	u.fragment = fragment
	u.authority.ipType, u.err = u.validate(o)

	return u
}

func (a Authority) withEnsuredAuthority() Authority {
	if a.userinfo != "" || a.host != "" || a.port != "" {
		a.prefix = authorityPrefix
	}

	return a
}
