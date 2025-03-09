package tag

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
)

// MarshalJSON implements custom JSON marshaling for View
func (v View) MarshalJSON() ([]byte, error) {
	if v.Fragments == nil {
		return []byte(`{"fragments":null}`), nil
	}

	fragmentsData := make([]json.RawMessage, 0, len(v.Fragments))

	for _, fragment := range v.Fragments {
		fragmentData, err := json.Marshal(fragment)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal fragment %q", fragment.Metadata().Key)
		}

		var fragmentMap map[string]any
		if err := json.Unmarshal(fragmentData, &fragmentMap); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal fragment %q", fragment.Metadata().Key)
		}

		fragmentMap["type"] = fragment.Type()

		enhancedData, err := json.Marshal(fragmentMap)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to re-marshal fragment %q", fragment.Metadata().Key)
		}

		fragmentsData = append(fragmentsData, enhancedData)
	}

	viewAlias := struct {
		Fragments []json.RawMessage `json:"fragments"`
	}{
		Fragments: fragmentsData,
	}

	return json.Marshal(viewAlias)
}

// UnmarshalJSON implements custom JSON unmarshaling for View
func (v *View) UnmarshalJSON(data []byte) error {
	return v.UnmarshalJSONWithRegistry(DefaultRegistry, data)
}

// UnmarshalJSONWithRegistry implements custom JSON unmarshaling for View with a specific registry
func (v *View) UnmarshalJSONWithRegistry(registry *Registry, data []byte) error {
	var viewAlias struct {
		Fragments []json.RawMessage `json:"fragments"`
	}

	if err := json.Unmarshal(data, &viewAlias); err != nil {
		return errors.Wrap(err, "failed to unmarshal view")
	}

	if len(viewAlias.Fragments) == 0 {
		v.Fragments = nil
		return nil
	}

	v.Fragments = make([]Fragment, 0, len(viewAlias.Fragments))

	var typeInfo struct {
		Type FragmentType `json:"type"`
	}

	for _, fragmentData := range viewAlias.Fragments {
		if err := json.Unmarshal(fragmentData, &typeInfo); err != nil {
			return errors.Wrapf(err, "failed to determine fragment type")
		}

		fragment, err := registry.CreateFragment(typeInfo.Type)
		if err != nil {
			return errors.Errorf("unknown fragment type: %q", typeInfo.Type)
		}

		if err := json.Unmarshal(fragmentData, &fragment); err != nil {
			return errors.Wrapf(err, "failed to unmarshal %q", typeInfo.Type)
		}

		v.Fragments = append(v.Fragments, fragment)
	}

	return nil
}

// Validate validates the view with the given parameters
func (v *View) Validate(ctx context.Context, params map[string]any) error {
	for _, fragment := range v.Fragments {
		validator, ok := fragment.(Validator)
		if !ok {
			continue
		}
		if err := validator.Validate(ctx, params); err != nil {
			if errors.Is(err, ErrorShouldSkipValidate) {
				continue
			}
			return err
		}
	}
	return nil
}
