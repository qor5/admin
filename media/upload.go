package media

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	// set MediaLibraryURL to change the default url /system/{{class}}/{{primary_key}}/{{column}}.{{extension}}
	MediaLibraryURL = ""
)

func cropField(field *schema.Field, db *gorm.DB) (cropped bool) {

	if !field.ReflectValueOf(db.Statement.ReflectValue).CanAddr() {
		return
	}

	media, ok := field.ReflectValueOf(db.Statement.ReflectValue).Addr().Interface().(Media)
	if !ok {
		return
	}

	if media.Cropped() {
		return
	}

	option := parseTagOption(field.Tag.Get("mediaLibrary"))
	if MediaLibraryURL != "" {
		option.Set("url", MediaLibraryURL)
	}

	if media.GetFileHeader() == nil && !media.NeedCrop() {
		return
	}

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

	if mediaFile == nil {
		return
	}

	defer mediaFile.Close()
	var handled = false

	for _, handler := range mediaHandlers {
		if !handler.CouldHandle(media) {
			continue
		}

		mediaFile.Seek(0, 0)
		if db.AddError(handler.Handle(media, mediaFile, option)) == nil {
			handled = true
		}
	}

	// Save File
	if !handled {
		db.AddError(media.Store(media.URL(), option, mediaFile))
		return false
	}

	return true
}

func SaveUploadAndCropImage(db *gorm.DB, obj interface{}) (err error) {
	db = db.Model(obj).Save(obj)
	err = db.Error
	if err != nil {
		return
	}

	var updateColumns = map[string]interface{}{}

	for _, field := range db.Statement.Schema.Fields {
		if cropField(field, db) {
			updateColumns[field.DBName] = field.ReflectValueOf(db.Statement.ReflectValue).Addr().Interface()
		}
	}

	if len(updateColumns) == 0 {
		return
	}

	err = db.UpdateColumns(updateColumns).Error

	return
}
