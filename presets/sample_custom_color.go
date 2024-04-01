package presets

import "sync"

//color system

type colors struct {
	Primary   string `json:"primary,omitempty"`
	Secondary string `json:"secondary,omitempty"`
	Accent    string `json:"accent,omitempty"`
	Error     string `json:"error,omitempty"`
	Info      string `json:"info,omitempty"`
	Success   string `json:"success,omitempty"`
	Warning   string `json:"warning,omitempty"`
}

var instance *colors
var once sync.Once

// Color singleton of colors
func Color() *colors {
	once.Do(func() {
		instance = &colors{
			Primary:   "#3E63DD",
			Secondary: "#5B6471",
			Accent:    "#82B1FF",
			Error:     "#82B1FF",
			Info:      "#0091FF",
			Success:   "#30A46C",
			Warning:   "#F76808",
		}
	})
	return instance
}
