package docsrc

import (
	. "github.com/theplant/docgo"
)

var SearchAndFilterAndScope = Doc()

var Filter = Doc(
	Markdown(`
	
	## How to add basic filter 
	Call ~FilterDataFunc~ on listing builder, you need to configure the ~FilterItem~ to build the filter. see below code for detail.
	
	~~~go
		// v "github.com/goplaid/x/vuetifyx"

		// create a ModelBuilder
		videoBuilder := b.Model(&Video{})

		// get its ListingBuilder
		listing := videoBuilder.Listing()

		// Call FilterDataFunc to construct proper FilterData.
		listing.FilterDataFunc(func(ctx *web.EventContext) v.FilterData {
			var options []*v.SelectItem
			for _, status := range STATUSES {
				options = append(options, &v.SelectItem{Value: status, Text: status})
			}
	
			return []*v.FilterItem{
				{
					Key:          "status",
					Label:        "Status",
					ItemType:     v.ItemTypeSelect,
					SQLCondition: "status %s ?"", // %s is the db query condition like >, >=, =, <, <=, like. ？is the query value
					Options:      options,  // options list
				},
			}
		})
	~~~

`))
