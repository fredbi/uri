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

				nuri := auri.WithUserInfo(test.name).WithHost("newdomain.com").WithScheme("http").WithPort("443")
				assert.Equal(t, "//"+test.name+"@newdomain.com:443", nuri.Authority().String())
				assert.Equal(t, "443", nuri.Authority().Port())
				val := nuri.String()

				assert.Equalf(t, val, test.uriChanged,
					"val: %#v", val,
					"test: %#v", test.uriChanged,
					"values don't match: %v != %v (actual: %#v, expected: %#v)", val, test.uriChanged,
				)
				assert.Equal(t, "http", nuri.Scheme())

				nuri = nuri.WithPath("/abcd")
				assert.Equal(t, "/abcd", nuri.Authority().Path())

				nuri = nuri.WithQuery("a=b&x=5").WithFragment("chapter")
				assert.Equal(t, url.Values{"a": []string{"b"}, "x": []string{"5"}}, nuri.Query())
				assert.Equal(t, "chapter", nuri.Fragment())
				assert.Equal(t, test.uriChanged+"/abcd?a=b&x=5#chapter", nuri.String())
				assert.Equal(t, test.uriChanged+"/abcd?a=b&x=5#chapter", nuri.String())
			})
		}
	})

	t.Run("when building from scratch", func(t *testing.T) {
		u, err := Parse("http:")
		require.NoError(t, err)

		require.Empty(t, u.Authority())
		assert.Equal(t, "", u.Authority().UserInfo())

		v := u.WithUserInfo("user:pwd").WithHost("newdomain").WithPort("444")
		assert.Equal(t, "http://user:pwd@newdomain:444", v.String())
	})

	t.Run("when overriding with an invalid value", func(t *testing.T) {
		const uriRaw = "https://host:8080/a?query=value#fragment"

		u, err := Parse(uriRaw)
		require.NoError(t, err)

		u = u.WithPort("X8080")
		auth := u.Authority()
		require.Error(t, auth.Err())
	})
}
