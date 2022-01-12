package listeditor

type Sorter struct {
	Items []SorterItem `json:"items"`
}

type SorterItem struct {
	Index int    `json:"index"`
	Label string `json:"label"`
}
