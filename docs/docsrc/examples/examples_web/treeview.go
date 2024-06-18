package examples_web

import (
	"context"
	"fmt"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_web/compo"
	"github.com/qor5/web/v3"
	"github.com/rs/xid"
	h "github.com/theplant/htmlgo"
)

func init() {
	compo.RegisterType((*TreeItem)(nil))
}

type TreeItem struct {
	ID     string    `json:"id"`
	Model  *TreeNode `json:"model"`
	IsOpen bool      `json:"is_open"`
}

type TreeNode struct {
	Name     string      `json:"name"`
	Value    string      `json:"value"`
	Children []*TreeNode `json:"children"`
}

func (t *TreeItem) CompoName() string {
	return fmt.Sprintf("TreeItem:%s", t.ID)
}

func (t *TreeItem) MarshalHTML(ctx context.Context) ([]byte, error) {
	isFolder := t.IsFolder()
	div := h.Div().StyleIf("font-weight: bold;", isFolder).
		Children(
			h.Text(t.Model.Name),
			h.Iff(isFolder, func() h.HTMLComponent {
				return h.Span(fmt.Sprintf("[%s]", t.toggleSymbol()))
			}),
		)
	if isFolder {
		div.Attr("@click", compo.ReloadAction(ctx, t, func(cloned *TreeItem) {
			cloned.Toggle()
		}).Go())
	} else {
		div.Attr("@dblclick", compo.ReloadAction(ctx, t, func(cloned *TreeItem) {
			cloned.ChangeType()
		}).Go())
	}
	return compo.Reloadify(t,
		h.Li(
			div,
			h.Iff(t.IsOpen && isFolder, func() h.HTMLComponent {
				var childComponents h.HTMLComponents
				for _, child := range t.Model.Children {
					childComponents = append(childComponents, &TreeItem{
						ID:    fmt.Sprintf("%s/%s", t.ID, child.Value),
						Model: child,
					})
				}
				childComponents = append(childComponents,
					h.Li(h.Text("+")).Attr("@click", compo.ReloadAction(ctx, t, func(cloned *TreeItem) {
						cloned.AddChild()
					}).Go()),
				)
				return h.Ul(childComponents...)
			}),
		),
	).MarshalHTML(ctx)
}

func (t *TreeItem) IsFolder() bool {
	return t.Model.Children != nil && len(t.Model.Children) > 0
}

func (t *TreeItem) toggleSymbol() string {
	if t.IsOpen {
		return "-"
	}
	return "+"
}

func (t *TreeItem) Toggle() {
	t.IsOpen = !t.IsOpen
}

func (t *TreeItem) ChangeType() {
	if !t.IsFolder() {
		t.Model.Children = []*TreeNode{}
		t.AddChild()
		t.IsOpen = true
	}
}

func (t *TreeItem) AddChild() {
	t.Model.Children = append(t.Model.Children, &TreeNode{
		Name:  "new stuff",
		Value: xid.New().String(),
	})
}

func TreeViewExample(cx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = h.Components(
		&TreeItem{
			ID: "TreeItem0",
			Model: &TreeNode{
				Name:  "My Tree",
				Value: "root",
				Children: []*TreeNode{
					{Name: "hello", Value: "hello"},
					{Name: "world", Value: "world"},
					{
						Name:  "child folder",
						Value: "child-folder",
						Children: []*TreeNode{
							{
								Name:  "child folder",
								Value: "child-folder0",
								Children: []*TreeNode{
									{Name: "hello", Value: "hello1"},
									{Name: "world", Value: "world1"},
								},
							},
							{Name: "hello", Value: "hello2"},
							{Name: "world", Value: "world2"},
							{
								Name:  "child folder",
								Value: "child-folder1",
								Children: []*TreeNode{
									{Name: "hello", Value: "hello3"},
									{Name: "world", Value: "world3"},
								},
							},
						},
					},
				},
			},
		},
		h.Br(), h.Br(), h.Br(),
	)
	return
}

var TreeViewExamplePB = web.Page(TreeViewExample)

var TreeViewExamplePath = examples.URLPathByFunc(TreeViewExample)
