package tag

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

// Registry manages tag builders
type Registry struct {
	builders      []Builder
	buildersMap   map[string]Builder
	fragments     map[FragmentType]func() Fragment
	categories    []*Category
	categoriesMap map[string]*Category
	lock          sync.RWMutex
}

// NewRegistry creates a new registry instance
func NewRegistry() *Registry {
	return &Registry{
		builders:      []Builder{},
		buildersMap:   make(map[string]Builder),
		fragments:     make(map[FragmentType]func() Fragment),
		categories:    []*Category{},
		categoriesMap: make(map[string]*Category),
	}
}

// RegisterBuilder registers a builder
// Returns error if a builder with the same ID already exists
func (r *Registry) RegisterBuilder(builder Builder) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	metadata := builder.Metadata(context.Background())
	id := metadata.ID
	if _, exists := r.buildersMap[id]; exists {
		return errors.Errorf("builder with ID %q already registered", id)
	}

	// Check if the referenced category exists
	if metadata.CategoryID != "" {
		if _, exists := r.categoriesMap[metadata.CategoryID]; !exists {
			return errors.Errorf("category with ID %q not found", metadata.CategoryID)
		}
	}

	r.builders = append(r.builders, builder)
	r.buildersMap[id] = builder
	return nil
}

// MustRegisterBuilder registers a builder and panics if registration fails
func (r *Registry) MustRegisterBuilder(builder Builder) {
	if err := r.RegisterBuilder(builder); err != nil {
		panic(err)
	}
}

// GetBuilder retrieves a builder by ID
func (r *Registry) GetBuilder(id string) (Builder, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	builder, ok := r.buildersMap[id]
	return builder, ok
}

// MustGetBuilder retrieves a builder by ID and panics if it doesn't exist
func (r *Registry) MustGetBuilder(id string) Builder {
	builder, ok := r.GetBuilder(id)
	if !ok {
		panic(fmt.Sprintf("builder with ID %q not found", id))
	}
	return builder
}

// GetCategoriesWithBuilders returns all categories with their associated builder metadatas
func (r *Registry) GetCategoriesWithBuilders(ctx context.Context) []*CategoryWithBuilders {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// Create a map to group builder metadatas by category ID
	buildersByCategory := make(map[string][]*Metadata)

	// Collect all builder metadatas and group by category ID
	for _, builder := range r.builders {
		metadata := builder.Metadata(ctx)
		categoryID := metadata.CategoryID
		if categoryID != "" {
			buildersByCategory[categoryID] = append(buildersByCategory[categoryID], metadata)
		}
	}

	// Create the result array
	result := make([]*CategoryWithBuilders, 0, len(r.categories))

	// Populate categories with their associated builder metadatas
	for _, category := range r.categories {
		categoryWithBuilders := &CategoryWithBuilders{
			Category: category,
			Builders: buildersByCategory[category.ID],
		}
		result = append(result, categoryWithBuilders)
	}

	return result
}

// RegisterFragment registers a fragment factory function for a specific fragment type
func (r *Registry) RegisterFragment(fragmentType FragmentType, factory func() Fragment) error {
	fragment := factory()
	if fragment.Type() != fragmentType {
		return errors.Errorf("fragment type mismatch: expected %s, got %s", fragmentType, fragment.Type())
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	if _, exists := r.fragments[fragmentType]; exists {
		return errors.Errorf("fragment with type %q already registered", fragmentType)
	}

	r.fragments[fragmentType] = factory
	return nil
}

// MustRegisterFragment registers a fragment factory and panics if registration fails
func (r *Registry) MustRegisterFragment(fragmentType FragmentType, factory func() Fragment) {
	if err := r.RegisterFragment(fragmentType, factory); err != nil {
		panic(err)
	}
}

// GetFragmentFactory retrieves a fragment factory by type
func (r *Registry) GetFragmentFactory(fragmentType FragmentType) (func() Fragment, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	factory, ok := r.fragments[fragmentType]
	return factory, ok
}

// CreateFragment creates a new fragment instance of the specified type
func (r *Registry) CreateFragment(fragmentType FragmentType) (Fragment, error) {
	factory, ok := r.GetFragmentFactory(fragmentType)
	if !ok {
		return nil, errors.Errorf("fragment type %q not registered", fragmentType)
	}
	return factory(), nil
}

// RegisterCategory registers a category
// Returns error if a category with the same ID already exists
func (r *Registry) RegisterCategory(category *Category) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, exists := r.categoriesMap[category.ID]; exists {
		return errors.Errorf("category with ID %q already registered", category.ID)
	}

	r.categories = append(r.categories, category)
	r.categoriesMap[category.ID] = category
	return nil
}

// MustRegisterCategory registers a category and panics if registration fails
func (r *Registry) MustRegisterCategory(category *Category) {
	if err := r.RegisterCategory(category); err != nil {
		panic(err)
	}
}

// GetCategory retrieves a category by ID
func (r *Registry) GetCategory(id string) (*Category, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	category, ok := r.categoriesMap[id]
	return category, ok
}

// MustGetCategory retrieves a category by ID and panics if it doesn't exist
func (r *Registry) MustGetCategory(id string) *Category {
	category, ok := r.GetCategory(id)
	if !ok {
		panic(fmt.Sprintf("category with ID %q not found", id))
	}
	return category
}

// GetCategories returns all registered categories
func (r *Registry) GetCategories() []*Category {
	r.lock.RLock()
	defer r.lock.RUnlock()

	result := make([]*Category, len(r.categories))
	copy(result, r.categories)
	return result
}

// DefaultRegistry is the global registry instance
var DefaultRegistry = NewRegistry()

// RegisterBuilder registers a builder with the default registry
func RegisterBuilder(builder Builder) error {
	return DefaultRegistry.RegisterBuilder(builder)
}

// MustRegisterBuilder registers a builder with the default registry
// Panics if registration fails
func MustRegisterBuilder(builder Builder) {
	DefaultRegistry.MustRegisterBuilder(builder)
}

// GetBuilder retrieves a builder from the default registry
func GetBuilder(id string) (Builder, bool) {
	return DefaultRegistry.GetBuilder(id)
}

// MustGetBuilder retrieves a builder from the default registry
// Panics if the builder doesn't exist
func MustGetBuilder(id string) Builder {
	return DefaultRegistry.MustGetBuilder(id)
}

// RegisterFragment registers a fragment factory with the default registry
func RegisterFragment(fragmentType FragmentType, factory func() Fragment) error {
	return DefaultRegistry.RegisterFragment(fragmentType, factory)
}

// MustRegisterFragment registers a fragment factory with the default registry
// Panics if registration fails
func MustRegisterFragment(fragmentType FragmentType, factory func() Fragment) {
	DefaultRegistry.MustRegisterFragment(fragmentType, factory)
}

// GetFragmentFactory retrieves a fragment factory from the default registry
func GetFragmentFactory(fragmentType FragmentType) (func() Fragment, bool) {
	return DefaultRegistry.GetFragmentFactory(fragmentType)
}

// CreateFragment creates a new fragment instance using the default registry
func CreateFragment(fragmentType FragmentType) (Fragment, error) {
	return DefaultRegistry.CreateFragment(fragmentType)
}

// GetCategoriesWithBuilders returns all categories with their associated builder metadatas from the default registry
func GetCategoriesWithBuilders(ctx context.Context) []*CategoryWithBuilders {
	return DefaultRegistry.GetCategoriesWithBuilders(ctx)
}

// RegisterCategory registers a category with the default registry
func RegisterCategory(category *Category) error {
	return DefaultRegistry.RegisterCategory(category)
}

// MustRegisterCategory registers a category with the default registry
// Panics if registration fails
func MustRegisterCategory(category *Category) {
	DefaultRegistry.MustRegisterCategory(category)
}

// GetCategory retrieves a category from the default registry
func GetCategory(id string) (*Category, bool) {
	return DefaultRegistry.GetCategory(id)
}

// MustGetCategory retrieves a category from the default registry
// Panics if the category doesn't exist
func MustGetCategory(id string) *Category {
	return DefaultRegistry.MustGetCategory(id)
}

// GetCategories returns all registered categories from the default registry
func GetCategories() []*Category {
	return DefaultRegistry.GetCategories()
}
