package models

import (
	"github.com/qor5/admin/microsite"
)

type MicrositeModel struct {
	Name        string
	Description string
	microsite.MicroSite
}
