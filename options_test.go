package uri

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	t.Run("with default validation flags", func(t *testing.T) {
		o, redeem := applyURIOptions([]Option{})
		defer func() { redeem(o) }()

		/* TODO remove
		t.Logf("flagValidateScheme=%d", flagValidateScheme)
		t.Logf("flagValidateHost=%d", flagValidateHost)
		t.Logf("flagValidatePort=%d", flagValidatePort)
		t.Logf("flagValidateUserInfo=%d", flagValidateUserInfo)
		t.Logf("flagValidatePath=%d", flagValidatePath)
		t.Logf("flagValidateQuery=%d", flagValidateQuery)
		t.Logf("flagValidateFragment=%d", flagValidateFragment)
		*/

		require.True(t, o.validationFlags&flagValidateScheme > 0)
		require.True(t, o.validationFlags&flagValidateFragment > 0)
	})
}
