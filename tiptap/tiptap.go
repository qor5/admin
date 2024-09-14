package tiptap

import (
	"context"
	"fmt"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type TiptapEditorBuilder struct {
	editor          *vx.VXTiptapEditorBuilder
	db              *gorm.DB
	name            string
	imageGlueExists bool
	label           string
	errorMessages   []string
}

func TiptapEditor(db *gorm.DB, name string) (r *TiptapEditorBuilder) {
	r = &TiptapEditorBuilder{
		editor: vx.VXTiptapEditor(),
		db:     db,
		name:   name,
	}
	return
}

func (b *TiptapEditorBuilder) Label(v string) (r *TiptapEditorBuilder) {
	// b.editor.Label(v)
	b.label = v
	return b
}

func (b *TiptapEditorBuilder) ErrorMessages(v []string) (r *TiptapEditorBuilder) {
	b.errorMessages = v
	return b
}

func (b *TiptapEditorBuilder) Attr(vs ...any) (r *TiptapEditorBuilder) {
	b.editor.Attr(vs...)
	return b
}

func (b *TiptapEditorBuilder) Disabled(v bool) (r *TiptapEditorBuilder) {
	b.editor.Disabled(v)
	return b
}

func (b *TiptapEditorBuilder) Readonly(v bool) (r *TiptapEditorBuilder) {
	b.editor.Readonly(v)
	return b
}

func (b *TiptapEditorBuilder) Value(v string) (r *TiptapEditorBuilder) {
	b.editor.Value(v)
	return b
}

func (b *TiptapEditorBuilder) MarkdownTheme(v string) (r *TiptapEditorBuilder) {
	b.editor.MarkdownTheme(v)
	return b
}

func (b *TiptapEditorBuilder) Extensions(extensions []*vx.VXTiptapEditorExtension) (r *TiptapEditorBuilder) {
	if len(extensions) > 0 {
		imageGlue, exists := lo.Find(extensions, func(item *vx.VXTiptapEditorExtension) bool {
			return item.Name == "ImageGlue"
		})
		if exists {
			if imageGlue.Options == nil {
				imageGlue.Options = map[string]any{}
			}
			imageGlue.Options["onClick"] = `({editor, value, window})=> {
				window.document.getElementById("chooseFile").click();
				window.__currentImageGlueCallback = (images) => {
					if (!Array.isArray(images)) {
						images = [images]
					}
					for (const image of images) {
						editor.chain().focus().setImage({
							...value,
							display: 'block', // 'block' 'inline' 'left' 'right'
							src: image.Url,
							alt: image.FileName,
							width: image.Width,
							// height: image.Height,
							height: value.lockAspectRatio ? undefined : image.Height,
						}).run()
					}
				};
			}`
			b.imageGlueExists = true
		}
	}
	b.editor.Extensions(extensions)
	return b
}

func (b *TiptapEditorBuilder) MarshalHTML(ctx context.Context) ([]byte, error) {
	if !b.imageGlueExists {
		return b.editor.MarshalHTML(ctx)
	}

	fieldName := fmt.Sprintf("%s_tiptapeditor_medialibrary", b.name)

	r := h.Div().Class("d-flex").Children(
		h.Label(b.label).Class("v-label theme--light"),
		b.editor,
		// Body_tiptapeditor_medialibrary.Description: ""
		// Body_tiptapeditor_medialibrary.Values: "{"ID":1,"Url":"/system/media_libraries/1/file.jpeg","VideoLink":"","FileName":"main-qimg-d2290767bcbc9eb9748ca82934e6855c-lq.jpeg","Description":"","FileSizes":{"@qor_preview":20659,"default":73467,"original":73467},"Width":602,"Height":602}"
		h.Div().Class("hidden-screen-only").Children(
			media.QMediaBox(b.db).FieldName(fieldName).
				Value(&media_library.MediaBox{}).Config(&media_library.MediaBoxConfig{
				AllowType: "image",
			}),
		).Attr("v-on-mounted", fmt.Sprintf(`({watch,window}) => {
			let ignoreFlag = false;
			watch(() => form[%q], (val) => {
				if (ignoreFlag) {
					return
				}
				if (!!val) {
					const images = JSON.parse(val);
					if (window.__currentImageGlueCallback) {
						window.__currentImageGlueCallback(images);
					}
				}
				ignoreFlag = true;
				delete(form[%q]);
				ignoreFlag = false;
			}, { immediate: true })
		}`, fieldName+".Values", fieldName+".Values")),
	)
	return r.MarshalHTML(ctx)
}

func TiptapExtensions() []*vx.VXTiptapEditorExtension {
	extensions := vx.TiptapExtensions()
	image, exists := lo.Find(extensions, func(item *vx.VXTiptapEditorExtension) bool {
		return item.Name == "Image"
	})
	if exists {
		image.Name = "ImageGlue"
		image.Options = nil
	}
	return extensions
}
