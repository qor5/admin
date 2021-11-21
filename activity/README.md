# Activity

## Usage

- Initalize the activity with the `Activity` method.

  ```go
  activity := Activity().
      SetLogModel(&model.ActivityLog{}). // store activity log in model.ActivityLog
      SetDBContextKey("DB"). // set db context key
      SetCreatorContextKey("Creator") //set creator context key
  ```

- Register a model with the `RegisterModel` method.

  ```go
    activity.RegisterModel(&model.Page{}).
      AddKeys("VersionName"). // add keys
      SetLink(func(page interface{}) string {
  	    return fmt.Sprintf("/admin/pages/%d?version=%s", page.ID, page.VersionName)
      }). // set link
      AddIgnoredFields("ID", "Updatedat"). // ignore fields
      AddTypeHanders(...). // add type handlers
  ```

- Record a activity log

  ```go
   // record activity log from a context
    ctx := context.WithValue(context.Background(), 	DBContextKey, db)
    ctx  = ContextWithCreator(ctx, user)
    activity.AddRecords(ActivityEdit, ctx,oldpage, newpage)

    // record activity log directly using known db and creator
    activity.AddEditRecord(user,oldpage, newpage, db)

    // use db callback to automatically process the registered model
    activity.RegisterCallbackOnDB(db, "creator")
  ```

- Use the admin page to view the activity log

  ```go
  activity.ConfigureAdmin(preset,db)
  ```
