package examples_web

import (
	"context"
	"fmt"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_web/compo"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

func init() {
	compo.RegisterType((*ChildCompo)(nil))
	compo.RegisterType((*SampleCompo)(nil))
}

type ChildCompo struct {
	ID string

	Email        string
	ExtraContent string

	// ClickExtra string
	// clickExtra func(*ChildCompo) string
}

func (c *ChildCompo) CompoName() string {
	return fmt.Sprintf("ChildCompo:%s", c.ID)
}

func (c *ChildCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	return compo.Reloadify(c,
		Text("I'm a child:  "),
		Br(),
		Text(fmt.Sprintf("EmailInChildCompo: %s", c.Email)),
		Br(),
		Button("ChangeEmailViaChildReloadSelf").Attr("@click",
			compo.ReloadAction(c, func(cloned *ChildCompo) {
				cloned.Email += "-ChildSelfReloaded"
			}).Go(),
		),
		Br(),
		Text(c.ExtraContent),
		Br(),
		// Button("ClickExtra").Attr("@click", c.ClickExtra), // TODO: 这样信息不会丢失，但是貌似目前序列化有些问题
		// Button("ClickExtra").Attr("@click", c.clickExtra(c)),  // TODO: 只 reload child 的话，这个信息就会丢失了
	).MarshalHTML(ctx)
}

type SampleCompo struct {
	ID string

	ModelID string
	ShowPre bool

	Child *ChildCompo
}

func (c *SampleCompo) CompoName() string {
	return fmt.Sprintf("SampleCompo:%s", c.ID)
}

func (c *SampleCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	// c.Child.ClickExtra = ReloadAction(c, func(cloned *SampleCompo) {
	// 	cloned.Child.ExtraContent += "-ClickedExtra"
	// }).Go()
	// c.Child.clickExtra = func(child *ChildCompo) string {
	// 	return Reload(c, func(cloned *SampleCompo) {
	// 		cloned.Child.ExtraContent += "-ClickedExtra"
	// 	})
	// }
	return compo.Reloadify(c,
		Iff(c.ShowPre, func() HTMLComponent {
			return Pre(JSONString(c))
		}),
		Button("SwitchShowPre").Attr("@click",
			compo.ReloadAction(c, func(cloned *SampleCompo) {
				cloned.ShowPre = !cloned.ShowPre
			}).Go(),
		),
		Button("DeleteItem").Attr("@click",
			compo.PlaidAction(c, c.OnDeleteItem, DeleteItemRequest{
				Extra: "extra",
			}).Go(),
		),
		Div().Style("border: 1px solid black; padding: 10px; margin: 10px;").Children(
			c.Child,
		),
		Button("ChangeEmailViaReloadSelf").Attr("@click",
			compo.ReloadAction(c, func(cloned *SampleCompo) {
				cloned.Child.Email += "-ParentReloaded"
			}).Go(),
		),
		Button("ChangeEmailViaReloadChild").Attr("@click",
			compo.ReloadAction(c.Child, func(cloned *ChildCompo) {
				cloned.Email += "-ChildReloaded"
			}).Go(),
		),
	).MarshalHTML(ctx)
}

type DeleteItemRequest struct {
	Extra string
}

func (c *SampleCompo) OnDeleteItem(ctx context.Context, req DeleteItemRequest) (r web.EventResponse, err error) {
	r.RunScript = fmt.Sprintf("alert('Deleted item %s (%s)')", c.ModelID, req.Extra)
	return
}

func CompoExample(cx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = Components(
		&SampleCompo{
			ID:      "666",
			ModelID: "model666",
			Child: &ChildCompo{
				ID:    "child666",
				Email: "666@gmail.com",
			},
		},
		Br(), Br(), Br(),
		&SampleCompo{
			ID:      "888",
			ModelID: "model888",
			Child: &ChildCompo{
				ID:    "child888",
				Email: "888@gmail.com",
			},
		},
	)
	return
}

var CompoExamplePB = web.Page(CompoExample)

var CompoExamplePath = examples.URLPathByFunc(CompoExample)
