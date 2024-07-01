package presets

import "fmt"

type PayloadModelsUpdated struct {
	Ids    []string `json:"ids"`
	Models []any    `json:"models"`
}

func (mb *ModelBuilder) NotifModelsUpdated() string {
	return fmt.Sprintf("PresetsModelsUpdated:%s", mb.modelType.String())
}

func NotifModelsUpdated(v any) string {
	return fmt.Sprintf("PresetsModelsUpdated:%T", v)
}

type PayloadModelsDeleted struct {
	Ids []string `json:"ids"`
}

func (mb *ModelBuilder) NotifModelsDeleted() string {
	return fmt.Sprintf("PresetsModelsDeleted:%s", mb.modelType.String())
}

func NotifModelsDeleted(v any) string {
	return fmt.Sprintf("PresetsModelsDeleted:%T", v)
}
