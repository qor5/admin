package tailwind

import (
	. "github.com/theplant/htmlgo"
)

var ButtonPresets = []string{"primary", "secondary", "success", "info", "warning", "error"}

var SpaceOptions = []string{"0", "10", "20", "30", "40", "50", "60", "70", "80", "90", "100"}

func TailwindContainerWrapper(classes string, comp ...HTMLComponent) HTMLComponent {
	return Div(comp...).
		Class("container-instance").ClassIf(classes, classes != "")
}
