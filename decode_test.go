package uri

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestUnescapePercentEncoding(t *testing.T) {
	t.Run("with valid UTF8 sequence", func(t *testing.T) {
		t.Run("one byte", func(t *testing.T) {
			r, offset, err := unescapePercentEncoding("25")
			require.NoError(t, err)
			require.Equal(t, '%', r)
			require.Equal(t, 2, offset)
		})
		t.Run("two bytes", func(t *testing.T) {
			r, offset, err := unescapePercentEncoding("C3%b6")
			require.NoError(t, err)
			require.Equal(t, 'ö', r)
			require.Equal(t, 5, offset)
		})
		t.Run("three bytes", func(t *testing.T) {
			r, offset, err := unescapePercentEncoding("E3%a3%Af")
			require.NoError(t, err)
			require.Equal(t, '㣯', r)
			require.Equal(t, 8, offset)
		})
		t.Run("four bytes", func(t *testing.T) {
			r, offset, err := unescapePercentEncoding("F0%9F%8F%88")
			require.NoError(t, err)
			require.Equal(t, '🏈', r)
			require.Equal(t, 11, offset)
		})
	})
	t.Run("with invalid UTF8 sequence", func(t *testing.T) {
		t.Run("missing %", func(t *testing.T) {
			_, _, err := unescapePercentEncoding("F0")
			require.Error(t, err)
		})
		t.Run("missing digit", func(t *testing.T) {
			_, _, err := unescapePercentEncoding("F")
			require.Error(t, err)
		})
		t.Run("missing second %", func(t *testing.T) {
			_, _, err := unescapePercentEncoding("E3a3%Af")
			require.Error(t, err)
		})
		t.Run("missing second digit", func(t *testing.T) {
			_, _, err := unescapePercentEncoding("E3%a")
			require.Error(t, err)
		})
	})
	t.Run("missing third %", func(t *testing.T) {
		_, _, err := unescapePercentEncoding("F0%9F%88")
		require.Error(t, err)
	})
	t.Run("missing third digit", func(t *testing.T) {
		_, _, err := unescapePercentEncoding("F0%9F%8")
		require.Error(t, err)
	})
	t.Run("missing fourth %", func(t *testing.T) {
		_, _, err := unescapePercentEncoding("F0%9F%8F88")
		require.Error(t, err)
	})
	t.Run("missing fourth digit", func(t *testing.T) {
		_, _, err := unescapePercentEncoding("F0%9F%8F%8")
		require.Error(t, err)
	})
	t.Run("invalid hex digit", func(t *testing.T) {
		_, _, err := unescapePercentEncoding("F0%NF%8F%88")
		require.Error(t, err)
	})
	t.Run("invalid rune", func(t *testing.T) {
		_, _, err := unescapePercentEncoding("F0%9F%8F%01")
		require.Error(t, err)
	})
	t.Run("incomplete escape sequence (1)", func(t *testing.T) {
		_, _, err := unescapePercentEncoding("F0%9F")
		require.Error(t, err)
	})
	t.Run("incomplete escape sequence (2)", func(t *testing.T) {
		_, _, err := unescapePercentEncoding("F0%9FX")
		require.Error(t, err)
	})
	t.Run("bad rune in string", func(t *testing.T) {
		_, _, err := unescapePercentEncoding(string([]rune{utf8.RuneError}))
		require.Error(t, err)
	})
}

func TestUnhex(t *testing.T) {
	// edge case: not a hex digit (dev error)
	require.NotPanics(t, func() { _ = unhex('Z') })
}

func TestValidateUnreservedWithExtra(t *testing.T) {
	// edge case: invalid rune in string
	require.Error(t,
		validateUnreservedWithExtra(string([]rune{utf8.RuneError}), nil),
	)
}
