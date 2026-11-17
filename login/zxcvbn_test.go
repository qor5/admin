package login

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userWithInputs struct{}

func (userWithInputs) PasswordStrengthUserInputs() []string {
	return []string{"o'brien", "alice@example.com", "</script>"}
}

func TestZxcvbnLogicComponent(t *testing.T) {
	t.Run("passes user inputs to zxcvbn", func(t *testing.T) {
		b, err := ZxcvbnLogicComponent(passwordStrengthUserInputs(userWithInputs{})...).MarshalHTML(context.Background())
		require.NoError(t, err)
		html := string(b)
		assert.Contains(t, html, `window.zxcvbn(x, ["o&#39;brien","alice@example.com","\u003c/script\u003e"])`)
	})

	t.Run("user without inputs keeps default behavior", func(t *testing.T) {
		b, err := ZxcvbnLogicComponent(passwordStrengthUserInputs(struct{}{})...).MarshalHTML(context.Background())
		require.NoError(t, err)
		assert.Contains(t, string(b), `window.zxcvbn(x, []).score`)
	})
}
