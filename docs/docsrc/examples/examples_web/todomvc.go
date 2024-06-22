package examples_web

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/rs/xid"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const NotifTodosChanged = "NotifTodosChanged"

type Visibility string

const (
	VisibilityAll       Visibility = "all"
	VisibilityActive    Visibility = "active"
	VisibilityCompleted Visibility = "completed"
)

type Todo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TodoApp struct {
	dep *TodoAppDep `inject:""`

	ID         string     `json:"id"`
	Visibility Visibility `json:"visibility" query:"visibility"`
}

func (c *TodoApp) CompoName() string {
	return fmt.Sprintf("TodoApp:%s", c.ID)
}

func (c *TodoApp) MarshalHTML(ctx context.Context) ([]byte, error) {
	todos, err := c.dep.db.List()
	if err != nil {
		return nil, err
	}

	filteredTodos := c.filteredTodos(todos)
	remaining := len(filterTodos(todos, func(todo *Todo) bool { return !todo.Completed }))

	filteredTodoItems := make([]h.HTMLComponent, len(filteredTodos))
	for i, todo := range filteredTodos {
		filteredTodoItems[i] = stateful.MustApply(ctx, &TodoItem{
			ID:   todo.ID,
			todo: todo,
			// OnChanged: stateful.ReloadAction(ctx,a, nil).Go(),
		})
	}

	checkBoxID := fmt.Sprintf("%s-toggle-all", c.ID)
	return stateful.Reloadable(c,
		web.Scope().Observe(NotifTodosChanged, stateful.ReloadAction(ctx, c, nil).Go()),
		h.Section().Class("todoapp").Children(
			h.Header().Class("header").Children(
				h.H1("Todos"),
				h.Input("").Class("new-todo").Attr("id", fmt.Sprintf("%s-creator", c.ID)).
					Attr("placeholder", "What needs to be done?").
					Attr("@keyup.enter", strings.Replace(
						stateful.PostAction(ctx, c, c.CreateTodo, &CreateTodoRequest{Title: "_placeholder_"}).Go(),
						`"_placeholder_"`,
						"$event.target.value",
						1,
					)),
			),
			h.Section().Class("main").Attr("v-show", h.JSONString(len(todos) > 0)).Children(
				h.Input("").Type("checkbox").Attr("id", checkBoxID).Class("toggle-all").
					Attr("checked", remaining == 0).
					Attr("@change", stateful.PostAction(ctx, c, c.ToggleAll, nil).Go()),
				h.Label("Mark all as complete").Attr("for", checkBoxID),
				h.Ul().Class("todo-list").Children(filteredTodoItems...),
			),
			h.Footer().Class("footer").Attr("v-show", h.JSONString(len(todos) > 0)).Children(
				h.Span("").Class("todo-count").Children(
					h.Strong(fmt.Sprintf("%d", remaining)),
					h.Text(fmt.Sprintf(" %s left", pluralize(remaining, "item", "items"))),
				),
				h.Ul().Class("filters").Children(
					h.Li(
						h.A(h.Text("All")).ClassIf("selected", c.Visibility == VisibilityAll).
							Attr("@click", stateful.ReloadAction(ctx, c, func(cloned *TodoApp) {
								cloned.Visibility = VisibilityAll
							}).Go()),
					),
					h.Li(
						h.A(h.Text("Active")).ClassIf("selected", c.Visibility == VisibilityActive).
							Attr("@click", stateful.ReloadAction(ctx, c, func(cloned *TodoApp) {
								cloned.Visibility = VisibilityActive
							}).Go()),
					),
					h.Li(
						h.A(h.Text("Completed")).ClassIf("selected", c.Visibility == VisibilityCompleted).
							Attr("@click", stateful.ReloadAction(ctx, c, func(cloned *TodoApp) {
								cloned.Visibility = VisibilityCompleted
							}).Go()),
					),
				),
			),
		),
	).MarshalHTML(ctx)
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

func (c *TodoApp) filteredTodos(todos []*Todo) []*Todo {
	switch c.Visibility {
	case VisibilityActive:
		return filterTodos(todos, func(todo *Todo) bool { return !todo.Completed })
	case VisibilityCompleted:
		return filterTodos(todos, func(todo *Todo) bool { return todo.Completed })
	default:
		return todos
	}
}

func (c *TodoApp) ToggleAll(ctx context.Context) (r web.EventResponse, err error) {
	todos, err := c.dep.db.List()
	if err != nil {
		return r, err
	}

	allCompleted := true
	for _, todo := range todos {
		if !todo.Completed {
			allCompleted = false
			break
		}
	}
	for _, todo := range todos {
		todo.Completed = !allCompleted
		if err := c.dep.db.Update(todo); err != nil {
			return r, err
		}
	}

	web.AppendRunScripts(&r, web.NotifyScript(NotifTodosChanged, nil))
	// stateful.AppendReloadToResponse(&r, a)
	return
}

type CreateTodoRequest struct {
	Title string `json:"title"`
}

func (c *TodoApp) CreateTodo(ctx context.Context, req *CreateTodoRequest) (r web.EventResponse, err error) {
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		r.RunScript = "alert('title can not be empty')"
		return
	}

	if err := c.dep.db.Create(&Todo{
		ID:        xid.New().String(),
		Title:     req.Title,
		Completed: false,
	}); err != nil {
		return r, err
	}
	web.AppendRunScripts(&r, web.NotifyScript(NotifTodosChanged, nil))
	// stateful.AppendReloadToResponse(&r, a)
	return
}

type TodoItem struct {
	db  Storage     `inject:""` // try inject db directly
	dep *TodoAppDep `inject:""`

	ID   string `json:"id"`
	todo *Todo  // use this if not nil, otherwise load with ID from Storage
	// OnChanged string `json:"on_changed"`
}

func (c *TodoItem) MarshalHTML(ctx context.Context) ([]byte, error) {
	todo := c.todo
	if todo == nil {
		var err error
		todo, err = c.db.Read(c.ID)
		if err != nil {
			return nil, err
		}
	}

	var itemTitleCompo h.HTMLComponent
	if c.dep.itemTitleCompo != nil {
		itemTitleCompo = c.dep.itemTitleCompo(todo)
	} else {
		itemTitleCompo = h.Label(todo.Title)
	}
	return h.Li().ClassIf("completed", todo.Completed).Children(
		h.Div().Class("view").Children(
			h.Input("").Type("checkbox").Class("toggle").Attr("checked", todo.Completed).
				Attr("@change", stateful.PostAction(ctx, c, c.Toggle, nil).Go()),
			itemTitleCompo,
			h.Button("").Class("destroy").
				Attr("@click", stateful.PostAction(ctx, c, c.Remove, nil).Go()),
		),
	).MarshalHTML(ctx)
}

func (c *TodoItem) Toggle(ctx context.Context) (r web.EventResponse, err error) {
	todo, err := c.db.Read(c.ID)
	if err != nil {
		return r, err
	}

	todo.Completed = !todo.Completed
	if err := c.db.Update(todo); err != nil {
		return r, err
	}

	web.AppendRunScripts(&r, web.NotifyScript(NotifTodosChanged, nil))
	// r.RunScript = t.OnChanged
	return
}

func (c *TodoItem) Remove(ctx context.Context) (r web.EventResponse, err error) {
	if err := c.db.Delete(c.ID); err != nil {
		return r, err
	}

	web.AppendRunScripts(&r, web.NotifyScript(NotifTodosChanged, nil))
	// r.RunScript = t.OnChanged
	return
}

type TodoAppDep struct {
	db             Storage
	itemTitleCompo func(todo *Todo) h.HTMLComponent
}

const (
	InjectorTop = "top"
	InjectorSub = "top/sub"
)

func init() {
	stateful.RegisterActionableType(
		(*TodoApp)(nil),
		(*TodoItem)(nil),
	)

	stateful.RegisterInjector(InjectorTop)
	stateful.RegisterInjector(InjectorSub, stateful.WithParent(InjectorTop))

	stateful.MustProvide(InjectorTop,
		func() Storage {
			return &MemoryStorage{}
		},
		func(db Storage) *TodoAppDep {
			return &TodoAppDep{
				db:             db,
				itemTitleCompo: nil,
			}
		},
	)

	stateful.MustProvide(InjectorSub, func(db Storage) *TodoAppDep {
		return &TodoAppDep{
			db: db,
			itemTitleCompo: func(todo *Todo) h.HTMLComponent {
				if todo.Completed {
					return h.Label(todo.Title).Style("color: red;")
				}
				return h.Label(todo.Title).Style("color: green;")
			},
		}
	})
}

func TodoMVCExample(ctx *web.EventContext) (r web.PageResponse, err error) {
	r.Body = h.Div().Style("display: flex; justify-content: center;").Children(
		h.Div().Style("width: 550px; margin-right: 40px;").Children(
			stateful.MustInject(InjectorTop, stateful.SyncQuery(&TodoApp{
				ID:         "TodoApp0",
				Visibility: VisibilityAll,
			})),
		),
		h.Div().Style("width: 550px;").Children(
			stateful.MustInject(InjectorSub, &TodoApp{
				ID:         "TodoApp1",
				Visibility: VisibilityCompleted,
			}),
		),
	)
	ctx.Injector.HeadHTML(`
	<link rel="stylesheet" type="text/css" href="https://unpkg.com/todomvc-app-css@2.4.1/index.css">
	<style>
		body{
			max-width: 100%;
		}
	</style>
	`)
	return
}

var TodoMVCExamplePB = web.Page(TodoMVCExample)

var TodoMVCExamplePath = examples.URLPathByFunc(TodoMVCExample)

type Storage interface {
	List() ([]*Todo, error)
	Create(todo *Todo) error
	Read(id string) (*Todo, error)
	Update(todo *Todo) error
	Delete(id string) error
}

type MemoryStorage struct {
	mu    sync.RWMutex
	todos []*Todo
}

func (m *MemoryStorage) List() ([]*Todo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return stateful.MustClone(m.todos), nil
}

func (m *MemoryStorage) Create(todo *Todo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.todos = append(m.todos, todo)
	return nil
}

func (m *MemoryStorage) Read(id string) (*Todo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, todo := range m.todos {
		if todo.ID == id {
			return stateful.MustClone(todo), nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MemoryStorage) Update(todo *Todo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.todos {
		if t.ID == todo.ID {
			m.todos[i] = todo
			return nil
		}
	}
	return nil
}

func (m *MemoryStorage) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, todo := range m.todos {
		if todo.ID == id {
			m.todos = append(m.todos[:i], m.todos[i+1:]...)
			return nil
		}
	}
	return nil
}
