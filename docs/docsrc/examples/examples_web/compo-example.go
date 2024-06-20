package examples_web

import (
	"context"
	"fmt"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_web/stateful"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

func init() {
	stateful.RegisterActionableType((*ChildCompo)(nil))
	stateful.RegisterActionableType((*SampleCompo)(nil))
}

type ChildCompo struct {
	ID string

	Email      string
	ClickExtra string
}

func (c *ChildCompo) CompoName() string {
	return fmt.Sprintf("ChildCompo:%s", c.ID)
}

func (c *ChildCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	return stateful.Reloadify(c,
		Text("I'm a child:  "),
		Br(),
		Text(fmt.Sprintf("EmailInChildCompo: %s", c.Email)),
		Br(),
		Button("ChangeEmailViaChildReloadSelf").Attr("@click",
			stateful.ReloadAction(ctx, c, func(cloned *ChildCompo) {
				cloned.Email += "-ChildSelfReloaded"
			}).Go(),
		),
		Br(),
		Button("ClickExtra").Attr("@click", c.ClickExtra),
	).MarshalHTML(ctx)
}

type SampleCompo struct {
	ID string

	ShowPre     bool
	EmailSuffix string

	ModelID string
}

func (c *SampleCompo) CompoName() string {
	return fmt.Sprintf("SampleCompo:%s", c.ID)
}

func (c *SampleCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	child := &ChildCompo{
		ID:    fmt.Sprintf("%s-child", c.ID),
		Email: fmt.Sprintf("%s@gmail.com-%s", c.ID, c.EmailSuffix),
		ClickExtra: stateful.ReloadAction(ctx, c, func(cloned *SampleCompo) {
			cloned.EmailSuffix += "-ClickedExtra"
		}).Go(),
	}
	return stateful.Reloadify(c,
		Iff(c.ShowPre, func() HTMLComponent {
			return Pre(JSONString(c))
		}),
		Button("SwitchShowPre").Attr("@click",
			stateful.ReloadAction(ctx, c, func(cloned *SampleCompo) {
				cloned.ShowPre = !cloned.ShowPre
			}).Go(),
		),
		Button("DeleteItem").Attr("@click",
			stateful.PlaidAction(ctx, c, c.OnDeleteItem, DeleteItemRequest{
				Extra: "extra",
			}).Go(),
		),
		Div().Style("border: 1px solid black; padding: 10px; margin: 10px;").Children(
			child,
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
		},
		Br(), Br(), Br(),
		&SampleCompo{
			ID:      "888",
			ModelID: "model888",
		},
	)
	return
}

var CompoExamplePB = web.Page(CompoExample)

var CompoExamplePath = examples.URLPathByFunc(CompoExample)
