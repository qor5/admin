package note

import (
	"github.com/qor5/admin/v3/presets"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type AfterCreateFunc func(db *gorm.DB) error

type Builder struct {
	db              *gorm.DB
	afterCreateFunc AfterCreateFunc
}

func New(db *gorm.DB) *Builder {
	b := &Builder{
		db: db,
	}
	return b
}

func (b *Builder) AfterCreate(f AfterCreateFunc) (r *Builder) {
	b.afterCreateFunc = f
	return b
}

func (b *Builder) Install(pb *presets.Builder) error {
	db := b.db
	if err := db.AutoMigrate(QorNote{}, UserNote{}); err != nil {
		return err
	}

	pb.GetI18n().
		RegisterForModule(language.English, I18nNoteKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nNoteKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nNoteKey, Messages_ja_JP)
	return nil
}

func (b *Builder) ModelInstall(pb *presets.Builder, m *presets.ModelBuilder) error {
	db := b.db
	if m.Info().HasDetailing() {
		m.Detailing().AppendTabsPanelFunc(tabsPanel(db, m))
	}
	m.Editing().AppendTabsPanelFunc(tabsPanel(db, m))
	m.RegisterEventFunc(createNoteEvent, createNoteAction(b, m))
	m.RegisterEventFunc(updateUserNoteEvent, updateUserNoteAction(b, m))
	m.Listing().Field("Notes").ComponentFunc(noteFunc(db, m))
	return nil
}
