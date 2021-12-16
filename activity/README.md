# Activity

## Usage

- Initalize the activity with the `Activity` method.

  ```go
  activity := New(presetsBuilder, db, &model.ActivityLog{}).
      SetDBContextKey("DB"). // set db context key
      SetCreatorContextKey("Creator") //set creator context key
  ```

- Register mutiple models with the `RegisterModels` method.

  ```go
  activity.RegisterModels(postBuilder,productBuilder)
  ```

- Register a model with the `RegisterModel` method.

  ```go
    activity.RegisterModel(postBuilder).
      SetKeys("VersionName"). // add keys
      SetLink(func(page interface{}) string {
  	    return fmt.Sprintf("/admin/pages/%d?version=%s", page.ID, page.VersionName)
      }). // set link
      AddIgnoredFields("ID", "Updatedat"). // ignore fields
      AddTypeHanders(...). // add type handlers
      SkipCreate(). // skip Create
      SkipUpdate(). //  skip Update
      SkipDelete(). // skip Delete
      UseDefaultTab(). // use default tab

  ```

- Record a activity log

  ```go
   // record activity log from a context
    ctx := context.WithValue(context.Background(), 	DBContextKey, db)
    ctx  = ContextWithCreator(ctx, user)
    activity.AddRecords(ActivityEdit, ctx, newpage)

    // record activity log directly using known db and creator
    activity.AddEditRecord(user,newpage, db)
    activity.AddEditRecordWithOld(user,oldpage,newpage, db)

    // use db callback to automatically process the registered model
    activity.RegisterCallbackOnDB(db, "creator")
  ```

- Fetch the activity log

  ```go
    activity.GetActivityLogs(Post{ID: 1}, db) // use the default log model
    activity.GetCustomizeActivityLogs(Post{ID: 1}, db) // use the customize log model
  ```
