# Activity

## Usage

- Initalize the activity with the `Activity` method.

  ```go
  activity := Activity().
      SetLogModel(&model.ActivityLog{}). // store activity log in model.ActivityLog
      SetDBContextKey("DB"). // set db context key
      SetCreatorContextKey("Creator") //set creator context key
  ```

- Register mutiple models with the `RegisterModel` method.

  ```go
  activity.RegisterModels(&Post{},&Product{})
  ```

- Register a model with the `RegisterModel` method.

  ```go
    activity.RegisterModel(&model.Page{}).
      SetKeys("VersionName"). // add keys
      SetLink(func(page interface{}) string {
  	    return fmt.Sprintf("/admin/pages/%d?version=%s", page.ID, page.VersionName)
      }). // set link
      AddIgnoredFields("ID", "Updatedat"). // ignore fields
      AddTypeHanders(...). // add type handlers
      DisableOnCallback(All). // disable Create,Edit,Delete on callback
      DisableOnCallback(Create, Delete)  // disable Create,Delete on callback

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

- Use the admin page to view the activity log

  ```go
  activity.ConfigureAdmin(preset,db)
  ```
