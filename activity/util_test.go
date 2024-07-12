package activity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvatarText(t *testing.T) {
	assert.Equal(t, strings.ToUpper(string([]rune("xxx")[0:1])), "X")
	assert.Equal(t, strings.ToUpper(string([]rune("Yxx")[0:1])), "Y")
	assert.Equal(t, strings.ToUpper(string([]rune("你好")[0:1])), "你")
	assert.Equal(t, strings.ToUpper(string([]rune("フィールド")[0:1])), "フ")
}
