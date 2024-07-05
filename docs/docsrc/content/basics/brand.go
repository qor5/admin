package basics

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_presets"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	"github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var Brand = Doc(
	Markdown(`
Brand refers to the top area of the left menu bar, we provide two functions ~BrandTitle~ and ~BrandFunc~ to customize it.

## Simple customization
If you want only to change the brand string, you can use ~BrandTitle~ to set the string, the string will be displayed in the brand area with ~<H1>~ tag.
`),

	ch.Code(generated.BrandTitleSample).Language("go"),
	utils.DemoWithSnippetLocation("Brand Title", examples.URLPathByFunc(examples_presets.PresetsBrandTitle)+"/brands", generated.BrandTitleSampleLocation),

	Markdown(`
## Full customization
When you opt-in to full brand customization, you can use ~BrandFunc~ to be responsible for drawing for the entire brand area, such as you can put your own logo image in it.
`),

	ch.Code(generated.BrandFuncSample).Language("go"),
	utils.DemoWithSnippetLocation("Brand Func", examples.URLPathByFunc(examples_presets.PresetsBrandFunc)+"/brands", generated.BrandFuncSampleLocation),

	Markdown(`
## Profile
Profile is below the brand area, where you can put the current user's information or others. We provide ~ProfileFunc~ to customize it.
`),

	ch.Code(generated.ProfileSample).Language("go"),
	utils.DemoWithSnippetLocation("Profile", examples.URLPathByFunc(examples_presets.PresetsProfile)+"/brands", generated.ProfileSampleLocation),
).Title("Brand").
	Slug("basics/brand")
