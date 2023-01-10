package l10n

// Locale embed this struct into GROM-backend models to enable localization feature for your model
type Locale struct {
	LocaleCode string `sql:"size:20" gorm:"primary_key;default:''"`
}

// GetLocale get model's locale
func (l *Locale) GetLocale() string {
	return l.LocaleCode
}

// SetLocale set model's locale
func (l *Locale) SetLocale(locale string) {
	l.LocaleCode = locale
}
