package models

import (
	"github.com/qor/qor5/microsite"
)

type MicrositeModel struct {
	Name        string
	Description string
	microsite.MicroSite
}
