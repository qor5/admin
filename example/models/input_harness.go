package models

import (
	"time"

	"github.com/lib/pq"
)

type InputHarness struct {
	ID               uint
	TextField1       string
	TextArea1        string
	Switch1          bool
	Slider1          int
	Select1          string
	RangeSlider1     pq.Int64Array `gorm:"type:integer[]"`
	Radio1           string
	FileInput1       string
	Combobox1        string
	Checkbox1        bool
	Autocomplete1    pq.StringArray `gorm:"type:text[]"`
	ButtonGroup1     string
	ChipGroup1       string
	ItemGroup1       string
	ListItemGroup1   string
	SlideGroup1      string
	ColorPicker1     string
	DatePicker1      string
	DatePickerMonth1 string
	TimePicker1      string
	UpdatedAt        time.Time
	CreatedAt        time.Time
}
