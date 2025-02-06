package redirection

import (
	"strings"

	"github.com/qor5/x/v3/i18n"
)

type Messages struct {
	RepeatedSourceErrorTemplate    string
	SourceInvalidFormatTemplate    string
	TargetInvalidFormatTemplate    string
	TargetUnreachableErrorTemplate string
	TargetObjectNotExistedTemplate string
	NormalErrorTemplate            string
	RedirectErrorTemplate          string
	FileUploadFailed               string
	ErrorTips                      string
}

const I18nRedirectionKey i18n.ModuleKey = "I18nRedirectionKey"

var Messages_en_US = &Messages{
	RepeatedSourceErrorTemplate:    "Row {Rows}: Source Is Duplicated.",
	SourceInvalidFormatTemplate:    "Source Invalid Format",
	TargetInvalidFormatTemplate:    "Target Invalid Format",
	TargetUnreachableErrorTemplate: "Row {Rows}: Target Is Unreachable.",
	TargetObjectNotExistedTemplate: "Row {Rows}: Target Object Not Existed",
	NormalErrorTemplate:            "Row {Rows}:{Message}",
	RedirectErrorTemplate:          "Row {Rows}: Redirection Failed.",
	FileUploadFailed:               "File Upload Failed",
	ErrorTips:                      "ErrorTips",
}

var Messages_zh_CN = &Messages{
	RepeatedSourceErrorTemplate:    "第{Rows}行：源数据重复。",
	SourceInvalidFormatTemplate:    "源数据格式无效。",
	TargetInvalidFormatTemplate:    "目标格式无效。",
	TargetUnreachableErrorTemplate: "第{Rows}行：目标无法访问。",
	TargetObjectNotExistedTemplate: "第{Rows}行：目标对象不存在。",
	NormalErrorTemplate:            "第{Rows}行：{Message}",
	RedirectErrorTemplate:          "第{Rows}行：重定向失败。",
	FileUploadFailed:               "文件上传失败。",
	ErrorTips:                      "错误提示",
}

var Messages_ja_JP = &Messages{
	RepeatedSourceErrorTemplate:    "{Rows}行目: Source が重複しています。",
	SourceInvalidFormatTemplate:    "Source のフォーマットが無効です。",
	TargetInvalidFormatTemplate:    "Target のフォーマットが無効です。",
	TargetUnreachableErrorTemplate: "{Rows}行目: Target に到達できません。",
	TargetObjectNotExistedTemplate: "ターゲットオブジェクトが存在しません。",
	NormalErrorTemplate:            "{Rows}行目: {Message}",
	RedirectErrorTemplate:          "{Rows}行目: リダイレクトに失敗しました。",
	FileUploadFailed:               "ファイルのアップロードに失敗しました。",
	ErrorTips:                      "エラーのヒント",
}

func (msgr *Messages) RepeatedSource(vs map[string][]string) string {
	var messages []string
	for _, rows := range vs {
		messages = append(messages, strings.NewReplacer(
			"{Rows}", strings.Join(rows, ","),
		).Replace(msgr.RepeatedSourceErrorTemplate))
	}
	return strings.Join(messages, "\n")
}

func (msgr *Messages) TargetUnreachableError(vs map[string][]string) string {
	var messages []string
	for _, rows := range vs {
		messages = append(messages, strings.NewReplacer(
			"{Rows}", strings.Join(rows, ","),
		).Replace(msgr.TargetUnreachableErrorTemplate))
	}
	return strings.Join(messages, "\n")
}

func (msgr *Messages) RedirectError(vs []string) string {
	return strings.NewReplacer(
		"{Rows}", strings.Join(vs, ","),
	).Replace(msgr.RedirectErrorTemplate)
}

func (msgr *Messages) SourceInvalidFormat(name string) string {
	return strings.NewReplacer(
		"{Name}", name,
	).Replace(msgr.SourceInvalidFormatTemplate)
}

func (msgr *Messages) TargetInvalidFormat(name string) string {
	return strings.NewReplacer(
		"{Name}", name,
	).Replace(msgr.TargetInvalidFormatTemplate)
}

func (msgr *Messages) InvalidFormat(vs map[string]string) string {
	var messages []string
	for rows, message := range vs {
		messages = append(messages, strings.NewReplacer(
			"{Rows}", rows, "{Message}", message,
		).Replace(msgr.NormalErrorTemplate))
	}
	return strings.Join(messages, "\n")
}

func (msgr *Messages) TargetObjectNotExisted(vs []string) string {
	return strings.NewReplacer(
		"{Rows}", strings.Join(vs, ","),
	).Replace(msgr.TargetObjectNotExistedTemplate)
}
