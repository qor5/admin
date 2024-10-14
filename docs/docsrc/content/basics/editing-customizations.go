package basics

import (
	"github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	. "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_presets"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
)

var EditingCustomizations = Doc(
	Markdown(`
Editing an object will be always in a drawer popup. select which fields can edit for each model
by using the ~.Only~ func of ~EditingBuilder~, There are different ways to configure the type
of component that is used to do the editing.

`),
	utils.Anchor(H2(""), "Configure field for a single model"),
	Markdown(`
Use a customized component is as simple as add the extra asset to the preset instance.
And configure the component func on the field:
`),
	ch.Code(generated.PresetsEditingCustomizationDescriptionSample).Language("go"),
	utils.DemoWithSnippetLocation("Presets Editing Customization Description Field", examples.URLPathByFunc(examples_presets.PresetsEditingCustomizationDescription)+"/customers", generated.PresetsEditingCustomizationDescriptionSampleLocation),
	Markdown(`
- Added the redactor javascript and css component pack as an extra asset
- Configure the description field to use the component func that returns the ~tiptap.TipTapEditor~ component
- Set the field name and value of the component
`),
	utils.Anchor(H2(""), "Configure field type for all models"),
	Markdown(`
Set a global field type to component func like the following:
`),
	ch.Code(generated.PresetsEditingCustomizationFileTypeSample).Language("go"),
	utils.DemoWithSnippetLocation("Presets Editing Customization File Type", examples.URLPathByFunc(examples_presets.PresetsEditingCustomizationFileType)+"/products", generated.PresetsEditingCustomizationFileTypeSampleLocation),
	Markdown(`
- We define ~MyFile~ to actually be a string
- We set ~FieldDefaults~ for writing, which is the editing drawer popup to be a customized component
- The component show an img tag with the string as src if it's not empty
- The component add a file input for user to upload new file
- The ~SetterFunc~ is called before save the object, it uploads the file to transfer.sh, and get the url back,
  then set the value to ~MainImage~ field

With ~FieldDefaults~ we can write libraries that add customized type for different models to reuse. It can take care
of how to display the edit controls, and How to save the object.

`),
	utils.Anchor(H2(""), "Tabs"),
	Markdown(`
Tabs can be added by using ~AppendTabsPanelFunc~ func on ~EditingBuilder~:
`),
	ch.Code(generated.PresetsEditingCustomizationTabsSample).Language("go"),
	utils.DemoWithSnippetLocation("Presets Editing Customization Tabs", examples.URLPathByFunc(examples_presets.PresetsEditingCustomizationTabs)+"/companies", generated.PresetsEditingCustomizationTabsSampleLocation),
	utils.Anchor(H2(""), "Validation"),
	Markdown(`
Field level validation and display on field can be added by implement ~ValidateFunc~,
and set the ~web.ValidationErrors~ result:
`),
	ch.Code(generated.PresetsEditingCustomizationValidationSample).Language("go"),
	utils.DemoWithSnippetLocation("Presets Editing Customization Validation", examples.URLPathByFunc(examples_presets.PresetsEditingCustomizationValidation)+"/customers", generated.PresetsEditingCustomizationValidationSampleLocation),
	Markdown(`
- We validate the ~Name~ of the customer must be longer than 10
- If the error happens, If will show below the field

`),
).Title("Editing").
	Slug("presets-guide/editing-customizations")
