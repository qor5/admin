package publish

import (
	"context"
	"errors"
	"reflect"
	"strconv"

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
	getOldItemsFunc    func(record interface{}) (result []interface{}, err error)
	totalNumberPerPage int
	publishActionsFunc func(lp ListPublisher, result []*OnePageItems, indexPage *OnePageItems) (objs []*PublishAction)
}

func NewListBuilder(db *gorm.DB, storage oss.StorageInterface) *ListBuilder {
	return &ListBuilder{
		db:      db,
		storage: storage,
		context: context.Background(),

		needNextPageFunc: func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool {
			return currentPageNumber*totalNumberPerPage+int(0.5*float64(totalNumberPerPage)) <= totalNumberOfItems
		},
		getOldItemsFunc: func(record interface{}) (result []interface{}, err error) {
			err = db.Where("status = ? AND page_number <> ? AND list_updated = ? AND list_deleted = ?", StatusOnline, 0, false, false).Find(&record).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return
			}
			return sliceutils.Wrap(record), nil
		},
		totalNumberPerPage: 30,
		publishActionsFunc: func(lp ListPublisher, result []*OnePageItems, indexPage *OnePageItems) (objs []*PublishAction) {
			for _, onePageItems := range result {
				objs = append(objs, &PublishAction{
					Url:      lp.GetListUrl(strconv.Itoa(onePageItems.PageNumber)),
					Content:  lp.GetListContent(onePageItems),
					IsDelete: false,
				})
			}
			objs = append(objs, &PublishAction{
				Url:      lp.GetListUrl("index"),
				Content:  lp.GetListContent(indexPage),
				IsDelete: false,
			})
			return
		},
	}
}

func getAddItems(db *gorm.DB, record interface{}) (result []interface{}, err error) {
	err = db.Where("list_updated = ?", true).Find(&record).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	return sliceutils.Wrap(record), nil
}

func getDeleteItems(db *gorm.DB, record interface{}) (result []interface{}, err error) {
	err = db.Where("list_deleted = ?", true).Find(&record).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	return sliceutils.Wrap(record), nil
}

type ListPublisher interface {
	GetListUrl(pageNumber string) string
	GetListContent(onePageItems *OnePageItems) string
	Sort(array []interface{})
}

func (b *ListBuilder) PublishList(model interface{}) (err error) {
	//If model is Product{}
	//Generate a modelSlice: []*Product{}
	modelSlice := reflect.MakeSlice(reflect.SliceOf(reflect.New(reflect.TypeOf(model)).Type()), 0, 0).Interface()

	addItems, err := getAddItems(b.db, modelSlice)
	if err != nil {
		return
	}
	deleteItems, err := getDeleteItems(b.db, modelSlice)
	if err != nil {
		return
	}

	if len(deleteItems) == 0 && len(addItems) == 0 {
		return nil
	}

	oldItems, err := b.getOldItemsFunc(modelSlice)
	if err != nil {
		return
	}

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

	lp := model.(ListPublisher)
	lp.Sort(newItems)
	var oldResult []*OnePageItems
	if len(oldItems) > 0 {
		oldResult = paginate(oldItems)
	}
	newResult := rePaginate(newItems, b.totalNumberPerPage, b.needNextPageFunc)
	restartFrom := getRestartFromIndex(oldResult, newResult)

	//newResult[restartFrom:] is the pages that will be published to storage
	var objs []*PublishAction
	objs = b.publishActionsFunc(lp, newResult[restartFrom:], newResult[len(newResult)-1])

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

func (b *ListBuilder) GetOldItemsFunc(f func(record interface{}) (result []interface{}, err error)) *ListBuilder {
	b.getOldItemsFunc = f
	return b
}

func (b *ListBuilder) TotalNumberPerPage(number int) *ListBuilder {
	b.totalNumberPerPage = number
	return b
}

func (b *ListBuilder) PublishActionsFunc(f func(lp ListPublisher, result []*OnePageItems, indexPage *OnePageItems) (objs []*PublishAction)) *ListBuilder {
	b.publishActionsFunc = f
	return b
}

type OnePageItems struct {
	Items      []interface{}
	PageNumber int
}

//Repaginate completely
//Regardless of the PageNumber and Position of the old data
//Resort and repaginate all data
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

//For old data
//Old data has PageNumber and Position
//Sort pages according to the PageNumber and position
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

//Compare new pages and old pages
//Pick out which is the first page that different on both sides
//Return the page's index
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
