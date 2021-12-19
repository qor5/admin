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

type ListPublishBuilder struct {
	db                 *gorm.DB
	storage            oss.StorageInterface
	context            context.Context
	needNextPageFunc   func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool
	getOldItemsFunc    func(record interface{}) (result []interface{}, err error)
	totalNumberPerPage int
	publishActionsFunc func(db *gorm.DB, lp ListPublisher, result []*OnePageItems, indexPage *OnePageItems) (objs []*PublishAction)
}

func NewListPublishBuilder(db *gorm.DB, storage oss.StorageInterface) *ListPublishBuilder {
	return &ListPublishBuilder{
		db:      db,
		storage: storage,
		context: context.Background(),

		needNextPageFunc: func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool {
			return currentPageNumber*totalNumberPerPage+int(0.5*float64(totalNumberPerPage)) <= totalNumberOfItems
		},
		getOldItemsFunc: func(record interface{}) (result []interface{}, err error) {
			err = db.Where("page_number <> ? ", 0).Find(&record).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return
			}
			return sliceutils.Wrap(record), nil
		},
		totalNumberPerPage: 30,
		publishActionsFunc: func(db *gorm.DB, lp ListPublisher, result []*OnePageItems, indexPage *OnePageItems) (objs []*PublishAction) {
			for _, onePageItems := range result {
				objs = append(objs, &PublishAction{
					Url:      lp.GetListUrl(strconv.Itoa(onePageItems.PageNumber)),
					Content:  lp.GetListContent(db, onePageItems),
					IsDelete: false,
				})
			}
			if indexPage != nil {
				objs = append(objs, &PublishAction{
					Url:      lp.GetListUrl("index"),
					Content:  lp.GetListContent(db, indexPage),
					IsDelete: false,
				})
			}
			return
		},
	}
}

func (b *ListPublishBuilder) WithValue(key, val interface{}) *ListPublishBuilder {
	b.context = context.WithValue(b.context, key, val)
	return b
}

func getAddItems(db *gorm.DB, record interface{}) (result []interface{}, err error) {
	err = db.Where("page_number = ? AND list_updated = ?", 0, true).Find(&record).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	return sliceutils.Wrap(record), nil
}

func getDeleteItems(db *gorm.DB, record interface{}) (result []interface{}, err error) {
	err = db.Where("page_number <> ? AND list_deleted = ?", 0, true).Find(&record).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	return sliceutils.Wrap(record), nil
}

func getRepublishItems(db *gorm.DB, record interface{}) (result []interface{}, err error) {
	err = db.Where("page_number <> ? AND list_updated = ?", 0, true).Find(&record).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	return sliceutils.Wrap(record), nil
}

type ListPublisher interface {
	GetListUrl(pageNumber string) string
	GetListContent(db *gorm.DB, onePageItems *OnePageItems) string
	Sort(array []interface{})
}

func (b *ListPublishBuilder) Run(model interface{}) (err error) {
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

	republishItems, err := getRepublishItems(b.db, modelSlice)
	if err != nil {
		return
	}

	lp := model.(ListPublisher)
	lp.Sort(newItems)
	var oldResult []*OnePageItems
	if len(oldItems) > 0 {
		oldResult = paginate(oldItems)
	}

	var republishResult []*OnePageItems
	if len(republishItems) > 0 {
		republishResult = paginate(republishItems)
	}

	newResult := rePaginate(newItems, b.totalNumberPerPage, b.needNextPageFunc)

	needPublishResults, indexResult := getNeedPublishResultsAndIndexResult(oldResult, newResult, republishResult)

	var objs []*PublishAction
	objs = b.publishActionsFunc(b.db, lp, needPublishResults, indexResult)

	err = utils.Transact(b.db, func(tx *gorm.DB) (err1 error) {
		if err = UploadOrDelete(objs, b.storage); err != nil {
			return
		}

		for _, items := range needPublishResults {
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

func (b *ListPublishBuilder) NeedNextPageFunc(f func(totalNumberPerPage, currentPageNumber, totalNumberOfItems int) bool) *ListPublishBuilder {
	b.needNextPageFunc = f
	return b
}

func (b *ListPublishBuilder) GetOldItemsFunc(f func(record interface{}) (result []interface{}, err error)) *ListPublishBuilder {
	b.getOldItemsFunc = f
	return b
}

func (b *ListPublishBuilder) TotalNumberPerPage(number int) *ListPublishBuilder {
	b.totalNumberPerPage = number
	return b
}

func (b *ListPublishBuilder) PublishActionsFunc(f func(db *gorm.DB, lp ListPublisher, result []*OnePageItems, indexPage *OnePageItems) (objs []*PublishAction)) *ListPublishBuilder {
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
//Pick out the pages which are needed to republish
func getNeedPublishResultsAndIndexResult(oldResults, newResults, republishResults []*OnePageItems) (needPublishResults []*OnePageItems, indexResult *OnePageItems) {
	if len(oldResults) == 0 {
		return newResults, newResults[len(newResults)-1]
	}

	var republishMap = make(map[int]bool)
	for _, republishResult := range republishResults {
		republishMap[republishResult.PageNumber] = true
	}

	for i, newResult := range newResults {

		// Add need publish pages to needPublishResults
		if _, exist := republishMap[newResult.PageNumber]; exist {
			needPublishResults = append(needPublishResults, newResult)
			if i == len(newResults)-1 {
				indexResult = newResult
			}
			continue
		}

		// Add new page whose page number over old page's max page number
		if i > len(oldResults)-1 {
			needPublishResults = append(needPublishResults, newResult)
			if i == len(newResults)-1 {
				indexResult = newResult
			}
			continue
		}

		// Compare new page and old page
		// If the items are different, add to needPublishResults
		if len(newResult.Items) == len(oldResults[i].Items) {
			for position, item := range newResult.Items {
				model := item.(ListInterface)
				oldModel := oldResults[i].Items[position].(ListInterface)
				if model.GetPosition() != oldModel.GetPosition() {
					needPublishResults = append(needPublishResults, newResult)
					if i == len(newResults)-1 {
						indexResult = newResult
					}
					continue
				}
			}
		} else {
			needPublishResults = append(needPublishResults, newResult)
			if i == len(newResults)-1 {
				indexResult = newResult
			}
			continue
		}
	}
	return
}
