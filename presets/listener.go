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
	Ids    []string       `json:"ids"`
	Models map[string]any `json:"models"`
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

func (*ModelBuilder) NotifRowUpdated() string {
	return "presets_NotifRowUpdated"
}

type PayloadRowUpdated struct {
	Id string `json:"id"`
}

func (mb *ModelBuilder) NotifModelsValidate() string {
	return fmt.Sprintf("presets_NotifModelsValidate_%T", mb.model)
}

func (mb *ModelBuilder) NotifModelsSectionValidate(name string) string {
	return fmt.Sprintf("presets_NotifModelsValidate_%v_%T", name, mb.model)
}

type PayloadModelsSetter struct {
	Id          string      `json:"id"`
	FieldErrors interface{} `json:"field_errors"`
	Passed      bool        `json:"passed"`
}
