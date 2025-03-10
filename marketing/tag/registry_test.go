package tag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistry_RegisterNewBuilder verifies that new builders are correctly registered in the tag registry
// to ensure the tag system can be properly extended with custom implementations
func TestRegistry_RegisterNewBuilder(t *testing.T) {
	registry := NewRegistry()

	// Register category first
	registry.MustRegisterCategory(&Category{
		ID:          "new_users",
		Name:        "New Users",
		Description: "Filters for new user acquisition",
	})

	newBuilder := NewSQLTemplate(
		&Metadata{
			ID:          "new_user",
			Name:        "New User",
			Description: "Filter for new users registered within N days",
			CategoryID:  "new_users",
			View: &View{
				Fragments: []Fragment{
					&TextFragment{
						Text: "New user (registered within ",
					},
					&NumberInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:          "days",
							Required:     true,
							DefaultValue: float64(30),
							Validation: &Validation{
								Pattern:      "^\\d+$",
								ErrorMessage: "Please enter a valid number of days",
							},
						},
						Min: 0,
						Max: 100,
					},
					&TextFragment{
						Text: " days)",
					},
				},
			},
		},
		"SELECT uid FROM users WHERE DATEDIFF(CURRENT_DATE(), register_date) <= {{.days}}",
	)

	err := registry.RegisterBuilder(newBuilder)
	require.NoError(t, err)

	builder, exists := registry.GetBuilder("new_user")
	require.True(t, exists)

	sqlBuilder, ok := builder.(SQLBuilder)
	require.True(t, ok, "builder should implement SQLBuilder")
	sql, err := sqlBuilder.BuildSQL(context.Background(), map[string]any{
		"days": float64(7),
	})
	require.NoError(t, err)

	assert.Equal(t, CompactSQLQuery("SELECT uid FROM users WHERE DATEDIFF(CURRENT_DATE(), register_date) <= 7"), CompactSQLQuery(sql.Query))
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewRegistry()

	// Register category first
	registry.MustRegisterCategory(&Category{
		ID:          "demographic",
		Name:        "Demographics",
		Description: "Demographic filters",
	})

	registry.MustRegisterBuilder(createTestGenderTemplate("demographic"))

	duplicateBuilder := NewSQLTemplate(
		&Metadata{
			ID:          "gender", // Existing ID
			Name:        "Gender Duplicate",
			Description: "Duplicate builder with existing ID",
			CategoryID:  "demographic",
			View: &View{
				Fragments: []Fragment{
					&TextFragment{
						Text: "Duplicate",
					},
				},
			},
		},
		"SELECT 1",
	)

	err := registry.RegisterBuilder(duplicateBuilder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_GetBuilderNotFound(t *testing.T) {
	registry := NewRegistry()

	builder, exists := registry.GetBuilder("non_existent")
	assert.False(t, exists)
	assert.Nil(t, builder)
}

func TestRegistry_GetBuilders(t *testing.T) {
	registry := NewRegistry()

	// Register category first
	registry.MustRegisterCategory(&Category{
		ID:          "demographic",
		Name:        "Demographics",
		Description: "Demographic filters",
	})

	// Register several builders
	registry.MustRegisterBuilder(createTestGenderTemplate("demographic"))
	registry.MustRegisterBuilder(createTestAgeTemplate("demographic"))

	// Get all registered builders
	builders := make([]Builder, 0)
	for _, id := range []string{"gender", "age"} {
		if builder, exists := registry.GetBuilder(id); exists {
			builders = append(builders, builder)
		}
	}
	assert.Len(t, builders, 2)

	// Verify that all builders can be retrieved by ID
	ids := make(map[string]bool)
	ctx := context.Background()
	for _, builder := range builders {
		ids[builder.Metadata(ctx).ID] = true
	}

	assert.True(t, ids["gender"])
	assert.True(t, ids["age"])
}

func TestRegistry_RegisterCategory(t *testing.T) {
	registry := NewRegistry()

	demographicCategory := &Category{
		ID:          "demographic",
		Name:        "Demographics",
		Description: "Demographic filters",
	}

	behaviorCategory := &Category{
		ID:          "behavior",
		Name:        "Behavior",
		Description: "Behavioral filters",
	}

	// Register categories
	err := registry.RegisterCategory(demographicCategory)
	require.NoError(t, err)

	err = registry.RegisterCategory(behaviorCategory)
	require.NoError(t, err)

	// Test duplicate registration
	err = registry.RegisterCategory(&Category{
		ID:   "demographic",
		Name: "Duplicate",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test GetCategory
	category, exists := registry.GetCategory("demographic")
	assert.True(t, exists)
	assert.Equal(t, "Demographics", category.Name)

	// Test category not found
	category, exists = registry.GetCategory("non_existent")
	assert.False(t, exists)
	assert.Nil(t, category)

	// Test GetCategories
	categories := registry.GetCategories()
	assert.Len(t, categories, 2)
	assert.Equal(t, "demographic", categories[0].ID)
	assert.Equal(t, "behavior", categories[1].ID)
}

func TestRegistry_GetCategoriesWithBuilders(t *testing.T) {
	registry := NewRegistry()

	// Register categories
	registry.MustRegisterCategory(&Category{
		ID:          "demographic",
		Name:        "Demographics",
		Description: "Demographic filters",
	})

	registry.MustRegisterCategory(&Category{
		ID:          "behavior",
		Name:        "Behavior",
		Description: "Behavioral filters",
	})

	// Register builders
	registry.MustRegisterBuilder(createTestGenderTemplate("demographic"))
	registry.MustRegisterBuilder(createTestAgeTemplate("demographic"))
	registry.MustRegisterBuilder(createTestPurchaseTemplate("behavior"))

	// Get all categories with builders
	ctx := context.Background()
	categoriesWithBuilders := registry.GetCategoriesWithBuilders(ctx)

	// Verify categories and builders
	assert.Len(t, categoriesWithBuilders, 2)

	// Find demographic category
	var demographicCategory *CategoryWithBuilders
	var behaviorCategory *CategoryWithBuilders

	for _, cwb := range categoriesWithBuilders {
		if cwb.ID == "demographic" {
			demographicCategory = cwb
		} else if cwb.ID == "behavior" {
			behaviorCategory = cwb
		}
	}

	// Verify demographic category
	require.NotNil(t, demographicCategory)
	assert.Equal(t, "Demographics", demographicCategory.Name)
	assert.Len(t, demographicCategory.Builders, 2)

	// Check builder IDs in demographic category
	builderIDs := make(map[string]bool)
	for _, b := range demographicCategory.Builders {
		builderIDs[b.ID] = true
	}
	assert.True(t, builderIDs["gender"])
	assert.True(t, builderIDs["age"])

	// Verify behavior category
	require.NotNil(t, behaviorCategory)
	assert.Equal(t, "Behavior", behaviorCategory.Name)
	assert.Len(t, behaviorCategory.Builders, 1)
	assert.Equal(t, "purchase_amount", behaviorCategory.Builders[0].ID)

	// No need to test the global DefaultRegistry since it's environment-dependent
}
