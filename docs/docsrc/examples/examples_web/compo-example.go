package examples_web

import (
	"context"
	"fmt"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

type ChildCompo struct {
	ID string

	Email        string
	ExtraContent string

	// ClickExtra string
	// clickExtra func(*ChildCompo) string
}

func (c *ChildCompo) PortalName() string {
	return fmt.Sprintf("ChildCompo:%s", c.ID)
}

func (c *ChildCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	return Div(
		Text("I'm a child:  "),
		Br(),
		Text(fmt.Sprintf("EmailInChildCompo: %s", c.Email)),
		Br(),
		Button("ChangeEmailViaChildReloadSelf").Attr("@click",
			ReloadAction(c, func(cloned *ChildCompo) {
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

	ModelId string
	ShowPre bool

	Child *ChildCompo
}

func (c *SampleCompo) PortalName() string {
	return fmt.Sprintf("SampleCompo:%s", c.ID)
}

func (c *SampleCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	// c.Child.ClickExtra = Reload(c, func(cloned *SampleCompo) {
	// 	cloned.Child.ExtraContent += "-ClickedExtra"
	// })
	// c.Child.clickExtra = func(child *ChildCompo) string {
	// 	return Reload(c, func(cloned *SampleCompo) {
	// 		cloned.Child.ExtraContent += "-ClickedExtra"
	// 	})
	// }
	return Div(
		Iff(c.ShowPre, func() HTMLComponent {
			return Pre(JSONString(c))
		}),
		Button("SwitchShowPre").Attr("@click",
			ReloadAction(c, func(cloned *SampleCompo) {
				cloned.ShowPre = !cloned.ShowPre
			}).Go(),
		),
		Button("DeleteItem").Attr("@click",
			PlaidAction(c, "DeleteItem", DeleteItemRequest{
				ModelId: c.ModelId,
			}).Go(),
		),
		Div().Style("border: 1px solid black; padding: 10px; margin: 10px;").Children(
			Reloadify(c.Child),
		),
		Button("ChangeEmailViaReloadSelf").Attr("@click",
			ReloadAction(c, func(cloned *SampleCompo) {
				cloned.Child.Email += "-ParentReloaded"
			}).Go(),
		),
		Button("ChangeEmailViaReloadChild").Attr("@click",
			ReloadAction(c.Child, func(cloned *ChildCompo) {
				cloned.Email += "-ChildReloaded"
			}).Go(),
		),
	).MarshalHTML(ctx)
}

type DeleteItemRequest struct {
	ModelId string
}

func (c *SampleCompo) OnDeleteItem(req DeleteItemRequest) (r web.EventResponse, err error) {
	r.RunScript = fmt.Sprintf("alert('Deleted item %s')", req.ModelId)
	return
}

func CompoExample(cx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = Components(
		Reloadify(&SampleCompo{
			ID:      "666",
			ModelId: "model666",
			Child: &ChildCompo{
				ID:    "child666",
				Email: "666@gmail.com",
			},
		}),
		Br(), Br(), Br(),
		Reloadify(&SampleCompo{
			ID:      "888",
			ModelId: "model888",
			Child: &ChildCompo{
				ID:    "child888",
				Email: "888@gmail.com",
			},
		}),
	)
	return
}

var CompoExamplePB = web.Page(CompoExample)

var CompoExamplePath = examples.URLPathByFunc(CompoExample)
