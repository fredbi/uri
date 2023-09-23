package uri

import "errors"

// errSentinelTest is used in test cases for when we want to assert an error
// but do not want to check specifically which error was returned.
var errSentinelTest = Error(errors.New("test"))
