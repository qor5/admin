package base

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// set MediaLibraryURL to change the default url /system/{{class}}/{{primary_key}}/{{column}}.{{extension}}
var MediaLibraryURL = ""

func cropField(field *schema.Field, db *gorm.DB) (cropped bool, err error) {
	if !field.ReflectValueOf(db.Statement.Context, db.Statement.ReflectValue).CanAddr() {
		return
	}

	media, ok := field.ReflectValueOf(db.Statement.Context, db.Statement.ReflectValue).Addr().Interface().(Media)
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
	if fileHeader := media.GetFileHeader(); fileHeader != nil {
		mediaFile, err = media.GetFileHeader().Open()
	} else {
		mediaFile, err = media.Retrieve(media.URL("original"))
	}

	if err != nil {
		return false, err
	}
	// TODO: this is a defensive condition. probably not needed anymore
	if mediaFile == nil {
		return false, errors.New("can't find mediaFile")
	}
	defer mediaFile.Close()
	media.Cropped(true)

	if url := media.GetURL(option, db, field, media); url == "" {
		return false, errors.New("invalid URL")
	} else {
		result, _ := json.Marshal(map[string]string{"Url": url})
		media.Scan(string(result))
	}

	handled := false

	for _, handler := range mediaHandlers {
		if !handler.CouldHandle(media) {
			continue
		}

		mediaFile.Seek(0, 0)
		if handler.Handle(media, mediaFile, option) == nil {
			handled = true
		}
	}

	// Save File
	if !handled {
		err = media.Store(media.URL(), option, mediaFile)
		return true, err
	}

	return true, nil
}

func SaveUploadAndCropImage(db *gorm.DB, obj interface{}) (err error) {
	db = db.Model(obj).Save(obj)
	err = db.Error
	if err != nil {
		return
	}

	updateColumns := map[string]interface{}{}

	for _, field := range db.Statement.Schema.Fields {
		ok, err := cropField(field, db)
		if err != nil {
			return err
		}
		if ok {
			updateColumns[field.DBName] = field.ReflectValueOf(db.Statement.Context, db.Statement.ReflectValue).Addr().Interface()
		}
	}

	if len(updateColumns) == 0 {
		return
	}

	err = db.UpdateColumns(updateColumns).Error

	return
}
