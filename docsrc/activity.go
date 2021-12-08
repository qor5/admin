package docsrc

import (
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	"github.com/theplant/htmlgo"
)

var Activity = Doc(
	htmlgo.Text("Activity provides a more flexible way to record the user's CRUD operation."),
	htmlgo.H2("Definition"),
	htmlgo.H5(`ActivityBuilder is used to register all model builders and add CRUD operation log.`),
	ch.Code(ActivityBuilder).Language("go"),
	htmlgo.H5(`ModelBuilder is mainly used to provide recording behavior when a model data is changed.`),
	ch.Code(ActivityModelBuilder).Language("go"),
	htmlgo.H5(`TypeHandler is used to process special types when diffing a modified data.`),
	ch.Code(ActivityTypeHandle).Language("go"),
	htmlgo.H5(`ActivityDefaultTypeHandles is the global type handler`),
	ch.Code(ActivityDefaultTypeHandles).Language("go"),
	htmlgo.H5(`DefaultIgnoredFields is the global default field that needs to be ignored when diffing a modified data.`),
	ch.Code(ActivityDefaultIgnoredFields).Language("go"),
	Markdown(`
## Usage

- Initalize the activity with the ~~~Activity~~~ method.

~~~go
activity := Activity().
	SetLogModel(&model.ActivityLog{}). // store activity log in model.ActivityLog
	SetDBContextKey("DB"). // set db context key
	SetCreatorContextKey("Creator") //set creator context key
~~~

- Register mutiple models with the ~~~RegisterModel~~~ method.

~~~go
activity.RegisterModels(&Post{},&Product{})
~~~

- Register a model with the ~~~RegisterModel~~~	 method.

~~~go
	activity.RegisterModel(&model.Page{}).
	SetKeys("VersionName"). // add keys
	SetLink(func(page interface{}) string {
		return fmt.Sprintf("/admin/pages/%d?version=%s", page.ID, page.VersionName)
	}). // set link
	AddIgnoredFields("ID", "Updatedat"). // ignore fields
	AddTypeHanders(...). // add type handlers
	SetAddExplicitly(ture) // will ingore this model when using the callback db

~~~

- Record a activity log

~~~go
	// record activity log from a context
	ctx := context.WithValue(context.Background(), 	DBContextKey, db)
	ctx  = ContextWithCreator(ctx, user)
	activity.AddRecords(ActivityEdit, ctx, newpage)

	// record activity log directly using known db and creator
	activity.AddEditRecord(user,newpage, db)
	activity.AddEditRecordWithOld(user,oldpage,newpage, db)

	// use db callback to automatically process the registered model
	activity.RegisterCallbackOnDB(db, "creator")
~~~	
	`),
	htmlgo.H2("Example"),
	ch.Code(ActivityExample).Language("go"),
).Title("Activity Log")
