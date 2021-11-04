package models

import (
	"context"
	"fmt"
	"sort"

	"github.com/qor/qor5/publish"
	"github.com/theplant/sliceutils"
	"gorm.io/gorm"
)

type ListModel struct {
	ID    uint
	Title string

	publish.Status
	publish.List
}

func (this ListModel) TableName() string {
	return "list_models"
}

func (this ListModel) GetPublishActions(db *gorm.DB, ctx context.Context) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      this.getPublishUrl(),
		Content:  this.getPublishContent(),
		IsDelete: false,
	})

	if this.GetStatus() == publish.StatusOnline && this.GetOnlineUrl() != this.getPublishUrl() {
		objs = append(objs, &publish.PublishAction{
			Url:      this.GetOnlineUrl(),
			IsDelete: true,
		})
	}

	this.SetOnlineUrl(this.getPublishUrl())
	return
}
func (this ListModel) GetUnPublishActions(db *gorm.DB, ctx context.Context) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      this.GetOnlineUrl(),
		IsDelete: true,
	})
	return
}

func (this ListModel) getPublishUrl() string {
	return fmt.Sprintf("/list_model/%v/index.html", this.ID)
}

func (this ListModel) getPublishContent() string {
	return fmt.Sprintf("id: %v, title: %v", this.ID, this.Title)
}

func (this ListModel) GetListUrl(pageNumber int) string {
	return fmt.Sprintf("/list_model/list/%v.html", pageNumber)
}

func (this ListModel) GetListContent(onePageItems *publish.OnePageItems) string {
	pageNumber := onePageItems.PageNumber
	var result string
	for _, item := range onePageItems.Items {
		result = result + fmt.Sprintf("%v</br>", item)
	}
	result = result + fmt.Sprintf("</br>pageNumber: %v</br>", pageNumber)
	return result
}

func (this ListModel) Sort(array []interface{}) {
	var temp []*ListModel
	sliceutils.Unwrap(array, &temp)
	sort.Sort(ModelListModel(temp))
	for k, v := range temp {
		array[k] = v
	}
	return
}

type ModelListModel []*ListModel

func (x ModelListModel) Len() int           { return len(x) }
func (x ModelListModel) Less(i, j int) bool { return x[i].Title < x[j].Title }
func (x ModelListModel) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
