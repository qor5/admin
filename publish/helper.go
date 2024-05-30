package publish

import (
	"errors"
	"reflect"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
)

func wrapEventFuncWithShowError(f web.EventFunc) web.EventFunc {
	return func(ctx *web.EventContext) (web.EventResponse, error) {
		r, err := f(ctx)
		if err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
		}
		return r, nil
	}
}

func convertSlice(from any, to any) error {
	vFrom := reflect.ValueOf(from)
	vTo := reflect.ValueOf(to)

	if vFrom.Kind() == reflect.Pointer {
		vFrom = vFrom.Elem()
	}
	if vTo.Kind() == reflect.Pointer {
		vTo = vTo.Elem()
	}
	if vFrom.Kind() != reflect.Slice || vTo.Kind() != reflect.Slice {
		return errors.New("types are not slices")
	}

	elemType := vTo.Type().Elem()
	newSlice := reflect.MakeSlice(reflect.SliceOf(elemType), vFrom.Len(), vFrom.Cap())

	for i := 0; i < vFrom.Len(); i++ {
		currentElem := vFrom.Index(i)

		// If the source is an interface and the destination is not directly assignable
		if currentElem.Kind() == reflect.Interface {
			actualElem := currentElem.Elem() // Get the underlying value from the interface
			if actualElem.Type().AssignableTo(elemType) {
				newSlice.Index(i).Set(actualElem) // Assign directly if assignable
			} else {
				// Try to convert if possible
				if actualElem.CanConvert(elemType) {
					newSlice.Index(i).Set(actualElem.Convert(elemType))
				} else {
					return errors.New("element type does not match the required type and cannot be converted")
				}
			}
		} else {
			if currentElem.Type().Implements(elemType) {
				newSlice.Index(i).Set(currentElem)
			} else {
				return errors.New("element type does not implement interface")
			}
		}
	}
	vTo.Set(newSlice)
	return nil
}
