package presets

const autocompleteDataSourceEvent = "autocomplete-data-source-event"

type AutocompleteDataResult struct {
	Items   []OptionItem `json:"items"`
	Total   int          `json:"total"`
	Current int          `json:"current"`
	Pages   int          `json:"pages"`
}

type OptionItem struct {
	Text  string `json:"text,omitempty"`
	Value string `json:"value,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

type AutocompleteDataSourceConfig struct {
	OptionValue string
	OptionText  interface{} // func(interface{}) string or string
	OptionIcon  func(interface{}) string

	IsPaging bool

	KeywordColumns []string
	SQLConditions  []*SQLCondition
	OrderBy        string
	PerPage        int64
}
