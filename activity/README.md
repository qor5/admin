# Activity

## Usage

- Firstly, you should create an activity instance in your project.

  ```go
  activity := activity.New(db, currentUserFunc)
  ```

  - db (Required): The database where activity_logs is stored.
  - currentUserFunc (Required): You need to provide this method to get the creator.

- Register activity into presets

  ```go
  activityBuilder.Install(presetsBuilder)
  ```

- Register normal model or a `presets.ModelBuilder` into activity

  ```go
  activity.RegisterModel(normalModel) // It need you to record the activity log manually
  activity.RegisterModel(presetModel) // It will record the activity log automatically when you create, update or delete the model data via preset admin
  ```

- Skip recording activity log for preset model if you don't want to record the activity log automatically

  ```go
  activity.RegisterModel(presetModel).SkipCreate().SkipUpdate().SkipDelete()
  ```

- Configure more options for the `presets.ModelBuilder` to record more custom information

  ```go
  activity.RegisterModel(presetModel).UseDefaultTab() //use activity tab on the admin model edit page
  activity.RegisterModel(presetModel).AddKeys("ID", "Version") // will record value of the ID and Version field as the keyword of a model table
  activity.RegisterModel(presetModel).AddIgnoredFields("UpdateAt") // will ignore the UpdateAt field when recording activity log for update operation
  activity.RegisterModel(presetModel).AddTypeHanders(
    time.Time{},
    func(old, new any, prefixField string) []Diff {
  		oldString := old.(time.Time).Format(time.RFC3339)
  		newString := new.(time.Time).Format(time.RFC3339)
  		if oldString != newString {
  			return []Diff{
  				{Field: prefixField, Old: oldString, New: newString},
  			}
  		}
  		return []Diff{}
    }
    ) // you define your own type handler to record some custom type for update operation
  ```

- Record log manually when you use a normal model or save the model data via db directly

  - When a struct type only have one `activity.ModelBuilder`, you can use `activity` to record the log directly.

    ```go
      activity.OnEdit(ctx, old, new)
      activity.OnCreate(ctx, obj)
    ```

  - When a struct type have multiple `activity.ModelBuilder`, you need to get the corresponding `activity.ModelBuilder` and then use it to record the log.

    ```go
      activity.MustGetModelBuilder(presetModel1).OnEdit(ctx, old, new)
      activity.MustGetModelBuilder(presetModel2).OnCreate(ctx, obj)
    ```
