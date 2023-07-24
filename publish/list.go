package publish

// @snippet_begin(PublishList)
type List struct {
	PageNumber  int
	Position    int
	ListDeleted bool
	ListUpdated bool
}

// @snippet_end

func (this List) GetPageNumber() int {
	return this.PageNumber
}

func (this *List) SetPageNumber(pageNumber int) {
	this.PageNumber = pageNumber
}

func (this List) GetPosition() int {
	return this.Position
}

func (this *List) SetPosition(position int) {
	this.Position = position
}

func (this List) GetListDeleted() bool {
	return this.ListDeleted
}

func (this *List) SetListDeleted(listDeleted bool) {
	this.ListDeleted = listDeleted
}

func (this List) GetListUpdated() bool {
	return this.ListUpdated
}

func (this *List) SetListUpdated(listUpdated bool) {
	this.ListUpdated = listUpdated
}
