package utils

import "reflect"

func Contains(a interface{}, b interface{}) bool {
	if reflect.ValueOf(a).Len() == 0 {
		return false
	}
	if _, ok := b.(int); ok {
		if tempB, ok := a.([]int); ok {
			for _, v := range tempB {
				if b == v {
					return true
				}
			}
		}
	}

	if _, ok := b.(uint); ok {
		if tempB, ok := a.([]uint); ok {
			for _, v := range tempB {
				if b == v {
					return true
				}
			}
		}
	}

	if _, ok := b.(float32); ok {
		if tempB, ok := a.([]float32); ok {
			for _, v := range tempB {
				if b == v {
					return true
				}
			}
		}
	}

	if _, ok := b.(string); ok {
		if tempB, ok := a.([]string); ok {
			for _, v := range tempB {
				if b == v {
					return true
				}
			}
		}

	}

	return false
}
