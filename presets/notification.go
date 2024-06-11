package presets

import "fmt"

type PayloadModelsUpdated struct {
	Models []any `json:"models"`
}

func (mb *ModelBuilder) NotifModelsUpdated() string {
	return fmt.Sprintf("PresetsModelsUpdated:%s", mb.modelType.String())
}

type PayloadModelsDeleted struct {
	IDs []string `json:"ids"`
}

func (mb *ModelBuilder) NotifModelsDeleted() string {
	return fmt.Sprintf("PresetsModelsDeleted:%s", mb.modelType.String())
}
