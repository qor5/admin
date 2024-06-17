package examples_web

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_web/compo"
	"github.com/qor5/web/v3"
	"github.com/rs/xid"
	h "github.com/theplant/htmlgo"
)

func init() {
	compo.RegisterType((*TodoApp)(nil))
	compo.RegisterType((*TodoItem)(nil))
}

// Global storage for todos
// TODO: 需要重构数据源，线程安全等等
var (
	todos     = []*Todo{}
	todosLock sync.RWMutex
)

type Todo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TodoApp struct {
	ID         string `json:"id"`
	Visibility string `json:"visibility"`
	EditedTodo *Todo  `json:"edited_todo"`
}

func (a *TodoApp) CompoName() string {
	return fmt.Sprintf("TodoApp:%s", a.ID)
}

func (a *TodoApp) MarshalHTML(ctx context.Context) ([]byte, error) {
	filteredTodos := a.FilteredTodos()
	remaining := a.Remaining()

	filterCompos := make([]h.HTMLComponent, len(filteredTodos))
	for i, todo := range filteredTodos {
		filterCompos[i] = &TodoItem{
			ID:        todo.ID,
			OnChanged: compo.ReloadAction(a, nil).Go(),
		}
	}

	return compo.Reloadify(a,
		h.Section().Class("todoapp").Children(
			h.Header().Class("header").Children(
				h.H1("Todos"),
				h.Input("").Class("new-todo").Attr("placeholder", "What needs to be done?").
					Attr("@keyup.enter", strings.Replace(
						compo.PlaidAction(a, a.AddTodo, &AddTodoRequest{Title: "_placeholder_"}).Go(),
						`"_placeholder_"`,
						"$event.target.value",
						1,
					)),
			),
			h.Section().Class("main").Attr("v-show", h.JSONString(len(todos) > 0)).
				Children(
					h.Input("").Type("checkbox").Attr("id", "toggle-all").Class("toggle-all").
						Attr("checked", remaining == 0).
						Attr("@change", compo.PlaidAction(a, a.ToggleAll, nil).Go()),
					h.Label("Mark all as complete").Attr("for", "toggle-all"),
					h.Ul().Class("todo-list").Children(filterCompos...),
				),
			h.Footer().Class("footer").Attr("v-show", h.JSONString(len(todos) > 0)).Children(
				h.Span("").Class("todo-count").Children(
					h.Strong(fmt.Sprintf("%d", remaining)),
					h.Text(fmt.Sprintf(" %s left", pluralize(remaining, "item", "items"))),
				),
				h.Ul().Class("filters").Children(
					h.Li(
						h.A(h.Text("All")).Attr("@click", compo.ReloadAction(a, func(cloned *TodoApp) {
							cloned.Visibility = "all"
						}).Go()),
					),
					h.Li(
						h.A(h.Text("Active")).Attr("@click", compo.ReloadAction(a, func(cloned *TodoApp) {
							cloned.Visibility = "active"
						}).Go()),
					),
					h.Li(
						h.A(h.Text("Completed")).Attr("@click", compo.ReloadAction(a, func(cloned *TodoApp) {
							cloned.Visibility = "completed"
						}).Go()),
					),
				),
			),
		),
	).MarshalHTML(ctx)
}

func (a *TodoApp) FilteredTodos() []*Todo {
	todosLock.RLock()
	defer todosLock.RUnlock()
	switch a.Visibility {
	case "active":
		return filterTodos(todos, func(todo *Todo) bool { return !todo.Completed })
	case "completed":
		return filterTodos(todos, func(todo *Todo) bool { return todo.Completed })
	default:
		return todos
	}
}

func (a *TodoApp) Remaining() int {
	todosLock.RLock()
	defer todosLock.RUnlock()
	return len(filterTodos(todos, func(todo *Todo) bool { return !todo.Completed }))
}

func (a *TodoApp) ToggleAll(ctx context.Context) (r web.EventResponse, err error) {
	todosLock.Lock()
	defer todosLock.Unlock()
	allCompleted := true
	for _, todo := range todos {
		if !todo.Completed {
			allCompleted = false
			break
		}
	}
	for _, todo := range todos {
		todo.Completed = !allCompleted
	}

	r.RunScript = compo.ReloadAction(a, nil).Go() // TODO: 需要直接反馈 portal 结果
	return
}

type AddTodoRequest struct {
	Title string `json:"title"`
}

func (a *TodoApp) AddTodo(ctx context.Context, req *AddTodoRequest) (r web.EventResponse, err error) {
	todosLock.Lock()
	defer todosLock.Unlock()
	todos = append(todos, &Todo{
		ID:        xid.New().String(),
		Title:     req.Title,
		Completed: false,
	})

	r.RunScript = compo.ReloadAction(a, nil).Go() // TODO: 需要直接反馈 portal 结果
	return
}

type TodoItem struct {
	ID        string `json:"id"`
	OnChanged string `json:"on_changed"`
}

func (t *TodoItem) CompoName() string {
	return fmt.Sprintf("TodoItem:%s", t.ID)
}

func (t *TodoItem) MarshalHTML(ctx context.Context) ([]byte, error) {
	todo := fetchTodo(t.ID)
	return compo.Reloadify(t,
		h.Iff(todo != nil, func() h.HTMLComponent {
			return h.Li().ClassIf("completed", todo.Completed).Children(
				h.Div().Class("view").Children(
					h.Input("").Type("checkbox").Class("toggle").Attr("checked", todo.Completed).
						Attr("@change", compo.PlaidAction(t, t.Toggle, nil).Go()),
					h.Label(todo.Title),
					h.Button("").Class("destroy").
						Attr("@click", compo.PlaidAction(t, t.Remove, nil).Go()),
				),
			)
		}),
	).MarshalHTML(ctx)
}

func fetchTodo(id string) *Todo {
	todosLock.RLock()
	defer todosLock.RUnlock()
	for _, v := range todos {
		if v.ID == id {
			return v
		}
	}
	return nil
}

func (t *TodoItem) Toggle(ctx context.Context) (r web.EventResponse, err error) {
	// TODO: 这样并不会线程安全
	todo := fetchTodo(t.ID)
	todo.Completed = !todo.Completed

	r.RunScript = t.OnChanged
	return
}

func (t *TodoItem) Remove(ctx context.Context) (r web.EventResponse, err error) {
	todosLock.Lock()
	defer todosLock.Unlock()
	index := -1
	for i, todo := range todos {
		if todo.ID == t.ID {
			index = i
			break
		}
	}
	if index != -1 {
		todos = append(todos[:index], todos[index+1:]...)
	}

	r.RunScript = t.OnChanged
	return
}

func filterTodos(todos []*Todo, predicate func(*Todo) bool) []*Todo {
	var result []*Todo
	for _, todo := range todos {
		if predicate(todo) {
			result = append(result, todo)
		}
	}
	return result
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func TodoMVCExample(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = h.Components(
		&TodoApp{
			ID:         "TodoApp0",
			Visibility: "all",
		},
		h.Br(), h.Br(), h.Br(),
	)
	ctx.Injector.HeadHTML(`
	<link rel="stylesheet" type="text/css" href="https://unpkg.com/todomvc-app-css@2.4.1/index.css">
		`)
	return
}

var TodoMVCExamplePB = web.Page(TodoMVCExample)

var TodoMVCExamplePath = examples.URLPathByFunc(TodoMVCExample)
