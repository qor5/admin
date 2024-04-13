package models

import (
	"github.com/qor5/admin/v3/microsite"
)

type MicrositeModel struct {
	Name        string
	Description string
	microsite.MicroSite
}
