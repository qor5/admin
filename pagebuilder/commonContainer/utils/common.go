package utils

import (
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	. "github.com/theplant/htmlgo"
)

var ButtonPresets = []string{"unset", "primary", "secondary", "success", "info", "warning", "error"}

var SpaceOptions = []string{"0", "10", "20", "30", "40", "50", "60", "70", "80", "90", "100"}

var VerticalAlign = []struct {
	Label string
	Value string
}{
	{Label: "top", Value: "justify-start"},
	{Label: "center", Value: "justify-center"},
	{Label: "bottom", Value: "justify-end"},
	{Label: "space-between", Value: "justify-between"},
}

var HorizontalAlign = []struct {
	Label string
	Value string
}{
	{Label: "left", Value: "items-start"},
	{Label: "center", Value: "items-center"},
	{Label: "right", Value: "items-end"},
}

var ImageWithTextVisibilityOptions = []struct {
	Label string
	Value string
}{
	{Label: "title", Value: "title"},
	{Label: "content", Value: "content"},
	{Label: "button", Value: "button"},
	{Label: "image", Value: "image"},
}

func TailwindContainerWrapper(classes string, comp ...HTMLComponent) HTMLComponent {
	return Div(comp...).
		Class("container-instance").ClassIf(classes, classes != "")
}

func TiptapExtensions(enabledExtensions ...string) []*vx.VXTiptapEditorExtension {
	extensions := []*vx.VXTiptapEditorExtension{
		{
			Name: "BaseKit",
			Options: map[string]any{
				"placeholder": map[string]any{
					"placeholder": "Enter some text...",
				},
			},
		},
		{
			Name: "Bold",
		},
		{
			Name: "Italic",
		},
		{
			Name: "Underline",
		},
		{
			Name: "Strike",
		},
		{
			Name: "Code",
			Options: map[string]any{
				"divider": true,
			},
		},
		{
			Name: "Heading",
		},
		{
			Name: "TextAlign",
			Options: map[string]any{
				"types": []string{"heading", "paragraph", "image"},
			},
		},
		{
			Name: "FontFamily",
		},
		{
			Name: "FontSize",
		},
		{
			Name: "Color",
		},
		{
			Name: "Highlight",
			Options: map[string]any{
				"divider": true,
			},
		},
		{
			Name: "SubAndSuperScript",
			Options: map[string]any{
				"divider": true,
			},
		},
		{
			Name: "BulletList",
		},
		{
			Name: "OrderedList",
			Options: map[string]any{
				"divider": true,
			},
		},
		{
			Name: "TaskList",
		},
		{
			Name: "Indent",
			Options: map[string]any{
				"divider": true,
			},
		},
		{
			Name: "Link",
		},
		{
			Name: "Image",
		},
		{
			Name: "Video",
			Options: map[string]any{
				"divider": true,
			},
		},
		{
			Name: "Table",
			Options: map[string]any{
				"divider": true,
			},
		},
		{
			Name: "Blockquote",
		},
		{
			Name: "HorizontalRule",
		},
		{
			Name: "CodeBlock",
			Options: map[string]any{
				"divider": true,
			},
		},
		{
			Name: "Clear",
		},
		{
			Name: "History",
			Options: map[string]any{
				"divider": true,
			},
		},
		// {
		// 	Name: "Fullscreen",
		// },
	}

	// Filter extensions if enabledExtensions is provided
	if len(enabledExtensions) > 0 {
		enabledMap := make(map[string]bool)
		for _, name := range enabledExtensions {
			enabledMap[name] = true
		}

		// Always include BaseKit
		enabledMap["BaseKit"] = true

		// Filter extensions based on enabledMap
		extensions = lo.Filter(extensions, func(ext *vx.VXTiptapEditorExtension, _ int) bool {
			return enabledMap[ext.Name]
		})
	}

	image, exists := lo.Find(extensions, func(item *vx.VXTiptapEditorExtension) bool {
		return item.Name == "Image"
	})
	if exists {
		image.Name = "ImageGlue"
		image.Options = nil
	}
	return extensions
}
