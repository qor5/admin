package presets

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetifyx"
)

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

func (b *ListingBuilder) ConfigureAutocompleteDataSource(config *AutocompleteDataSourceConfig, name ...string) *vuetifyx.AutocompleteDataSource {
	if config == nil {
		panic("config is required")
	}

	if config.OptionValue == "" {
		config.OptionValue = "ID"
	}

	if config.OptionText == nil {
		config.OptionText = "ID"
	}

	if config.KeywordColumns == nil {
		config.KeywordColumns = b.searchColumns
	}

	if config.OrderBy == "" {
		config.OrderBy = b.orderBy
	}

	if config.PerPage == 0 {
		config.PerPage = b.perPage
	}

	if config.PerPage == 0 {
		config.PerPage = 20
	}

	eventName := autocompleteDataSourceEvent
	if len(name) > 0 {
		eventName = eventName + "-" + strings.ToLower(name[0])
	}

	b.mb.RegisterEventFunc(eventName, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			objs       interface{}
			totalCount int
			page       int64
		)

		if v, err := strconv.ParseInt(ctx.R.FormValue("page"), 10, 64); err == nil {
			page = v
		}

		searchParams := &SearchParams{
			KeywordColumns: b.searchColumns,
			Keyword:        ctx.R.FormValue("keyword"),
			PerPage:        config.PerPage,
			OrderBy:        config.OrderBy,
			Page:           page,
		}

		if config.SQLConditions != nil {
			searchParams.SQLConditions = config.SQLConditions
		}

		objs, totalCount, err = b.Searcher(b.mb.NewModelSlice(), searchParams, ctx)
		if err != nil {
			return web.EventResponse{}, err
		}

		reflectValue := reflect.Indirect(reflect.ValueOf(objs))
		var items []OptionItem
		for i := 0; i < reflectValue.Len(); i++ {
			value := fmt.Sprintf("%v", reflect.Indirect(reflectValue.Index(i)).FieldByName(config.OptionValue).Interface())

			var text string
			switch config.OptionText.(type) {
			case func(interface{}) string:
				text = config.OptionText.(func(interface{}) string)(reflectValue.Index(i).Interface())
			case string:
				text = fmt.Sprintf("%v", reflect.Indirect(reflectValue.Index(i)).FieldByName(config.OptionText.(string)).Interface())
			}

			var icon string
			if config.OptionIcon != nil {
				icon = config.OptionIcon(reflectValue.Index(i).Interface())
			}

			items = append(items, OptionItem{
				Text:  text,
				Value: value,
				Icon:  icon,
			})
		}

		current := int((page-1)*config.PerPage) + len(items)
		if current > totalCount {
			current = totalCount
		}

		pages := totalCount / int(config.PerPage)
		if totalCount%int(config.PerPage) > 0 {
			pages++
		}

		r.Data = AutocompleteDataResult{
			Total:   totalCount,
			Current: current,
			Pages:   pages,
			Items:   items,
		}
		return
	})

	return &vuetifyx.AutocompleteDataSource{
		RemoteURL: b.mb.Info().ListingHref(),
		EventName: eventName,
		IsPaging:  config.IsPaging,
		HasIcon:   config.OptionIcon != nil,
	}
}
