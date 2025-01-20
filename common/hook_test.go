package common_test

import (
	"fmt"
	"testing"

	"github.com/qor5/admin/v3/common"
	"github.com/stretchr/testify/assert"
)

func TestChainHookWith(t *testing.T) {
	type CreateFunc func(input string) string

	baseFunc := func(input string) string {
		return fmt.Sprintf("Base: %s", input)
	}

	hook1 := func(next CreateFunc) CreateFunc {
		return func(input string) string {
			return fmt.Sprintf("Hook1 -> %s", next(input))
		}
	}

	hook2 := func(next CreateFunc) CreateFunc {
		return func(input string) string {
			return fmt.Sprintf("Hook2 -> %s", next(input))
		}
	}

	hook3 := func(next CreateFunc) CreateFunc {
		return func(input string) string {
			return fmt.Sprintf("Hook3 -> %s", next(input))
		}
	}

	combinedHook := common.ChainHookWith(nil, hook1, hook2, hook3)
	{
		finalFunc := combinedHook(baseFunc)
		result := finalFunc("input")

		expected := "Hook1 -> Hook2 -> Hook3 -> Base: input"
		assert.Equal(t, expected, result, "The hooks should execute in the correct order")
	}

	combinedHook = common.ChainHookWith(combinedHook, hook1, hook2, hook3)
	{
		finalFunc := combinedHook(baseFunc)
		result := finalFunc("input")

		expected := "Hook1 -> Hook2 -> Hook3 -> Hook1 -> Hook2 -> Hook3 -> Base: input"
		assert.Equal(t, expected, result, "The hooks should execute in the correct order")
	}

	combinedHook = common.ChainHookWith(combinedHook) // nothing to append
	{
		finalFunc := combinedHook(baseFunc)
		result := finalFunc("input")

		expected := "Hook1 -> Hook2 -> Hook3 -> Hook1 -> Hook2 -> Hook3 -> Base: input"
		assert.Equal(t, expected, result, "The hooks should execute in the correct order")
	}

	combinedHook = common.ChainHookWith[CreateFunc](nil)
	assert.Nil(t, combinedHook, "The hook should be nil if no hooks are added")
}
