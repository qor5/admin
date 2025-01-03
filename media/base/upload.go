package base

import (
	"encoding/json"
	"errors"

	"github.com/qor5/web/v3"
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
		mediaFile, err = media.Retrieve(media.URL(OriginalSizeKey))
	}
	defer mediaFile.Close()

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

func SaveUploadAndCropImage(db *gorm.DB, obj interface{}, _ string, _ *web.EventContext) (err error) {
	err = db.Transaction(func(tx *gorm.DB) (dbErr error) {
		tx = tx.Model(obj).Save(obj)
		if dbErr = tx.Error; dbErr != nil {
			return
		}
		var (
			updateColumns = make(map[string]interface{})
			ok            bool
		)

		for _, field := range tx.Statement.Schema.Fields {
			if ok, dbErr = cropField(field, tx); dbErr != nil {
				return
			}
			if ok {
				updateColumns[field.DBName] = field.ReflectValueOf(tx.Statement.Context, tx.Statement.ReflectValue).Addr().Interface()
			}
		}

		if len(updateColumns) == 0 {
			return
		}

		return tx.UpdateColumns(updateColumns).Error
	})
	return
}
