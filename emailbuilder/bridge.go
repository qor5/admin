package emailbuilder

import (
	"context"
	"fmt"
	"strings"

	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
)

type UtilDialogPayloadType struct {
	Title     string
	ContentEl h.HTMLComponent
	LoadingOk string
	OkAction  string
}

func ShowDialogScript(portalName string, vxDialogConfig UtilDialogPayloadType) string {
	dialog := vx.VXDialog(vxDialogConfig.ContentEl).
		Persistent(true).
		Title(vxDialogConfig.Title).
		Attr(":loading-ok", vxDialogConfig.LoadingOk).
		Attr(":model-value", "true").
		Attr("@click:ok", vxDialogConfig.OkAction)

	html, err := dialog.MarshalHTML(context.Background())
	if err != nil {
		return ""
	}
	html = []byte(strings.TrimSpace(string(html)))

	return UpdateCompInPortal(portalName, html)
}

func UpdateCompInPortal(portalName string, comp interface{}) string {
	// script := fmt.Sprintf("const aa = vars.__window.__goplaid.portals[%q];updatePortalTemplate && updatePortalTemplate(%q);", portalName, comp)
	script1 := fmt.Sprintf("const { updatePortalTemplate } = vars.__window.__goplaid.portals[%q]; console.log(updatePortalTemplate);", portalName)
	script2 := script1 + fmt.Sprintf("updatePortalTemplate(%q)", comp)

	return script2
}
