package publish

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strconv"

	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/x/v3/oss"
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

// model is a empty struct
// example: Product{}
func (b *ListPublishBuilder) Run(ctx context.Context, model interface{}) (err error) {
	// If model is Product{}
	// Generate a records: []*Product{}
	records := reflect.MakeSlice(reflect.SliceOf(reflect.New(reflect.TypeOf(model)).Type()), 0, 0).Interface()

	addItems, err := getAddItems(b.db, records)
	if err != nil {
		return
	}
	deleteItems, err := getDeleteItems(b.db, records)
	if err != nil {
		return
	}
	republishItems, err := getRepublishItems(b.db, records)
	if err != nil {
		return
	}

	if len(deleteItems) == 0 && len(addItems) == 0 && len(republishItems) == 0 {
		return nil
	}

	oldItems, err := b.getOldItemsFunc(records)
	if err != nil {
		return
	}

	newItems := oldItems
	if len(deleteItems) != 0 {
		newItems = []interface{}{}
		deleteMap := make(map[int][]int)
		for _, item := range deleteItems {
			lp := item.(ListInterface)
			deleteMap[lp.EmbedList().PageNumber] = append(deleteMap[lp.EmbedList().PageNumber], lp.EmbedList().Position)
		}
		for _, item := range oldItems {
			lp := item.(ListInterface)
			if position, exist := deleteMap[lp.EmbedList().PageNumber]; exist && slices.Contains(position, lp.EmbedList().Position) {
				continue
			}
			newItems = append(newItems, item)
		}
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

	var republishResult []*OnePageItems
	if len(republishItems) > 0 {
		republishResult = paginate(republishItems)
	}

	newResult := rePaginate(newItems, b.totalNumberPerPage, b.needNextPageFunc)

	needPublishResults, indexResult := getNeedPublishResultsAndIndexResult(oldResult, newResult, republishResult)

	objs := b.publishActionsFunc(b.db, lp, needPublishResults, indexResult)

	err = utils.Transact(b.db, func(tx *gorm.DB) (err1 error) {
		if err1 = UploadOrDelete(ctx, objs, b.storage); err1 != nil {
			return
		}

		for _, items := range needPublishResults {
			for _, item := range items.Items {
				if listItem, ok := item.(ListInterface); ok {
					if err1 = b.db.Model(item).Updates(map[string]interface{}{
						"list_updated": listItem.EmbedList().ListUpdated,
						"list_deleted": listItem.EmbedList().ListDeleted,
						"page_number":  listItem.EmbedList().PageNumber,
						"position":     listItem.EmbedList().Position,
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
	return
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

// Repaginate completely
// Regardless of the PageNumber and Position of the old data
// Resort and repaginate all data
func rePaginate(array []interface{}, totalNumberPerPage int, needNextPageFunc func(itemsCountInOnePage, currentPageNumber, allItemsCount int) bool) (result []*OnePageItems) {
	var pageNumber int
	for i := 1; needNextPageFunc(totalNumberPerPage, i, len(array)); i++ {
		pageNumber = i
		items := array[(i-1)*totalNumberPerPage : i*totalNumberPerPage]
		for k := range items {
			model := items[k].(ListInterface).EmbedList()
			model.PageNumber = pageNumber
			model.Position = k
			model.ListUpdated = false
			model.ListDeleted = false
		}
		result = append(result, &OnePageItems{
			Items:      items,
			PageNumber: pageNumber,
		})
	}

	items := array[(pageNumber * totalNumberPerPage):]
	pageNumber = pageNumber + 1
	for k := range items {
		model := items[k].(ListInterface).EmbedList()
		model.PageNumber = pageNumber
		model.Position = k
		model.ListUpdated = false
		model.ListDeleted = false
	}
	result = append(result, &OnePageItems{
		Items:      items,
		PageNumber: pageNumber,
	})
	return
}

// For old data
// Old data has PageNumber and Position
// Sort pages according to the PageNumber and position
func paginate(array []interface{}) (result []*OnePageItems) {
	lp := array[0].(ListPublisher)
	lp.Sort(array)
	pageMap := make(map[int][]interface{})
	for _, item := range array {
		data := item.(ListInterface)
		pageMap[data.EmbedList().PageNumber] = append(pageMap[data.EmbedList().PageNumber], item)
	}
	for pageNumber, items := range pageMap {
		result = append(result, &OnePageItems{items, pageNumber})
	}
	return
}

// Compare new pages and old pages
// Pick out the pages which are needed to republish
func getNeedPublishResultsAndIndexResult(oldResults, newResults, republishResults []*OnePageItems) (needPublishResults []*OnePageItems, indexResult *OnePageItems) {
	if len(oldResults) == 0 {
		return newResults, newResults[len(newResults)-1]
	}

	republishMap := make(map[int]bool)
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
				if model.EmbedList().Position != oldModel.EmbedList().Position {
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
