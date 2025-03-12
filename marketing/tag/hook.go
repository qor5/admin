package tag

type Hook[T any] func(next T) T

func ChainHookWith[T any](previousHook Hook[T], hooks ...Hook[T]) Hook[T] {
	if previousHook != nil {
		hooks = append(hooks, previousHook)
	}
	return ChainHook(hooks...)
}

// ChainHook returns a single hook function that chains the given hooks together.
// When the hook is called, each hook in the chain is called with the result of
// the previous hook as its argument. The last hook in the chain is called with
// the argument passed to the chained hook. If the chain is empty, this function
// returns nil.
//
// Example
//
//   hooks := []Hook[int]{  // Hooks to chain together
//     func(next int) int { return next * 2 }, // Multiply by 2
//     func(next int) int { return next + 1 }, // Add 1
//   }
//
//   chainedHook := ChainHook(hooks...)
//   result := chainedHook(0) // Run the hook chain with argument 0
//   fmt.Println(result)     // Output: 2

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
