package publish

import (
	"context"

	"github.com/qor/oss"
	"github.com/qor/qor5/utils"
	"gorm.io/gorm"
)

type ListBuilder struct {
	db                 *gorm.DB
	storage            oss.StorageInterface
	context            context.Context
	needNextPageFunc   func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool
	getOldItemsFunc    func(modelSlice interface{})
	totalNumberPerPage int
	publishActionsFunc func(v []*OnePageItems) []PublishAction
}

func NewListBuilder(db *gorm.DB, storage oss.StorageInterface) *ListBuilder {
	return &ListBuilder{
		db:      db,
		storage: storage,
		context: context.Background(),

		needNextPageFunc: func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool {
			return currentPageNumber*totalNumberPerPage+int(0.5*float64(totalNumberPerPage)) <= totalNumberOfItems
		},
		getOldItemsFunc: func(modelSlice interface{}) {
			db.Where("status = ?", StatusOnline).Find(&modelSlice)
		},
		totalNumberPerPage: 30,
		publishActionsFunc: nil,
	}
}

type ListPublisher interface {
	TableName() string
	Sort(array []interface{})
	GenerateAndPutPages(result []*OnePageItems)
}

func (b *ListBuilder) PublishList(model interface{}, modelSlice []interface{}) {
	lp := model.(ListPublisher)

	var oldItems, addItems, deleteItems []interface{}
	b.getOldItemsFunc(&modelSlice)
	oldItems = modelSlice

	getAddItems(b.db, &modelSlice)
	addItems = modelSlice

	getDeleteItems(b.db, &modelSlice)
	deleteItems = modelSlice

	var newItems []interface{}
	if deleteItems != nil {
		var deleteMap = make(map[int][]int)
		for _, item := range deleteItems {
			lp := item.(ListInterface)
			deleteMap[lp.GetPageNumber()] = append(deleteMap[lp.GetPageNumber()], lp.GetPosition())
		}
		for _, item := range oldItems {
			lp := item.(ListInterface)
			if position, exist := deleteMap[lp.GetPageNumber()]; exist && utils.Contains(position, lp.GetPosition()) {
				continue
			}
			newItems = append(newItems, item)
		}
	} else {
		newItems = oldItems
	}

	if addItems != nil {
		newItems = append(newItems, addItems...)
	}

	lp.Sort(newItems)
	old := paginate(oldItems)
	newResult := rePaginate(newItems, b.totalNumberPerPage, b.needNextPageFunc)
	restartFrom := getRestartFromIndex(old, newResult)
	lp.GenerateAndPutPages(newResult[restartFrom:])
}

func (b *ListBuilder) NeedNextPageFunc(f func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool) *ListBuilder {
	b.needNextPageFunc = f
	return b
}

func (b *ListBuilder) GetOldItemsFunc(f func(modelSlice interface{})) *ListBuilder {
	b.getOldItemsFunc = f
	return b
}

func (b *ListBuilder) TotalNumberPerPage(number int) *ListBuilder {
	b.totalNumberPerPage = number
	return b
}

func (b *ListBuilder) PublishActionsFunc(f func(v []*OnePageItems) []PublishAction) *ListBuilder {
	b.publishActionsFunc = f
	return b
}

type OnePageItems struct {
	Items      []interface{}
	PageNumber int
}

func rePaginate(array []interface{}, totalNumberPerPage int, needNextPageFunc func(itemsCountInOnePage, currentPageNumber, allItemsCount int) bool) (result []*OnePageItems) {
	var pageNumber int
	for i := 1; needNextPageFunc(totalNumberPerPage, i, len(array)); i++ {
		pageNumber = i
		var items = array[(i-1)*totalNumberPerPage : i*totalNumberPerPage]
		for k := range items {
			model := items[k].(ListInterface)
			model.SetPageNumber(pageNumber)
			model.SetPosition(k)
		}
		result = append(result, &OnePageItems{
			Items:      items,
			PageNumber: pageNumber,
		})
	}

	var items = array[(pageNumber * totalNumberPerPage):]
	pageNumber = pageNumber + 1
	for k := range items {
		model := items[k].(ListInterface)
		model.SetPageNumber(pageNumber)
		model.SetPosition(k)
	}
	result = append(result, &OnePageItems{
		Items:      items,
		PageNumber: pageNumber,
	})
	return
}

func paginate(array []interface{}) (result []*OnePageItems) {
	lp := array[0].(ListPublisher)
	lp.Sort(array)
	var pageMap = make(map[int][]interface{})
	for _, item := range array {
		data := item.(ListInterface)
		pageMap[data.GetPageNumber()] = append(pageMap[data.GetPageNumber()], item)
	}
	for pageNumber, items := range pageMap {
		result = append(result, &OnePageItems{items, pageNumber})
	}
	return
}

func getRestartFromIndex(old, new []*OnePageItems) int {
	for indexNumber, array := range new {
		if len(array.Items) == len(old[indexNumber].Items) {
			for position, item := range array.Items {
				model := item.(ListInterface)
				oldModel := old[indexNumber].Items[position].(ListInterface)
				if model.GetPosition() != oldModel.GetPosition() {
					return indexNumber
				}
			}
		} else {
			return indexNumber
		}
	}
	return 0
}

func getAddItems(db *gorm.DB, modelSlice interface{}) {
	db.Where("status = ? AND list_updated = ?", StatusOnline, true).Find(&modelSlice)
	return
}

func getDeleteItems(db *gorm.DB, modelSlice interface{}) {
	db.Where("status = ? AND list_deleted = ?", StatusOnline, true).Find(&modelSlice)
	return
}
