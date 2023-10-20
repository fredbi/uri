package uri

type (
	NormalizeOption func(*normalizeOptions)

	normalizeOptions struct {
		escapeUnicode bool
		asciiHost     bool
	}
)

var defaultNormalizeOptions = &normalizeOptions{}

func normalizeOptionsWithDefaults(opts []NormalizeOption) *normalizeOptions {
	if len(opts) == 0 {
		return defaultNormalizeOptions
	}

	o := *defaultNormalizeOptions

	for _, apply := range opts {
		apply(&o)
	}

	return &o
}

func WithEscapeUnicode(enabled bool) NormalizeOption {
	return func(o *normalizeOptions) {
		o.escapeUnicode = enabled
	}
}

func WithASCIIHost(enabled bool) NormalizeOption {
	return func(o *normalizeOptions) {
		o.asciiHost = enabled
	}
}
