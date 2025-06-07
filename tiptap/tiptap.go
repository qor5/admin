package tiptap

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	vx "github.com/qor5/x/v3/ui/vuetifyx"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
)

type TiptapEditorBuilder struct {
	editor          *vx.VXTiptapEditorBuilder
	db              *gorm.DB
	uniqueKey       string
	imageGlueExists bool
	label           string
	classList       []string
	disabled        bool
	errorMessages   []string
}

func TiptapEditor(db *gorm.DB, uniqueKey string) (r *TiptapEditorBuilder) {
	r = &TiptapEditorBuilder{
		editor:    vx.VXTiptapEditor(),
		db:        db,
		uniqueKey: uniqueKey,
	}
	return
}

func (b *TiptapEditorBuilder) Class(v ...string) (r *TiptapEditorBuilder) {
	b.classList = append(b.classList, v...)
	return b
}

func (b *TiptapEditorBuilder) Label(v string) (r *TiptapEditorBuilder) {
	// b.editor.Label(v)
	b.label = v
	return b
}

func (b *TiptapEditorBuilder) ErrorMessages(v ...string) (r *TiptapEditorBuilder) {
	b.errorMessages = v
	return b
}

func (b *TiptapEditorBuilder) Attr(vs ...any) (r *TiptapEditorBuilder) {
	// check if class attribute is set
	for i := 0; i < len(vs)-1; i += 2 {
		if k, ok := vs[i].(string); ok && k == "class" {
			if v, ok := vs[i+1].(string); ok {
				// split class string and add to classList
				classes := strings.Fields(v)
				b.classList = append(b.classList, classes...)
			}
		}
	}
	b.editor.Attr(vs...)
	return b
}

func (b *TiptapEditorBuilder) SetAttr(k string, v interface{}) {
	// if set class attribute, add to classList
	if k == "class" {
		if classStr, ok := v.(string); ok {
			classes := strings.Fields(classStr)
			b.classList = append(b.classList, classes...)
		}
	}
	b.editor.SetAttr(k, v)
}

func (b *TiptapEditorBuilder) Disabled(v bool) (r *TiptapEditorBuilder) {
	b.disabled = v
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

			fieldName := fmt.Sprintf("%s_tiptapeditor_medialibrary", b.uniqueKey)
			imageGlue.Options["onClick"] = fmt.Sprintf(`({editor, value, window})=> {
				const el = window.document.getElementById(%q);
				if (!el) {
					return;
				}
				el.click();
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
			}`, media.ChooseFileButtonID(fieldName))
			b.imageGlueExists = true
		}
	}
	b.editor.Extensions(extensions)
	return b
}

func (b *TiptapEditorBuilder) MarshalHTML(ctx context.Context) ([]byte, error) {
	var mediaBox h.HTMLComponent
	if b.imageGlueExists {
		fieldName := fmt.Sprintf("%s_tiptapeditor_medialibrary", b.uniqueKey)
		// Body_tiptapeditor_medialibrary.Description: ""
		// Body_tiptapeditor_medialibrary.Values: "{"ID":1,"Url":"/system/media_libraries/1/file.jpeg","VideoLink":"","FileName":"main-qimg-d2290767bcbc9eb9748ca82934e6855c-lq.jpeg","Description":"","FileSizes":{"@qor_preview":20659,"default":73467,"original":73467},"Width":602,"Height":602}"
		mediaBox = h.Div().Class("hidden-screen-only").Children(
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
		}`, fieldName+".Values", fieldName+".Values"))
	}

	if len(b.errorMessages) > 0 && !b.disabled {
		b.editor.Attr("style", "border: 1px solid rgb(var(--v-theme-error));")
	} else {
		b.Class("border-thin")
	}

	if len(b.classList) > 0 {
		// remove duplicate classes
		classMap := make(map[string]bool)
		var uniqueClasses []string
		for _, class := range b.classList {
			if !classMap[class] {
				classMap[class] = true
				uniqueClasses = append(uniqueClasses, class)
			}
		}
		b.editor.Attr("class", strings.Join(uniqueClasses, " "))
	}

	r := h.Div().Class("d-flex flex-column ga-1").Children(
		h.Label(b.label).Class("v-label theme--light"),
		b.editor,
		mediaBox,
		h.Iff(len(b.errorMessages) > 0, func() h.HTMLComponent {
			var compos []h.HTMLComponent
			for _, errMsg := range b.errorMessages {
				compos = append(compos, h.Div().Attr("v-pre", true).Text(errMsg))
			}
			return h.Div().Class("d-flex flex-column ps-4 py-1 ga-1 text-caption").
				ClassIf("text-error", len(b.errorMessages) > 0 && !b.disabled).
				ClassIf("text-grey", b.disabled).Children(compos...)
		}),
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
