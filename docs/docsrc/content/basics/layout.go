package basics

import (
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"

	"github.com/qor5/admin/v3/docs/docsrc/generated"
)

var Layout = Doc(
	Markdown(`
Presets comes with a built-in layout that works out of the box.  
And there are some ways to customzie the layout/theme.
## Theme
Presets UI is based on [Vuetify](https://v2.vuetifyjs.com/en/), you can modify the Admin theme by configuring the [Vuetify options](https://v2.vuetifyjs.com/en/features/presets/#default-preset)
    `),
	ch.Code(generated.CustomizeVuetifyOptions).Language("go"),
	Markdown(`
you can also call Injector in AssetFunc to add meta, add custom HTML in HEAD and TAIL.
    `),
	ch.Code(generated.InjectAssetViaAssetFunc).Language("go"),
	Markdown(`
## Layout
You can change the entire layout via *LayoutFunc*. The default layout is https://github.com/qor5/admin/blob/1e97c0dd45615fb7593245575ab0fea4f98c58b3/presets/presets.go#L860-L969
### Plain Layout
And We provide [PlainLayout](https://github.com/qor5/admin/blob/1e97c0dd45615fb7593245575ab0fea4f98c58b3/presets/presets.go#L972) which has no UI content except necessary assets. 
It will be helpful when there are some pages completely independent of Presets layout but still need to be consistent with the Presets theme.
    `),
).Slug("basics/layout").Title("Layout")
