package note

import (
	"github.com/qor5/admin/v3/presets"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// AfterCreateFunc is a type for functions to be called after creating a record
type AfterCreateFunc func(db *gorm.DB) error

// Builder constructs and configures database settings and models
type Builder struct {
	db              *gorm.DB
	afterCreateFunc AfterCreateFunc
}

// New initializes and returns a new Builder instance
func New(db *gorm.DB) *Builder {
	return &Builder{db: db}
}

// AfterCreate sets the after create function and returns the Builder
func (b *Builder) AfterCreate(f AfterCreateFunc) *Builder {
	b.afterCreateFunc = f
	return b
}

// Install automigrates models and registers internationalization support
func (b *Builder) Install(pb *presets.Builder) error {
	if err := b.db.AutoMigrate(&QorNote{}, &UserNote{}); err != nil {
		return err
	}
	registerI18n(pb)
	return nil
}

func registerI18n(pb *presets.Builder) {
	pb.I18n().
		RegisterForModule(language.English, I18nNoteKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nNoteKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nNoteKey, Messages_ja_JP)
}

// ModelInstall installs model-specific settings, events, and components
func (b *Builder) ModelInstall(pb *presets.Builder, m *presets.ModelBuilder) error {
	if m.Info().HasDetailing() {
		m.Detailing().AppendTabsPanelFunc(tabsPanel(b.db, m))
	}
	m.Editing().AppendTabsPanelFunc(tabsPanel(b.db, m))
	m.RegisterEventFunc(createNoteEvent, createNoteAction(b, m))
	m.RegisterEventFunc(updateUserNoteEvent, updateUserNoteAction(b, m))
	m.Listing().Field("Notes").ComponentFunc(noteFunc(b.db, m))
	return nil
}
