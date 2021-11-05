package publish

import (
	"context"
	"errors"

	"github.com/qor/oss"
	"github.com/qor/qor5/utils"
	"github.com/theplant/sliceutils"
	"gorm.io/gorm"
)

type ListBuilder struct {
	db                 *gorm.DB
	storage            oss.StorageInterface
	context            context.Context
	needNextPageFunc   func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool
	getOldItemsFunc    func(record interface{}) (result []interface{})
	totalNumberPerPage int
	comparisonFunc     func(recordA interface{}, recordB interface{}) bool
	publishActionsFunc func(lp ListPublisher, result []*OnePageItems) (objs []*PublishAction)
}

func NewListBuilder(db *gorm.DB, storage oss.StorageInterface) *ListBuilder {
	return &ListBuilder{
		db:      db,
		storage: storage,
		context: context.Background(),

		needNextPageFunc: func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool {
			return currentPageNumber*totalNumberPerPage+int(0.5*float64(totalNumberPerPage)) <= totalNumberOfItems
		},

		//&[]models.ListModel{}
		getOldItemsFunc: func(record interface{}) (result []interface{}) {
			err := db.Where("status = ? AND page_number <> ? AND list_updated = ? AND list_deleted = ?", StatusOnline, 0, false, false).Find(&record).Error
			if err != nil {
				panic(err)
			}
			result = sliceutils.Wrap(record)
			return
		},
		totalNumberPerPage: 30,
		comparisonFunc: func(recordA interface{}, recordB interface{}) bool {
			return true
		},
		publishActionsFunc: func(lp ListPublisher, result []*OnePageItems) (objs []*PublishAction) {
			for _, onePageItems := range result {
				objs = append(objs, &PublishAction{
					Url:      lp.GetListUrl(onePageItems.PageNumber),
					Content:  lp.GetListContent(onePageItems),
					IsDelete: false,
				})
			}
			return
		},
	}
}

func getAddItems(db *gorm.DB, record interface{}) (result []interface{}) {
	db.Where("list_updated = ?", true).Find(&record)
	result = sliceutils.Wrap(record)
	return
}

func getDeleteItems(db *gorm.DB, record interface{}) (result []interface{}) {
	db.Where("list_deleted = ?", true).Find(&record)
	result = sliceutils.Wrap(record)
	return
}

type ListPublisher interface {
	TableName() string
	GetListUrl(pageNumber int) string
	GetListContent(onePageItems *OnePageItems) string
	Sort(array []interface{})
}

func (b *ListBuilder) PublishList(model interface{}, modelSlice interface{}) (err error) {
	lp := model.(ListPublisher)

	addItems := getAddItems(b.db, modelSlice)
	deleteItems := getDeleteItems(b.db, modelSlice)

	if len(deleteItems) == 0 && len(addItems) == 0 {
		return nil
	}

	oldItems := b.getOldItemsFunc(modelSlice)

	var newItems []interface{}
	if len(deleteItems) != 0 {
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

	if len(addItems) != 0 {
		newItems = append(newItems, addItems...)
	}

	lp.Sort(newItems)
	var oldResult []*OnePageItems
	if len(oldItems) > 0 {
		oldResult = paginate(oldItems)
	}
	newResult := rePaginate(newItems, b.totalNumberPerPage, b.needNextPageFunc)
	restartFrom := getRestartFromIndex(oldResult, newResult)

	var objs []*PublishAction
	objs = b.publishActionsFunc(lp, newResult[restartFrom:])

	err = utils.Transact(b.db, func(tx *gorm.DB) (err1 error) {
		if err = UploadOrDelete(objs, b.storage); err != nil {
			return
		}

		for _, items := range newResult[restartFrom:] {
			for _, item := range items.Items {
				if listItem, ok := item.(ListInterface); ok {
					if err1 = b.db.Model(item).Updates(map[string]interface{}{
						"list_updated": listItem.GetListUpdated(),
						"list_deleted": listItem.GetListDeleted(),
						"page_number":  listItem.GetPageNumber(),
						"position":     listItem.GetPosition(),
					}).Error; err1 != nil {
						return
					}
				} else {
					return errors.New("model must be ListInterface")
				}
			}

		}

		for _, item := range deleteItems {
			if _, ok := item.(ListInterface); ok {
				if err1 = b.db.Model(item).Updates(map[string]interface{}{
					"list_updated": false,
					"list_deleted": false,
					"page_number":  0,
					"position":     0,
				}).Error; err1 != nil {
					return
				}
			} else {
				return errors.New("model must be ListInterface")
			}

		}
		return
	})
	return nil
}

func (b *ListBuilder) NeedNextPageFunc(f func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool) *ListBuilder {
	b.needNextPageFunc = f
	return b
}

func (b *ListBuilder) GetOldItemsFunc(f func(record interface{}) (result []interface{})) *ListBuilder {
	b.getOldItemsFunc = f
	return b
}

func (b *ListBuilder) TotalNumberPerPage(number int) *ListBuilder {
	b.totalNumberPerPage = number
	return b
}

func (b *ListBuilder) ComparisonFunc(f func(recordA interface{}, recordB interface{}) bool) *ListBuilder {
	b.comparisonFunc = f
	return b
}

func (b *ListBuilder) PublishActionsFunc(f func(lp ListPublisher, result []*OnePageItems) (objs []*PublishAction)) *ListBuilder {
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
			model.SetListUpdated(false)
			model.SetListDeleted(false)
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
		model.SetListUpdated(false)
		model.SetListDeleted(false)
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
	if len(old) == 0 {
		return 0
	}
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
