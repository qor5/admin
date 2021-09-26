package media

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/qor/serializable_meta"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	// set MediaLibraryURL to change the default url /system/{{class}}/{{primary_key}}/{{column}}.{{extension}}
	MediaLibraryURL = ""
)

func cropField(field *schema.Field, db *gorm.DB) (cropped bool) {
	if field.ReflectValueOf(db.Statement.ReflectValue).CanAddr() {
		// TODO Handle scanner
		if media, ok := field.ReflectValueOf(db.Statement.ReflectValue).Addr().Interface().(Media); ok && !media.Cropped() {
			option := parseTagOption(field.Tag.Get("mediaLibrary"))
			if MediaLibraryURL != "" {
				option.Set("url", MediaLibraryURL)
			}
			if media.GetFileHeader() != nil || media.NeedCrop() {
				var mediaFile FileInterface
				var err error
				if fileHeader := media.GetFileHeader(); fileHeader != nil {
					mediaFile, err = media.GetFileHeader().Open()
				} else {
					mediaFile, err = media.Retrieve(media.URL("original"))
				}

				if err != nil {
					db.AddError(err)
					return false
				}

				media.Cropped(true)

				if url := media.GetURL(option, db, field, media); url == "" {
					db.AddError(errors.New("invalid URL"))
				} else {
					result, _ := json.Marshal(map[string]string{"Url": url})
					media.Scan(string(result))
				}

				if mediaFile != nil {
					defer mediaFile.Close()
					var handled = false
					for _, handler := range mediaHandlers {
						if handler.CouldHandle(media) {
							mediaFile.Seek(0, 0)
							if db.AddError(handler.Handle(media, mediaFile, option)) == nil {
								handled = true
							}
						}
					}

					// Save File
					if !handled {
						db.AddError(media.Store(media.URL(), option, mediaFile))
					}
				}
				return true
			}
		}
	}
	return false
}

func saveAndCropImage(isCreate bool) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Error != nil {
			return
		}

		var updateColumns = map[string]interface{}{}
		// Handle SerializableMeta
		if value, ok := db.Statement.Dest.(serializable_meta.SerializableMetaInterface); ok {
			var (
				isCropped        bool
				handleNestedCrop func(record interface{})
			)

			handleNestedCrop = func(record interface{}) {
				// TODO
				newdb := db.Find(record)
				for _, field := range newdb.Statement.Schema.Fields {
					if cropField(field, db) {
						isCropped = true
						continue
					}

					if field.IndirectFieldType.Kind() == reflect.Struct {
						handleNestedCrop(field.ReflectValueOf(db.Statement.ReflectValue).Addr().Interface())
					}

					if field.IndirectFieldType.Kind() == reflect.Slice {
						for i := 0; i < reflect.Indirect(field.ReflectValueOf(db.Statement.ReflectValue).Addr()).Len(); i++ {
							handleNestedCrop(reflect.Indirect(field.ReflectValueOf(db.Statement.ReflectValue).Addr()).Index(i).Addr().Interface())
						}
					}
				}
			}

			record := value.GetSerializableArgument(value)
			handleNestedCrop(record)
			if isCreate && isCropped {
				updateColumns["value"], _ = json.Marshal(record)
			}
		}

		// Handle Normal Field
		for _, field := range db.Statement.Schema.Fields {
			if cropField(field, db) && isCreate {
				updateColumns[field.DBName] = field.ReflectValueOf(db.Statement.ReflectValue).Addr().Interface()
			}
		}

		if db.Error == nil && len(updateColumns) != 0 {
			db.AddError(db.Session(&gorm.Session{NewDB: true}).Model(db.Statement.Model).UpdateColumns(updateColumns).Error)
		}
	}
}

// RegisterCallbacks register callback into GORM DB
func RegisterCallbacks(db *gorm.DB) {
	if db.Callback().Create().Get("media:save_and_crop") == nil {
		db.Callback().Create().After("gorm:after_create").Register("media:save_and_crop", saveAndCropImage(true))
	}
	if db.Callback().Update().Get("media:save_and_crop") == nil {
		db.Callback().Update().Before("gorm:before_update").Register("media:save_and_crop", saveAndCropImage(false))
	}
}
