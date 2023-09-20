package uri

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Builder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		uri, uriChanged string
		name            string
	}{
		{
			"mailto://user@domain.com",
			"http://yolo@newdomain.com:443",
			"yolo",
		},
		{
			"https://user@domain.com",
			"http://yolo2@newdomain.com:443",
			"yolo2",
		},
	}

	t.Run("when building from existing URI", func(t *testing.T) {
		for _, toPin := range tests {
			test := toPin

			t.Run(fmt.Sprintf("change to %q", test.uriChanged), func(t *testing.T) {
				t.Parallel()

				auri, err := Parse(test.uri)
				require.NoErrorf(t, err,
					"failed to parse uri: %v", err,
				)

				nuri := auri.Builder().SetUserInfo(test.name).SetHost("newdomain.com").SetScheme("http").SetPort("443")
				zuri, ok := nuri.(URI)
				require.True(t, ok)
				assert.Equal(t, "//"+test.name+"@newdomain.com:443", zuri.Authority().String())
				assert.Equal(t, "443", nuri.URI().Authority().Port())
				val := nuri.String()

				assert.Equalf(t, val, test.uriChanged,
					"val: %#v", val,
					"test: %#v", test.uriChanged,
					"values don't match: %v != %v (actual: %#v, expected: %#v)", val, test.uriChanged,
				)
				assert.Equal(t, "http", nuri.URI().Scheme())

				_ = nuri.SetPath("/abcd")
				assert.Equal(t, "/abcd", nuri.URI().Authority().Path())

				_ = nuri.SetQuery("a=b&x=5").SetFragment("chapter")
				assert.Equal(t, url.Values{"a": []string{"b"}, "x": []string{"5"}}, nuri.URI().Query())
				assert.Equal(t, "chapter", nuri.URI().Fragment())
				assert.Equal(t, test.uriChanged+"/abcd?a=b&x=5#chapter", nuri.URI().String())
				assert.Equal(t, test.uriChanged+"/abcd?a=b&x=5#chapter", nuri.String())
			})
		}
	})

	t.Run("when building from scratch", func(t *testing.T) {
		u, _ := Parse("http:")
		b := u.Builder()

		require.Empty(t, u.Authority())
		assert.Equal(t, "", u.Authority().UserInfo())

		b = b.SetUserInfo("user:pwd").SetHost("newdomain").SetPort("444")
		assert.Equal(t, "http://user:pwd@newdomain:444", b.String())
	})
}
