package uri

import (
	"sync"

	"golang.org/x/net/idna"
)

type (
	// Option allows for fine-grained tuning of tolerances to standards
	// when validating an URI
	Option func(*options)

	//nolint: unused
	options struct {
		schemeIsDNSFunc       func(string) bool
		defaultPortFunc       func(string) int
		idnaFlags             []idna.Option
		withDNSHostValidation bool
		withStrictASCII       bool
		withStrictIPv6        bool
		withURIReference      bool
		withStrictURI         bool
		withStrictIRI         bool
		withWindowsFriendly   bool
		withRedactedPassword  bool

		// select validations: this is used by builder methods to carry out
		// partial validation.
		validationFlags uint16
	}

	// optionsPool holds allocated options in a pool,
	// to avoid undue gc pressure when using custom options
	// intensively.
	// Notice that default options (possibly customized once)
	// do not allocate anything.
	optionsPool struct {
		*sync.Pool
	}
)

const (
	flagValidateScheme uint16 = 1 << iota
	flagValidateHost
	flagValidatePort
	flagValidateUserInfo
	flagValidatePath
	flagValidateQuery
	flagValidateFragment
)

var (
	packageLevelDefaults = options{
		schemeIsDNSFunc: UsesDNSHostValidation,
		validationFlags: ^uint16(0),
	}

	packageLevelReferenceDefaults = options{
		schemeIsDNSFunc:  UsesDNSHostValidation,
		withURIReference: true,
		validationFlags:  ^uint16(0),
	}

	muxDefaults   sync.Mutex
	poolOfOptions = optionsPool{
		Pool: &sync.Pool{
			New: func() any {
				return defaultOptions()
			},
		},
	}
)

// borrowOptions reuses a previously allocated option from the pool.
//
// Optional behavior is reset to the package-level defaults.
func borrowURIOptions() *options {
	o := poolOfOptions.Get().(*options)
	*o = packageLevelDefaults

	return o
}

func borrowURIReferenceOptions() *options {
	o := poolOfOptions.Get().(*options)
	*o = packageLevelReferenceDefaults

	return o
}

func redeemOptions(o *options) {
	if o == &packageLevelDefaults || o == &packageLevelReferenceDefaults {
		return
	}
	poolOfOptions.Put(o)
}

// defaultOptions allocates a new struct to hold options
func defaultOptions() *options {
	o := packageLevelDefaults // shallow-clone defaults

	return &o
}

// applyURIOptions applies options on a struct borrowed from the pool,
// with defaults reset to support URIs (not URI references).
//
// **Don't mutate the returned options**
func applyURIOptions(opts []Option) (*options, func(*options)) {
	if len(opts) == 0 {
		// no overrides, no need to allocate a copy of the options
		return &packageLevelDefaults, redeemOptions
	}

	o := borrowURIOptions()

	for _, apply := range opts {
		apply(o)
	}

	return o, redeemOptions
}

// applyURIOptions applies options on a struct borrowed from the pool,
// with defaults reset to support URI references.
//
// **Don't mutate the returned options**
func applyURIReferenceOptions(opts []Option) (*options, func(*options)) {
	if len(opts) == 0 {
		// no overrides, no need to allocate a copy of the options
		return &packageLevelReferenceDefaults, redeemOptions
	}

	o := borrowURIReferenceOptions()

	for _, apply := range opts {
		apply(o)
	}

	return o, redeemOptions
}

// SetDefaultOptions allows to tweak package level defaults.
//
// You should only use this in initialization steps, as this manipulates
// a package global variable.
func SetDefaultOptions(opts ...Option) {
	muxDefaults.Lock()
	defer muxDefaults.Unlock()

	o := &packageLevelDefaults
	p := &packageLevelReferenceDefaults

	for _, apply := range opts {
		apply(o)
		apply(p)
	}
}

func withValidationFlags(flags uint16) Option {
	return func(o *options) {
		o.validationFlags = flags
	}
}

// WithSchemeIsDNSFunc overrides the default DNS scheme identification function.
//
// The passed function is assumed to return true whenever a (lower cased) scheme
// should be considered to use Internet domain names.
func WithSchemeIsDNSFunc(fn func(string) bool) Option {
	return func(o *options) {
		o.schemeIsDNSFunc = fn
	}
}

// WithDefaultPortFunc overrides the default scheme to default port function.
//
// The passed function is assumed to return a port number for a (lower cased) scheme.
func WithDefaultPortFunc(fn func(string) int) Option {
	return func(o *options) {
		o.defaultPortFunc = fn
	}
}

// WithDNSSchemes adds extra schemes to the DNS host name validation.
func WithDNSSchemes(_ ...string) Option {
	return func(o *options) {
		// TODO
	}
}

// WithReference tells the validator whether to accept URI references.
func WithReference(enabled bool) Option {
	return func(o *options) {
		o.withURIReference = enabled
	}
}

// WithIDNAFlags sets golang.org/x/idna.Option's for domain name validation
func WithIDNAFlags(flags ...idna.Option) Option {
	return func(o *options) {
		o.idnaFlags = flags
	}
}

// WithStrictURI tells the validator to be strict regarding RFC3986 for URIs.
//
// This means that only ASCII characters are accepted (other unicode character MUST be escaped).
func WithStrictURI(enabled bool) Option {
	return func(o *options) {
		o.withStrictURI = enabled
	}
}

// WithStrictIRI tells the validator to be strict regarding RFC3987 for IRIs
//
// This means that all valid UCS characters are accepted without escaping.
func WithStrictIRI(enabled bool) Option {
	return func(o *options) {
		o.withStrictIRI = enabled
	}
}

// WithWindowsFriendly tells the validator to accept Windows file paths that
// are common, but formally invalid URI path (e.g. 'C:\folder\File.txt').
//
// This deviation is only supported for scheme "file", so the following URI is tolerated
// and parsed as a legit URI:
// file://C:\folder\file.txt
//
// is internally transformed to the expected:
// file:///C:/folder/file.txt
//
// Notice that URI references do not support this option if the scheme is not specified.
func WithWindowsFriendly(enabled bool) Option {
	return func(o *options) {
		o.withWindowsFriendly = enabled
	}
}
