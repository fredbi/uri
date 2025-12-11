package uri

import (
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

func TestValidateIPv4(t *testing.T) {
	require.NoError(t, validateIPv4("127.0.0.1"))
	require.NoError(t, validateIPv4("255.255.255.255"))
	require.NoError(t, validateIPv4("0.0.0.0"))

	require.Error(t, validateIPv4("01.0.0.0"))
	require.Error(t, validateIPv4("1.0.0.0.1"))
	require.Error(t, validateIPv4("256.0.0.0"))
	require.Error(t, validateIPv4("261.0.0.0"))
	require.Error(t, validateIPv4("255.255.255"))
	require.Error(t, validateIPv4("255.255.255.255.256"))
	require.Error(t, validateIPv4("1111.255.255.255"))
	require.Error(t, validateIPv4("::"))
	require.Error(t, validateIPv4("1.2.3.%31"))
	require.Error(t, validateIPv4("-1.2.3.4"))
	require.Error(t, validateIPv4("1..3.4"))
}
