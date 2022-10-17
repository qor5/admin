# Activity

## Usage

- Firstly, you should create an activity instance in your project.

  ```go
  activity := activity.New(presetsBuilder, db, logModel)
  ```

  - presetsBuilder (Required), it is a instance of `presets.Builder` to register activity into presets.
  - db (Required), it is a global database instance. we will use it if we don't find the specific database instance in a operation.
  - logModel (Optional), you can use your own model table to record the activity log as long as the model implements the `ActivityLogModel` interface.

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
    func(old, now interface{}, prefixField string) []Diff {
  		oldString := old.(time.Time).Format(time.RFC3339)
  		nowString := now.(time.Time).Format(time.RFC3339)
  		if oldString != nowString {
  			return []Diff{
  				{Field: prefixField, Old: oldString, Now: nowString},
  			}
  		}
  		return []Diff{}
    }
    ) // you define your own type handler to record some custom type for update operation
  ```

- Record log manually when you use a normal model or save the model data via db directly

  - When a struct type only have one `activity.ModelBuilder`, you can use `activity` to record the log directly.

    ```go
      activity.AddRecords(ActivityEdit, ctx, record)
      activity.AddRecords(ActivityCreate, ctx, record)
    ```

  - When a struct type have multiple `activity.ModelBuilder`, you need to get the corresponding `activity.ModelBuilder` and then use it to record the log.

    ```go
      activity.MustGetModelBuilder(presetModel1).AddRecords(ActivityEdit, ctx, record)
      activity.MustGetModelBuilder(presetModel2).AddRecords(ActivityEdit, ctx, record)
    ```
