package examples_web

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

func init() {
	Install(web.Default)
}

func Install(b *web.Builder) {
	b.RegisterEventFunc(eventDispatchCompoAction, eventDispatchCompoActionHandler)
}

func Copy(dst, src any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func Clone[T any](src T) (T, error) {
	var dst T
	if err := Copy(&dst, src); err != nil {
		return dst, err
	}
	return dst, nil
}

func MustClone[T any](src T) T {
	dst, err := Clone(src)
	if err != nil {
		panic(err)
	}
	return dst
}

func PrettyJSONString(v interface{}) (r string) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	r = string(b)
	return
}

const (
	paramCompoAction = "__compo_action__"
)

type CompoAction struct {
	CompoType     string `json:"compo_type,omitempty"`
	Compo         string `json:"compo,omitempty"` // json string
	ActionName    string `json:"action_name,omitempty"`
	ActionPayload string `json:"action_payload,omitempty"` // json string
}

func PlaidAction(compo h.HTMLComponent, actionName string, actionPayload any) *web.VueEventTagBuilder {
	// TODO: 最好是使用一个特别的 json 序列化和反序列化方法，可以自动将类型带进去并且注册，并且这样的话 compo 内部就可以支持嵌入动态 HTMLComponent 了
	// TODO: 并且的话，也能让整个上下文在 chrome network panel 里面有绝对的可读性
	if compo != nil {
		registerType(compo)
	}
	if actionPayload != nil {
		registerType(actionPayload)
	}

	vs := url.Values{}
	vs.Set(paramCompoAction, PrettyJSONString(CompoAction{
		CompoType:     fmt.Sprintf("%T", compo),
		Compo:         h.JSONString(compo),
		ActionName:    actionName,
		ActionPayload: h.JSONString(actionPayload),
	}))
	return web.Plaid().EventFunc(eventDispatchCompoAction).StringQuery(vs.Encode())
}

const eventDispatchCompoAction = "__dispatchCompoAction__"

func eventDispatchCompoActionHandler(ctx *web.EventContext) (r web.EventResponse, err error) {
	var action CompoAction
	if err = json.Unmarshal([]byte(ctx.R.FormValue(paramCompoAction)), &action); err != nil {
		return r, errors.Wrap(err, "failed to unmarshal compo action")
	}

	compo, err := newInstance(action.CompoType)
	if err != nil {
		return r, err
	}

	err = json.Unmarshal([]byte(action.Compo), compo)
	if err != nil {
		return r, err
	}

	// 通过反射找到并调用自定义 action 方法 OnXXX
	actionMethod := reflect.ValueOf(compo).MethodByName("On" + action.ActionName)
	if actionMethod.IsValid() && actionMethod.Kind() == reflect.Func {
		// 检查返回值类型是否符合 OnXXXX(...) (r web.EventResponse, err error)
		actionMethodType := actionMethod.Type()
		if actionMethodType.NumOut() != 2 ||
			actionMethodType.Out(0) != reflect.TypeOf(web.EventResponse{}) ||
			actionMethodType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
			return r, fmt.Errorf("method On%s has incorrect signature", action.ActionName)
		}

		var result []reflect.Value

		// 方法支持 0 参数或者 1 个参数
		if actionMethodType.NumIn() == 0 {
			result = actionMethod.Call(nil)
		} else if actionMethodType.NumIn() == 1 {
			// 根据所需参数类型，反序列化 payload
			argType := actionMethodType.In(0)
			argValue := reflect.New(argType).Interface()
			err := json.Unmarshal([]byte(action.ActionPayload), &argValue)
			if err != nil {
				return r, fmt.Errorf("failed to unmarshal action payload to %T: %v", argValue, err)
			}
			result = actionMethod.Call([]reflect.Value{reflect.ValueOf(argValue).Elem()})
		} else {
			return r, fmt.Errorf("method On%s has incorrect number of arguments", action.ActionName)
		}

		if len(result) != 2 || !result[0].CanInterface() || !result[1].CanInterface() {
			return r, fmt.Errorf("abnormal action result %v", result)
		}
		r = result[0].Interface().(web.EventResponse)
		if result[1].IsNil() {
			return r, nil
		}
		err = result[1].Interface().(error)
		return r, err
	}

	switch action.ActionName {
	case compoActionReload:
		reloadable, ok := compo.(Reloadable)
		if !ok {
			return r, fmt.Errorf("%s cannot handle the default Reload action", action.CompoType)
		}
		return OnReload(reloadable)
	default:
		return r, fmt.Errorf("unknown component action %q for %s", action.ActionName, action.CompoType)
	}
}

// TODO: should use a singleton to handle this
var typeRegistry = make(map[string]reflect.Type)

// TODO: 线程安全
func registerType(v any) {
	t := reflect.TypeOf(v)
	typeRegistry[t.String()] = t
}

func newInstance(typeName string) (any, error) {
	if t, ok := typeRegistry[typeName]; ok {
		return reflect.New(t.Elem()).Interface(), nil
	}
	return nil, fmt.Errorf("type not found: %s", typeName)
}

// Reloadable
type Reloadable interface {
	h.HTMLComponent
	PortalName() string
}

type wrappedReloadable interface {
	Reloadable
	__isWrapped() bool
}

type portalWrapper[T Reloadable] struct {
	*web.PortalBuilder
	name  string
	compo T
}

func (p *portalWrapper[T]) PortalName() string {
	return p.name
}

func (p *portalWrapper[T]) __isWrapped() bool {
	return true
}

func Reloadify[T Reloadable](compo T) Reloadable {
	if wrapper, ok := any(compo).(wrappedReloadable); ok && wrapper.__isWrapped() {
		return wrapper
	}
	return &portalWrapper[T]{
		PortalBuilder: web.Portal(compo).Name(compo.PortalName()),
		name:          compo.PortalName(),
		compo:         compo,
	}
}

const (
	compoActionReload = "Reload"
)

func ReloadAction[T Reloadable](compo any, f func(cloned T)) *web.VueEventTagBuilder {
	var cloned T
	switch c := compo.(type) {
	case *portalWrapper[T]:
		cloned = MustClone(c.compo)
	case T:
		cloned = MustClone(c)
	default:
		panic(fmt.Sprintf("unsupported type %T", compo))
	}
	if f != nil {
		f(cloned)
	}
	return PlaidAction(cloned, compoActionReload, struct{}{})
}

// OnReload default action handler
func OnReload(compo Reloadable) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: compo.PortalName(),
		Body: compo,
	})
	return
}
