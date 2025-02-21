package emailbuilder

import (
	"bytes"
	"html/template"
	"net/http"
	"regexp"

	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
)

var reEmail = regexp.MustCompile(`^[^@]+@[^@]+\.[^@.]+$`)

var EmailValidator = func(email string) bool {
	return reEmail.MatchString(email)
}

type Builder struct {
	db     *gorm.DB
	pb     *presets.Builder
	sender *DefaultSESDriver
	models []*ModelBuilder
	ab     *activity.Builder
}

func New(pb *presets.Builder, db *gorm.DB, tpl *presets.ModelBuilder) *Builder {
	b := &Builder{
		pb: pb,
		db: db,
	}
	b.Model(tpl, true)
	b.Sender(LoadSenderConfig())
	return b
}
func (b *Builder) DB(v *gorm.DB) *Builder {
	b.db = v
	return b
}
func (b *Builder) Activity(v *activity.Builder) *Builder {
	b.ab = v
	return b
}
func (b *Builder) Template(v *presets.ModelBuilder) *Builder {
	b.Model(v, true)
	return b

}

func (b *Builder) Install(pb *presets.Builder) (err error) {
	for _, model := range b.models {
		if err = model.Install(pb); err != nil {
			return
		}
	}
	return
}
func (b *Builder) ModelInstall(_ *presets.Builder, mb *presets.ModelBuilder) (err error) {
	b.Model(mb, false)
	return
}
func (b *Builder) Model(mb *presets.ModelBuilder, isTpl bool) (r *ModelBuilder) {
	r = &ModelBuilder{
		mb:    mb,
		b:     b,
		IsTpl: isTpl,
	}
	b.models = append(b.models, r)
	return
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, mb := range b.models {
		if mb.mb.Info().Verifier().Do(presets.PermGet).WithReq(r).IsAllowed() != nil {
			_, _ = w.Write([]byte(perm.PermissionDenied.Error()))
			return
		}
		mb.ServeHTTP(w, r)
		return
	}
}
func (b *Builder) AutoMigrate() *Builder {
	err := AutoMigrate(b.db)
	if err != nil {
		panic(err)
	}
	return b
}

func AutoMigrate(db *gorm.DB) (err error) {
	return db.AutoMigrate(&MailTemplate{}, &MailCampaign{})
}

func (b *Builder) Sender(config SESDriverConfig) *Builder {
	b.sender = NewDefaultSESDriver(config)
	return b
}

type SendRequest struct {
	TemplateID     int    `json:"template_id"`
	UserIds        []int  `json:"user_ids"`
	ToEmailAddress string `json:"to_email_address"`
}

type SendResult struct {
	UserId     int    `json:"user_id"`
	TemplateID int    `json:"template_id"`
	MessageID  string `json:"message_id"`
	ErrMsg     string `json:"err_msg"`
}

type UnifyResponse struct {
	Data interface{} `json:"data"`
}

type MailData struct {
	Name string
}

func GetTemplate(tmplStr string) (*template.Template, error) {
	tmpl, err := template.New("example").Parse(tmplStr)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func GetContent(tmpl *template.Template, mailData MailData) (string, error) {
	b := bytes.Buffer{}
	err := tmpl.Execute(&b, mailData)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
func (b *Builder) getTemplateModelBuilder() (mb *ModelBuilder) {
	for _, model := range b.models {
		if model.IsTpl {
			return model
		}
	}
	return
}
