package examples_admin

import (
	"cmp"
	"encoding/json"
	"math"
	"net/http"
	"strings"

	"github.com/qor5/web/v3"
)

type (
	LinkageSelectFilterItemRemoteServer struct{}
)

var cities = []Item{
	{ID: "1", Name: "浙江", Level: 1, ParentID: ""},
	{ID: "2", Name: "江苏", Level: 1, ParentID: ""},

	{ID: "3", Name: "杭州", Level: 2, ParentID: "1"},
	{ID: "4", Name: "宁波", Level: 2, ParentID: "1"},
	{ID: "5", Name: "南京", Level: 2, ParentID: "2"},
	{ID: "6", Name: "苏州", Level: 2, ParentID: "2"},

	{ID: "7", Name: "拱墅区", Level: 3, ParentID: "3"},
	{ID: "8", Name: "西湖区", Level: 3, ParentID: "3"},
	{ID: "9", Name: "镇海区", Level: 3, ParentID: "4"},
	{ID: "10", Name: "鄞州区", Level: 3, ParentID: "4"},
	{ID: "11", Name: "鼓楼区", Level: 3, ParentID: "5"},
	{ID: "12", Name: "玄武区", Level: 3, ParentID: "5"},
	{ID: "13", Name: "常熟区", Level: 3, ParentID: "6"},
	{ID: "14", Name: "吴江区", Level: 3, ParentID: "6"},
}

type (
	Item struct {
		ID       string `json:"ID"`
		Name     string `json:"Name"`
		Level    int    `json:"level"`
		ParentID string `json:"-"`
		Parent   *Item  `json:"parent"`
	}
	PaginatedResponse struct {
		Data    []Item `json:"data"`
		Total   int    `json:"total"`
		Pages   int    `json:"pages"`
		Current int    `json:"current"`
	}
	QueryParams struct {
		Page     int    `form:"page"`
		PageSize int    `form:"pageSize"`
		Level    int    `form:"level"`
		Search   string `form:"search"`
		ParentID string `form:"parentID"`
	}
)

func loadAllCity(item *Item, cities []Item) {
	if item.ParentID == "" {
		return
	}
	for _, city := range cities {
		if city.ID == item.ParentID {
			item.Parent = &city
			loadAllCity(&city, cities)
			return
		}
	}
}

func checkIsParent(city *Item, parentID string) bool {
	if city.ID == parentID {
		return true
	}
	if city.Parent != nil {
		return checkIsParent(city.Parent, parentID)
	}
	return false
}

func getItem(name string) *Item {
	for _, item := range cities {
		if item.Name == name {
			return &item
		}
	}
	return nil
}

func citiesResponse(page, pageSize, level int, search, parentID string) *PaginatedResponse {
	var findCities []Item
	for _, city := range cities {
		loadAllCity(&city, cities)
		if search != "" && !strings.Contains(city.Name, search) {
			continue
		}
		if parentID != "" && !checkIsParent(&city, parentID) {
			continue
		}
		if city.Level != level {
			continue
		}
		findCities = append(findCities, city)
	}
	total := len(findCities)
	start := (page - 1) * pageSize
	end := page * pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	findCities = findCities[start:end]

	return &PaginatedResponse{
		Data:    findCities,
		Total:   total,
		Pages:   int(math.Ceil(float64(total) / float64(pageSize))),
		Current: end,
	}
}

func (b *LinkageSelectFilterItemRemoteServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	ctx := web.EventContext{R: r}
	level := cmp.Or(ctx.ParamAsInt("level"), 1)
	page := cmp.Or(ctx.ParamAsInt("page"), 1)
	pageSize := cmp.Or(ctx.ParamAsInt("pageSize"), 2)
	search := ctx.Param("search")
	parentID := ctx.Param("parentID")
	w.Header().Add("Content-Type", "application/json")
	body, _ := json.Marshal(citiesResponse(page, pageSize, level, search, parentID))
	_, _ = w.Write(body)
}
