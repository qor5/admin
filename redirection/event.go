package redirection

import (
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"

	"github.com/qor5/admin/v3/presets"
)

const (
	UploadFileEvent = "redirection_UploadFileEvent"
)

func (b *Builder) uploadFile(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		uf      uploadFiles
		file    multipart.File
		body    []byte
		records []Redirection
		msgr    = i18n.MustGetModuleMessages(ctx.R, I18nRedirectionKey, Messages_en_US).(*Messages)
	)
	ctx.MustUnmarshalForm(&uf)
	if len(uf.NewFiles) == 0 {
		web.AppendRunScripts(&r, web.Emit(redirection_notify_error_msg, msgr.FileUploadFailed))
		return
	}
	if file, err = uf.NewFiles[0].Open(); err != nil {
		return
	}
	if body, err = io.ReadAll(file); err != nil {
		return
	}
	if err = gocsv.UnmarshalBytes(body, &records); err != nil {
		return
	}

	if passed := b.checkRecords(&r, msgr, records); !passed {
		return
	}
	if passed := b.checkObjects(ctx, &r, msgr, records); !passed {
		return
	}
	r.Emit(
		b.mb.NotifModelsCreated(),
		presets.PayloadModelsCreated{
			Models: []any{records},
		},
	)
	presets.ShowMessage(&r, "success", v.ColorSuccess)
	return
}

func (*Builder) checkRecords(r *web.EventResponse, msgr *Messages, records []Redirection) (passed bool) {
	var (
		existedSource = make(map[string][]string)
		invalidFormat = make(map[string]string)
		urls          = make(map[string][]string)
	)
	for index, item := range records {
		row := strconv.Itoa(index + 1)
		var messages []string
		existedSource[item.Source] = append(existedSource[item.Source], strconv.Itoa(index+1))
		if strings.HasPrefix(item.Source, "http") || !strings.HasPrefix(item.Source, "/") {
			messages = append(messages, msgr.SourceInvalidFormat(item.Source))
		}
		if strings.HasPrefix(item.Target, "http") {
			urls[item.Target] = append(urls[item.Target], row)
		} else if !strings.HasPrefix(item.Target, "/") {
			messages = append(messages, msgr.TargetInvalidFormat(item.Target))
		}
		if len(messages) > 0 {
			invalidFormat[row] = strings.Join(messages, ",")
		}
	}
	if len(invalidFormat) > 0 {
		web.AppendRunScripts(r, web.Emit(redirection_notify_error_msg, msgr.InvalidFormat(invalidFormat)))
		return
	}
	duplicateSources := make(map[string][]string)
	for source, rows := range existedSource {
		if len(rows) > 1 {
			duplicateSources[source] = rows
		}
	}
	if len(duplicateSources) > 0 {
		web.AppendRunScripts(r, web.Emit(redirection_notify_error_msg, msgr.RepeatedSource(duplicateSources)))
		return
	}

	// check all target urls is reachable
	if len(urls) > 0 {
		failedUrls := checkURLsBatch(urls)
		if len(failedUrls) > 0 {
			errorFieldUrls := map[string][]string{}
			for _, failedUrl := range failedUrls {
				errorFieldUrls[failedUrl] = urls[failedUrl]
			}
			web.AppendRunScripts(r, web.Emit(redirection_notify_error_msg, msgr.TargetUnreachableError(errorFieldUrls)))
			return
		}
	}

	return true
}

func (b *Builder) checkObjects(ctx *web.EventContext, r *web.EventResponse, msgr *Messages, records []Redirection) (passed bool) {
	var (
		errorRows         []string
		errorRedirectRows []string
	)
	// check  target object is exist
	for index, record := range records {
		row := strconv.Itoa(index + 1)
		if !strings.HasPrefix(record.Target, "http") && !b.checkObjectExists(ctx.R.Context(), record.Target) {
			errorRows = append(errorRows, row)
		}
	}
	if len(errorRows) > 0 {
		web.AppendRunScripts(r, web.Emit(redirection_notify_error_msg, msgr.TargetObjectNotExisted(errorRows)))
		return
	}
	for index, item := range records {
		row := strconv.Itoa(index + 1)
		if err := b.saver(ctx.R.Context(), &item); err != nil {
			errorRedirectRows = append(errorRedirectRows, row)
		}
	}
	if len(errorRedirectRows) > 0 {
		web.AppendRunScripts(r, web.Emit(redirection_notify_error_msg, msgr.RedirectError(errorRedirectRows)))
		return
	}
	return true
}
