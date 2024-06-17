package compo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

func init() {
	Install(web.Default)
}

func Install(b *web.Builder) {
	b.RegisterEventFunc(eventDispatchAction, eventDispatchActionHandler)
}

type Action struct {
	CompoType string          `json:"compo_type,omitempty"`
	Compo     json.RawMessage `json:"compo,omitempty"` // json string
	Method    string          `json:"method,omitempty"`
	Request   json.RawMessage `json:"request,omitempty"` // json string
}

const fieldKeyAction = "__compo_action__"

func PlaidAction(c h.HTMLComponent, method any, request any) *web.VueEventTagBuilder {
	var methodName string
	switch m := method.(type) {
	case string:
		methodName = m
	default:
		methodName = GetFuncName(method)
	}

	return web.Plaid().
		EventFunc(eventDispatchAction).
		FieldValue(fieldKeyAction, web.Var(
			fmt.Sprintf(`JSON.stringify(%s, null, "\t")`,
				PrettyJSONString(Action{
					CompoType: fmt.Sprintf("%T", c),
					Compo:     json.RawMessage(h.JSONString(c)),
					Method:    methodName,
					Request:   json.RawMessage(h.JSONString(request)),
				}),
			),
		))
}

const eventDispatchAction = "__dispatch_compo_action__"

var (
	outType0 = reflect.TypeOf(web.EventResponse{})
	outType1 = reflect.TypeOf((*error)(nil)).Elem()
	inType0  = reflect.TypeOf((*context.Context)(nil)).Elem()
)

func eventDispatchActionHandler(ctx *web.EventContext) (r web.EventResponse, err error) {
	var action Action
	if err = json.Unmarshal([]byte(ctx.R.FormValue(fieldKeyAction)), &action); err != nil {
		return r, errors.Wrap(err, "failed to unmarshal compo action")
	}

	c, err := newInstance(action.CompoType)
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(action.Compo, c)
	if err != nil {
		return r, err
	}

	method := reflect.ValueOf(c).MethodByName(action.Method)
	if method.IsValid() && method.Kind() == reflect.Func {
		methodType := method.Type()
		if methodType.NumOut() != 2 ||
			methodType.Out(0) != outType0 ||
			methodType.Out(1) != outType1 {
			return r, errors.Errorf("action method %q has incorrect signature", action.Method)
		}

		numIn := methodType.NumIn()
		if numIn <= 0 || numIn > 2 {
			return r, errors.Errorf("action method %q has incorrect number of arguments", action.Method)
		}
		if methodType.In(0) != inType0 {
			return r, errors.Errorf("action method %q has incorrect signature", action.Method)
		}
		ctxValue := reflect.ValueOf(ctx.R.Context())

		params := []reflect.Value{ctxValue}
		if numIn == 2 {
			argType := methodType.In(1)
			argValue := reflect.New(argType).Interface()
			err := json.Unmarshal([]byte(action.Request), &argValue)
			if err != nil {
				return r, errors.Wrapf(err, "failed to unmarshal action request to %T", argValue)
			}
			params = append(params, reflect.ValueOf(argValue).Elem())
		}

		result := method.Call(params)
		if len(result) != 2 || !result[0].CanInterface() || !result[1].CanInterface() {
			return r, errors.Errorf("action method %q has incorrect return values", action.Method)
		}
		r = result[0].Interface().(web.EventResponse)
		if result[1].IsNil() {
			return r, nil
		}
		err = result[1].Interface().(error)
		return r, errors.Wrapf(err, "failed to call action method %q", action.Method)
	}

	switch action.Method {
	case actionMethodReload:
		rc, ok := c.(Reloadable)
		if !ok {
			return r, errors.Errorf("compo %T does not implement Reloadable", c)
		}
		return OnReload(rc)
	default:
		return r, errors.Errorf("action method %q not found", action.Method)
	}
}

// TODO: 最好是使用一个特别的 json 序列化和反序列化方法，可以自动将类型带进去并且注册，并且这样的话 compo 内部就可以支持嵌入动态 HTMLComponent 了
// TODO: 并且的话，也能让整个上下文在 chrome network panel 里面有绝对的可读性
var typeRegistry = cmap.New[reflect.Type]()

func RegisterType(v any) {
	typeRegistry.SetIfAbsent(fmt.Sprintf("%T", v), reflect.TypeOf(v))
}

func newInstance(typeName string) (any, error) {
	if t, ok := typeRegistry.Get(typeName); ok {
		return reflect.New(t.Elem()).Interface(), nil
	}
	return nil, errors.Errorf("type not found: %s", typeName)
}
