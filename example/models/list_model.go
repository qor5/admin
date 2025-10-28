package models

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/x/v3/oss"
	"github.com/theplant/sliceutils"
	"gorm.io/gorm"
)

type ListModel struct {
	gorm.Model

	Title string

	publish.Status
	publish.Schedule
	publish.List
	publish.Version
}

func (this *ListModel) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", this.ID, this.Version.Version)
}

func (this *ListModel) PermissionRN() []string {
	return []string{"list_models", strconv.Itoa(int(this.ID)), this.Version.Version}
}

func (this *ListModel) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	return map[string]string{
		"id":      segs[0],
		"version": segs[1],
	}
}

func (this *ListModel) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	actions = append(actions, &publish.PublishAction{
		Url:      this.getPublishUrl(),
		Content:  this.getPublishContent(),
		IsDelete: false,
	})

	if this.Status.Status == publish.StatusOnline && this.OnlineUrl != this.getPublishUrl() {
		actions = append(actions, &publish.PublishAction{
			Url:      this.OnlineUrl,
			IsDelete: true,
		})
	}

	this.OnlineUrl = this.getPublishUrl()
	return
}

func (this *ListModel) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	actions = append(actions, &publish.PublishAction{
		Url:      this.OnlineUrl,
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

func (this ListModel) GetListUrl(pageNumber string) string {
	return fmt.Sprintf("/list_model/list/%v.html", pageNumber)
}

func (this ListModel) GetListContent(db *gorm.DB, onePageItems *publish.OnePageItems) string {
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
	sort.Sort(SliceListModel(temp))
	for k, v := range temp {
		array[k] = v
	}
}

type SliceListModel []*ListModel

func (x SliceListModel) Len() int           { return len(x) }
func (x SliceListModel) Less(i, j int) bool { return x[i].Title < x[j].Title }
func (x SliceListModel) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
