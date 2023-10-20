package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalized(t *testing.T) {
	t.Run("with normalized scheme", func(t *testing.T) {
		// * scheme is lower-cased
		u, err := Parse("hTTp:///target")
		require.NoError(t, err)
		// require.True(t, u.IsDefaultPort())

		n, _ := u.Normalized()
		normalized := n.String()
		require.Equal(t, "http:///target", normalized)
		normalizeString, _ := u.Normalize()
		require.Equal(t, normalized, normalizeString)
	})

	t.Run("with default port omitted", func(t *testing.T) {
		// * scheme is lower-cased
		// * default http port is omitted
		u, err := Parse("hTTp://host:80/target")
		require.NoError(t, err)
		// require.True(t, u.IsDefaultPort())

		n, _ := u.Normalized()
		normalized := n.String()
		require.Equal(t, "http://host/target", normalized)
		normalizedString, _ := u.Normalize()
		require.Equal(t, normalized, normalizedString)
	})

	t.Run("with normalized host case", func(t *testing.T) {
		// * scheme is lower-cased
		// * password is escaped
		// * host is lower-cased
		// * default http port is omitted
		// * double /, dots in path are simplified
		// * trailing / in path is omitted
		u, err := Parse("hTTp://fred:passw*oRd@Host:80/path//./ending/with/slash/")
		require.NoError(t, err)
		// require.True(t, u.IsDefaultPort())

		n, _ := u.Normalized()
		normalized := n.String()
		require.Equal(t, "http://fred:passw%2AoRd@host/path/ending/with/slash", normalized)
		normalizedString, _ := u.Normalize()
		require.Equal(t, normalized, normalizedString)
	})

	t.Run("with normalized path simplified", func(t *testing.T) {
		u, err := Parse("http://path//./ending/../with/slash/")
		require.NoError(t, err)
		// require.True(t, u.IsDefaultPort())

		n, _ := u.Normalized()
		normalized := n.String()
		require.Equal(t, "http://path/with/slash", normalized)
		normalizedString, _ := u.Normalize()
		require.Equal(t, normalized, normalizedString)
	})

	t.Run("with normalized percent-encoding", func(t *testing.T) {
		// * non-default port kept
		// * percent-encoding is upper-cased
		// * unnecessary encoding is removed
		u, err := Parse("https://fred:passw%5boRd@Host:80/path//./%41%5b%5d%4aencoded")
		require.NoError(t, err)
		// require.False(t, u.IsDefaultPort())

		n, _ := u.Normalized()
		t.Logf("%#v", n)
		normalized := n.String()
		t.Log(normalized)
		require.Equal(t, "https://fred:passw%5BoRd@host:80/path/A%5B%5DJencoded", normalized)
		normalizedString, err := u.Normalize()
		require.Equal(t, normalized, normalizedString)
	})

	t.Run("with percent-encoding of multi-byte UTF8 sequences", func(t *testing.T) {
		// TODO
		chars := `ば 	ぱ 	ひ 	び 	ぴ 	ふ 	ぶ 	ぷ 	へ 	べ`
		u, err := Parse(fmt.Sprintf("file://path/%s"))
		require.NoError(t, err)
	})

	t.Run("with normalized query", func(t *testing.T) {
		u, err := Parse("https://fred:passw%5boRd@Host:80/path//./%41%5b%5d%4aencoded?a=1&b=%25&c=%5b&d=%C3%A8")
		require.NoError(t, err)
		// require.False(t, u.IsDefaultPort())

		n, err := u.Normalized()
		require.NoError(t, err)
		t.Logf("%#v", n)
		normalized := n.String()
		t.Log(normalized)
		require.Equal(t, "https://fred:passw%5BoRd@host:80/path/A%5B%5DJencoded?a=1&b=%25&c=%5B&d=è", normalized)
		normalizedString, _ := u.Normalize()
		require.Equal(t, normalized, normalizedString)
	})

	t.Run("with normalized fragment", func(t *testing.T) {
	})

	t.Run("with ASCII hostname (punycode)", func(t *testing.T) {
		u, err := Parse("https://fred:passw%5boRd@hàôé.com")
		require.NoError(t, err)
		// require.False(t, u.IsDefaultPort())

		n, err := u.Normalized(WithASCIIHost(true))
		require.NoError(t, err)
		t.Logf("%#v", n)
		normalized := n.String()
		t.Log(normalized)
		require.Equal(t, "https://fred:passw%5BoRd@xn--h-sfa1a6b.com/", normalized)
		normalizedString, _ := u.Normalize(WithASCIIHost(true))
		require.Equal(t, normalized, normalizedString)
	})
}
