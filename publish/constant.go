package publish

var (
	NonVersionPublishModels map[string]interface{}
	VersionPublishModels    map[string]interface{}
	ListPublishModels       map[string]interface{}
)

func init() {
	NonVersionPublishModels = make(map[string]interface{})
	VersionPublishModels = make(map[string]interface{})
	ListPublishModels = make(map[string]interface{})
}
