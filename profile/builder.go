package profile

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/activity"
	plogin "github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	I18nProfileKey i18n.ModuleKey = "I18nProfileKey"
)

func (b *Builder) Install(pb *presets.Builder) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.pb != nil {
		return errors.Errorf("profile: already installed")
	}
	b.pb = pb
	pb.GetI18n().
		RegisterForModule(language.English, I18nProfileKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nProfileKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nProfileKey, Messages_ja_JP)
	pb.ProfileFunc(func(evCtx *web.EventContext) h.HTMLComponent {
		return b.NewCompo(evCtx, "")
	})

	dc := pb.GetDependencyCenter()
	injectorName := b.injectorName()
	dc.RegisterInjector(injectorName)
	dc.MustProvide(injectorName, func() *Builder {
		return b
	})
	return nil
}

type UserField struct {
	Name  string
	Value string
}

type User struct {
	ID          string
	Name        string
	Avatar      string
	Roles       []string
	Status      string
	Fields      []*UserField
	NotifCounts []*activity.NoteCount
}

func (u *User) GetFirstRole() string {
	role := "-"
	if len(u.Roles) > 0 {
		role = u.Roles[0]
	}
	return role
}

type Builder struct {
	mu sync.RWMutex
	pb *presets.Builder

	lsb             *plogin.SessionBuilder
	currentUserFunc func(ctx context.Context) (*User, error)
	renameCallback  func(ctx context.Context, newName string) error
}

func New(
	lsb *plogin.SessionBuilder,
	currentUserFunc func(ctx context.Context) (*User, error),
	renameCallback func(ctx context.Context, newName string) error,
) *Builder {
	return &Builder{
		lsb:             lsb,
		currentUserFunc: currentUserFunc,
		renameCallback:  renameCallback,
	}
}

func (b *Builder) injectorName() string {
	return "__profile__"
}

func (b *Builder) NewCompo(evCtx *web.EventContext, idSuffix string) h.HTMLComponent {
	b.mu.RLock()
	pb := b.pb
	b.mu.RUnlock()
	if pb == nil {
		panic("profile: not installed")
	}
	return h.Div().Class("d-flex flex-column align-stretch w-100").Children(
		b.pb.GetDependencyCenter().MustInject(b.injectorName(), &ProfileCompo{
			ID: b.pb.GetURIPrefix() + idSuffix,
		}),
	)
}
