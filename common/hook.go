package common

type Hook[T any] func(next T) T

func ChainHookWith[T any](previousHook Hook[T], hooks ...Hook[T]) Hook[T] {
	if previousHook != nil {
		hooks = append(hooks, previousHook)
	}
	return ChainHook(hooks...)
}

func ChainHook[T any](hooks ...Hook[T]) Hook[T] {
	if len(hooks) == 0 {
		return nil
	}
	if len(hooks) == 1 {
		return hooks[0]
	}
	return func(next T) T {
		for i := len(hooks); i > 0; i-- {
			next = hooks[i-1](next)
		}
		return next
	}
}
