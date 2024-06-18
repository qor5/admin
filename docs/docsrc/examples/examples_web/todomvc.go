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
	"gorm.io/gorm"
)

func init() {
	compo.RegisterType((*TodoApp)(nil))
	compo.RegisterType((*TodoItem)(nil))
}

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
	b *TodoAppBuilder

	ID         string     `json:"id"`
	Visibility Visibility `json:"visibility"`
}

func (c *TodoApp) CompoName() string {
	return fmt.Sprintf("TodoApp:%s", c.ID)
}

func (c *TodoApp) InjectDep(f compo.Dep) {
	c.b = f.(*TodoAppBuilder)
}

func (c *TodoApp) MarshalHTML(ctx context.Context) ([]byte, error) {
	todos, err := StorageFromContext(ctx).List()
	if err != nil {
		return nil, err
	}

	filteredTodos := c.filteredTodos(todos)
	remaining := len(filterTodos(todos, func(todo *Todo) bool { return !todo.Completed }))

	filteredTodoItems := make([]h.HTMLComponent, len(filteredTodos))
	for i, todo := range filteredTodos {
		filteredTodoItems[i] = &TodoItem{
			b:    c.b,
			ID:   todo.ID,
			todo: todo,
			// OnChanged: compo.ReloadAction(ctx,a, nil).Go(),
		}
	}

	checkBoxID := fmt.Sprintf("%s-toggle-all", c.ID)
	return compo.Reloadify(c,
		web.Scope().Observer(NotifTodosChanged, compo.ReloadAction(ctx, c, nil).Go()),
		h.Section().Class("todoapp").Children(
			h.Header().Class("header").Children(
				h.H1("Todos"),
				h.Input("").Class("new-todo").Attr("id", fmt.Sprintf("%s-creator", c.ID)).
					Attr("placeholder", "What needs to be done?").
					Attr("@keyup.enter", strings.Replace(
						compo.PlaidAction(ctx, c, c.CreateTodo, &CreateTodoRequest{Title: "_placeholder_"}).Go(),
						`"_placeholder_"`,
						"$event.target.value",
						1,
					)),
			),
			h.Section().Class("main").Attr("v-show", h.JSONString(len(todos) > 0)).Children(
				h.Input("").Type("checkbox").Attr("id", checkBoxID).Class("toggle-all").
					Attr("checked", remaining == 0).
					Attr("@change", compo.PlaidAction(ctx, c, c.ToggleAll, nil).Go()),
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
							Attr("@click", compo.ReloadAction(ctx, c, func(cloned *TodoApp) {
								cloned.Visibility = VisibilityAll
							}).Go()),
					),
					h.Li(
						h.A(h.Text("Active")).ClassIf("selected", c.Visibility == VisibilityActive).
							Attr("@click", compo.ReloadAction(ctx, c, func(cloned *TodoApp) {
								cloned.Visibility = VisibilityActive
							}).Go()),
					),
					h.Li(
						h.A(h.Text("Completed")).ClassIf("selected", c.Visibility == VisibilityCompleted).
							Attr("@click", compo.ReloadAction(ctx, c, func(cloned *TodoApp) {
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
	s := StorageFromContext(ctx)

	todos, err := s.List()
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
		if err := s.Update(todo); err != nil {
			return r, err
		}
	}

	web.AppendRunScripts(&r, web.NotifyScript(NotifTodosChanged, nil))
	// compo.ApplyReloadToResponse(&r, a)
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

	if err := StorageFromContext(ctx).Create(&Todo{
		ID:        xid.New().String(),
		Title:     req.Title,
		Completed: false,
	}); err != nil {
		return r, err
	}
	web.AppendRunScripts(&r, web.NotifyScript(NotifTodosChanged, nil))
	// compo.ApplyReloadToResponse(&r, a)
	return
}

type TodoItem struct {
	b *TodoAppBuilder

	ID   string `json:"id"`
	todo *Todo  // use this if not nil, otherwise load with ID from Storage
	// OnChanged string `json:"on_changed"`
}

// func (c *TodoItem) CompoName() string {
// 	return fmt.Sprintf("TodoItem:%s", c.ID)
// }

func (c *TodoItem) InjectDep(f compo.Dep) {
	c.b = f.(*TodoAppBuilder)
}

func (c *TodoItem) MarshalHTML(ctx context.Context) ([]byte, error) {
	todo := c.todo
	if todo == nil {
		var err error
		todo, err = StorageFromContext(ctx).Read(c.ID)
		if err != nil {
			return nil, err
		}
	}

	var itemTitleCompo h.HTMLComponent
	if c.b != nil && c.b.itemTitleCompo != nil {
		itemTitleCompo = c.b.itemTitleCompo(todo)
	} else {
		itemTitleCompo = h.Label(todo.Title)
	}
	return h.Li().ClassIf("completed", todo.Completed).Children(
		h.Div().Class("view").Children(
			h.Input("").Type("checkbox").Class("toggle").Attr("checked", todo.Completed).
				Attr("@change", compo.PlaidAction(ctx, c, c.Toggle, nil).Go()),
			itemTitleCompo,
			h.Button("").Class("destroy").
				Attr("@click", compo.PlaidAction(ctx, c, c.Remove, nil).Go()),
		),
	).MarshalHTML(ctx)
}

func (c *TodoItem) Toggle(ctx context.Context) (r web.EventResponse, err error) {
	s := StorageFromContext(ctx)

	todo, err := s.Read(c.ID)
	if err != nil {
		return r, err
	}

	todo.Completed = !todo.Completed
	if err := s.Update(todo); err != nil {
		return r, err
	}

	web.AppendRunScripts(&r, web.NotifyScript(NotifTodosChanged, nil))
	// r.RunScript = t.OnChanged
	return
}

func (c *TodoItem) Remove(ctx context.Context) (r web.EventResponse, err error) {
	if err := StorageFromContext(ctx).Delete(c.ID); err != nil {
		return r, err
	}

	web.AppendRunScripts(&r, web.NotifyScript(NotifTodosChanged, nil))
	// r.RunScript = t.OnChanged
	return
}

type TodoAppBuilder struct {
	*compo.CompoDep[*TodoApp]
	itemTitleCompo func(todo *Todo) h.HTMLComponent
}

func NewTodoAppBuilder(initial *TodoApp) *TodoAppBuilder {
	b := &TodoAppBuilder{}
	b.CompoDep = compo.NewDep(b, initial)
	return b
}

func (b *TodoAppBuilder) ItemTitleCompo(f func(todo *Todo) h.HTMLComponent) *TodoAppBuilder {
	b.itemTitleCompo = f
	return b
}

func TodoMVCExample(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = h.Div().Style("display: flex; justify-content: center;").Children(
		h.Div().Style("width: 550px; margin-right: 40px;").Children(
			// &TodoApp{
			// 	ID:         "TodoApp0",
			// 	Visibility: VisibilityAll,
			// },
			NewTodoAppBuilder(&TodoApp{
				ID:         "TodoApp0",
				Visibility: VisibilityAll,
			}),
		),
		h.Div().Style("width: 550px;").Children(
			// &TodoApp{
			// 	ID:         "TodoApp1",
			// 	Visibility: VisibilityCompleted,
			// },
			NewTodoAppBuilder(&TodoApp{
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

type storageCtxKey struct{}

func StorageFromContext(ctx context.Context) Storage {
	s, ok := ctx.Value(storageCtxKey{}).(Storage)
	if ok {
		return s
	}
	evCtx := web.MustGetEventContext(ctx)
	return evCtx.ContextValue(storageCtxKey{}).(Storage)
}

var TodoMVCExamplePB = web.Page(TodoMVCExample).
	Wrap(func(in web.PageFunc) web.PageFunc {
		return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
			// TODO: 模拟最上层给到数据源，其实也可通过 Builder 来处理
			ctx.WithContextValue(storageCtxKey{}, memoryStorage)
			return in(ctx)
		}
	}).
	WrapEventFunc(func(in web.EventFunc) web.EventFunc {
		return func(ctx *web.EventContext) (r web.EventResponse, err error) {
			ctx.WithContextValue(storageCtxKey{}, memoryStorage)
			return in(ctx)
		}
	})

var TodoMVCExamplePath = examples.URLPathByFunc(TodoMVCExample)

type Storage interface {
	List() ([]*Todo, error)
	Create(todo *Todo) error
	Read(id string) (*Todo, error)
	Update(todo *Todo) error
	Delete(id string) error
}

var memoryStorage = &MemoryStorage{}

type MemoryStorage struct {
	mu    sync.RWMutex
	todos []*Todo
}

func (m *MemoryStorage) List() ([]*Todo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return compo.MustClone(m.todos), nil
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
			return compo.MustClone(todo), nil
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
