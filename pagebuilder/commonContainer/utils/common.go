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

var ButtonAlign = []struct {
	Label string
	Value string
}{
	{Label: "left", Value: "text-left"},
	{Label: "center", Value: "text-center"},
	{Label: "right", Value: "text-right"},
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

var ImageRatioOptions = []struct {
	Label string
	Value string
}{
	{Label: "1:1", Value: "1/1"},
	{Label: "3:2", Value: "3/2"},
	{Label: "2:3", Value: "2/3"},
	{Label: "3:4", Value: "3/4"},
	{Label: "4:3", Value: "4/3"},
	{Label: "5:4", Value: "5/4"},
	{Label: "4:5", Value: "4/5"},
	{Label: "16:9", Value: "16/9"},
	{Label: "9:16", Value: "9/16"},
}

var CardListVisibilityOptions = []struct {
	Label string
	Value string
}{
	{Label: "title", Value: "title"},
	{Label: "image", Value: "image"},
	{Label: "productTitle", Value: "productTitle"},
	{Label: "description", Value: "description"},
}

var ImageWithTextImageWidthOptions = []struct {
	Label string
	Value string
}{
	{Label: "small", Value: "350px"},
	{Label: "medium", Value: "500px"},
	{Label: "large", Value: "700px"},
}

var ImageWithTextImageHeightOptions = []struct {
	Label string
	Value string
}{
	{Label: "adapt to image", Value: "auto"},
	{Label: "small", Value: "350px"},
	{Label: "medium", Value: "500px"},
	{Label: "large", Value: "700px"},
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
				"types": []string{"paragraph"},
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

		// If TextAlign is enabled, ensure Paragraph is included and handle Heading and Image
		if enabledMap["TextAlign"] {
			// TextAlign需要Paragraph，所以总是确保它被启用
			enabledMap["Paragraph"] = true

			// 准备TextAlign的types，默认包含paragraph
			textAlignTypes := []string{"paragraph"}

			// 如果启用了Heading，将它添加到types
			if enabledMap["Heading"] {
				textAlignTypes = append(textAlignTypes, "heading")
			}

			// 如果启用了Image，将它添加到types
			if enabledMap["Image"] {
				textAlignTypes = append(textAlignTypes, "image")
			}

			// 更新TextAlign的Options.types
			for i, ext := range extensions {
				if ext.Name == "TextAlign" {
					if ext.Options == nil {
						ext.Options = map[string]any{}
					}
					ext.Options["types"] = textAlignTypes
					extensions[i] = ext
					break
				}
			}
		}

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
