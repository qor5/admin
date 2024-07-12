package presets

import (
	"fmt"
)

type PayloadModelsCreated struct {
	Models []any `json:"models"`
}

func (mb *ModelBuilder) NotifModelsCreated() string {
	return fmt.Sprintf("presets_NotifModelsCreated_%T", mb.model)
}

func NotifModelsCreated(v any) string {
	return fmt.Sprintf("presets_NotifModelsCreated_%T", v)
}

type PayloadModelsUpdated struct {
	Ids    []string `json:"ids"`
	Models []any    `json:"models"`
}

func (mb *ModelBuilder) NotifModelsUpdated() string {
	return fmt.Sprintf("presets_NotifModelsUpdated_%T", mb.model)
}

func NotifModelsUpdated(v any) string {
	return fmt.Sprintf("presets_NotifModelsUpdated_%T", v)
}

type PayloadModelsDeleted struct {
	Ids []string `json:"ids"`
}

func (mb *ModelBuilder) NotifModelsDeleted() string {
	return fmt.Sprintf("presets_NotifModelsDeleted_%T", mb.model)
}

func NotifModelsDeleted(v any) string {
	return fmt.Sprintf("presets_NotifModelsDeleted_%T", v)
}

func (mb *ModelBuilder) NotifRowUpdated() string {
	return fmt.Sprintf("presets_NotifRowUpdated")
}

type PayloadRowUpdated struct {
	Id string `json:"id"`
}
