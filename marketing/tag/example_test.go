package tag_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/qor5/admin/v3/marketing/tag"
	"github.com/qor5/admin/v3/marketing/tag/bg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EventName defines the type for event name constants
type EventName string

// Event name constants
const (
	EventLogin         EventName = "LOGIN"
	EventViewPDP       EventName = "VIEW_PDP"
	EventAddToCart     EventName = "ADD_TO_CART"
	EventBeginCheckout EventName = "BEGIN_CHECKOUT"
	EventConfirm       EventName = "CONFIRM"
	EventPurchase      EventName = "PURCHASE"
)

// Gender defines the type for gender constants
type Gender string

// Gender constants
const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

// City defines the type for city constants
type City string

// City constants for Japanese cities
const (
	CityTokyo     City = "TOKYO"
	CityOsaka     City = "OSAKA"
	CityKyoto     City = "KYOTO"
	CitySapporo   City = "SAPPORO"
	CityYokohama  City = "YOKOHAMA"
	CityNagoya    City = "NAGOYA"
	CityFukuoka   City = "FUKUOKA"
	CityHiroshima City = "HIROSHIMA"
)

// SignupSource defines the type for signup source constants
type SignupSource string

// SignupSource constants
const (
	SignupSourceWebsite       SignupSource = "WEBSITE"
	SignupSourceMobileApp     SignupSource = "MOBILE_APP"
	SignupSourceReferral      SignupSource = "REFERRAL"
	SignupSourceAdvertisement SignupSource = "ADVERTISEMENT"
)

// User represents a user entity
type User struct {
	UserID       string       `json:"userID"`
	CreatedAt    time.Time    `json:"createdAt"`
	Gender       Gender       `json:"gender"`
	Age          int          `json:"age"`
	City         City         `json:"city"`
	SignupSource SignupSource `json:"signupSource"`
	LastActive   time.Time    `json:"lastActive"`
}

// Event represents an event entity
type Event struct {
	EventID   string    `json:"eventID"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    string    `json:"userID"`
	EventName EventName `json:"eventName"`
}

// TestExpressionToSQL tests the conversion of tag expressions to SQL queries
func TestExpressionToSQL(t *testing.T) {
	// Create a new registry for tag builders
	registry := tag.NewRegistry()

	// Register categories for the test
	registry.MustRegisterCategory(&tag.Category{
		ID:          "demographics",
		Name:        "Demographics",
		Description: "Demographic filters",
	})

	registry.MustRegisterCategory(&tag.Category{
		ID:          "activities",
		Name:        "Activities",
		Description: "User activity filters",
	})

	// Register a string select tag for gender
	registry.MustRegisterBuilder(bg.StringTagBuilder(
		"user_gender",
		"User Gender",
		"Filter users by gender",
		"demographics",
		"gender",
		[]*tag.Option{
			{Label: "Male", Value: string(GenderMale)},
			{Label: "Female", Value: string(GenderFemale)},
		},
	))

	// Register string tag for user name
	registry.MustRegisterBuilder(bg.StringTagBuilder(
		"user_name",
		"User Name",
		"Search by user name",
		"demographics",
		"username",
		nil,
	))

	// Register event tag for purchase event
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventPurchase), "made purchases", "activities"))

	// Create a composite expression: male users with username starting with "john"
	// who made a purchase in the last 7 days
	expr := &tag.Expression{
		Intersect: []*tag.Expression{
			{
				Tag: &tag.Tag{
					BuilderID: "user_gender",
					Params: map[string]any{
						"operator": string(bg.StringOperatorEQ),
						"value":    string(GenderMale),
					},
				},
			},
			{
				Tag: &tag.Tag{
					BuilderID: "user_name",
					Params: map[string]any{
						"operator": string(bg.StringOperatorStartsWith),
						"value":    "john",
					},
				},
			},
			{
				Tag: &tag.Tag{
					BuilderID: "event_purchase",
					Params: map[string]any{
						"accumulation":  string(bg.Accumulation(bg.AccumulationCount)),
						"countOperator": string(bg.NumberOperatorGTE),
						"countValue":    float64(1),
						"timeRange":     string(bg.TimeRange7Days),
					},
				},
			},
		},
	}

	// Verify basic expression structure
	assert.Equal(t, 3, len(expr.Intersect))

	// Create a SQL processor with BigQuery dialect
	processor := tag.NewSQLProcessor(registry, bg.NewSQLDialect())

	// Process the expression with the processor
	simplifiedExpr := expr.Simplify()
	sql, err := processor.Process(context.Background(), simplifiedExpr)
	require.NoError(t, err, "Expression processing should succeed")

	// Log the actual SQL for debugging
	t.Logf("Generated Query:\n%s", sql.Query)
	t.Logf("Generated Args: %v", sql.Args)

	expectedQuery := `
	(SELECT user_id FROM users WHERE gender = ?) 
	INTERSECT DISTINCT 
	(SELECT user_id FROM users WHERE username LIKE CONCAT(?, '%')) 
	INTERSECT DISTINCT 
	(SELECT user_id FROM events WHERE event_name = 'PURCHASE' AND created_at >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY) GROUP BY user_id HAVING COUNT(1) >= ?)
	`
	assert.Equal(t, tag.CompactSQLQuery(expectedQuery), tag.CompactSQLQuery(sql.Query), "Generated SQL should match expected SQL")

	// Verify arguments
	expectedArgs := []any{
		string(GenderMale),
		"john",
		float64(1),
	}
	assert.Equal(t, expectedArgs, sql.Args, "SQL arguments should match expected values")
}

// ExampleSQLProcessor_Process demonstrates how to use SQLProcessor.Process method
// to transform tag expressions into SQL queries.
func ExampleSQLProcessor_Process() {
	// Create a new registry for tag builders
	registry := tag.NewRegistry()

	// Register categories for demo
	registry.MustRegisterCategory(&tag.Category{
		ID:          "demographics",
		Name:        "Demographics",
		Description: "Demographic filters",
	})

	registry.MustRegisterCategory(&tag.Category{
		ID:          "activities",
		Name:        "Activities",
		Description: "User activity filters",
	})

	// Register gender tag builder
	registry.MustRegisterBuilder(bg.StringTagBuilder(
		"user_gender",
		"User Gender",
		"Filter users by gender",
		"demographics",
		"gender",
		[]*tag.Option{
			{Label: "Male", Value: string(GenderMale)},
			{Label: "Female", Value: string(GenderFemale)},
			{Label: "Other", Value: string(GenderOther)},
		},
	))

	// Register age tag builder
	registry.MustRegisterBuilder(bg.NumberTagBuilder(
		"user_age",
		"User Age",
		"Filter users by age range",
		"demographics",
		"age",
		0,
		120,
	))

	// Register city tag builder
	registry.MustRegisterBuilder(bg.StringTagBuilder(
		"user_city",
		"User City",
		"Filter users by city",
		"demographics",
		"city",
		[]*tag.Option{
			{Label: "Tokyo", Value: string(CityTokyo)},
			{Label: "Osaka", Value: string(CityOsaka)},
			{Label: "Kyoto", Value: string(CityKyoto)},
			{Label: "Sapporo", Value: string(CitySapporo)},
			{Label: "Yokohama", Value: string(CityYokohama)},
			{Label: "Nagoya", Value: string(CityNagoya)},
			{Label: "Fukuoka", Value: string(CityFukuoka)},
			{Label: "Hiroshima", Value: string(CityHiroshima)},
		},
	))

	// Register signup source tag builder
	registry.MustRegisterBuilder(bg.StringTagBuilder(
		"user_signup_source",
		"User Signup Source",
		"Filter users by signup source",
		"demographics",
		"signupSource",
		[]*tag.Option{
			{Label: "Website", Value: string(SignupSourceWebsite)},
			{Label: "Mobile App", Value: string(SignupSourceMobileApp)},
			{Label: "Referral", Value: string(SignupSourceReferral)},
			{Label: "Advertisement", Value: string(SignupSourceAdvertisement)},
		},
	))

	registry.MustRegisterBuilder(bg.DateRangeTagBuilder(
		"user_last_active",
		"User Last Active",
		"Filter users by last active time range",
		"activities",
		"lastActive",
		true,
	))

	// Register event tags
	// Event tags now belong to the Activities category
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventLogin), "logged in", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventViewPDP), "viewed products", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventAddToCart), "added to cart", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventBeginCheckout), "began checkout", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventConfirm), "confirmed orders", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventPurchase), "made purchases", "activities"))

	// Get all registered categories with builders
	// categoriesWithBuilders := registry.GetCategoriesWithBuilders(context.Background())
	// jsonData, _ := json.Marshal(categoriesWithBuilders)
	// fmt.Println("All registered categories with builders:", string(jsonData))

	// Create a SQL processor for processing expressions
	processor := tag.NewSQLProcessor(registry, bg.NewSQLDialect())

	// Create a complex expression:
	exprJSON := `
{
    "intersect": [
        {
            "intersect": [
                {
                    "tag": {
                        "builderID": "user_gender",
                        "params": {
                            "operator": "EQ",
                            "value": "FEMALE"
                        }
                    }
                },
                {
                    "tag": {
                        "builderID": "user_age",
                        "params": {
                            "max": 35,
                            "min": 25,
                            "operator": "BETWEEN"
                        }
                    }
                },
                {
                    "tag": {
                        "builderID": "user_city",
                        "params": {
                            "operator": "IN",
                            "values": ["TOKYO", "OSAKA"]
                        }
                    }
                },
                {
					"union": [
						{
							"tag": {
								"builderID": "user_signup_source",
								"params": {
									"operator": "EQ",
									"value": "WEBSITE"
								}
							}
						},
						{
							"tag": {
								"builderID": "user_signup_source",
								"params": {
									"operator": "EQ",
									"value": "MOBILE_APP"
								}
							}
						}
					]
				}
            ]
        },
        {
            "tag": {
                "builderID": "event_purchase",
                "params": {
                    "accumulation": "DAYS",
                    "countOperator": "GTE",
                    "countValue": 2,
                    "timeRange": "30D"
                }
            }
        }
    ]
}
`

	var expr *tag.Expression
	err := json.Unmarshal([]byte(exprJSON), &expr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	simplifiedExpr := expr.Simplify()

	// Process the expression to generate SQL
	sql, err := processor.Process(context.Background(), simplifiedExpr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print that we generated SQL successfully
	fmt.Println("Generated complex SQL query successfully!")

	// Print the actual SQL query generated
	// Replace trailing spaces to match expected output format
	output := strings.ReplaceAll(sql.Query, " \n", "\n")
	fmt.Println(output)

	// Print the arguments
	fmt.Println("Arguments:")
	for i, arg := range sql.Args {
		fmt.Printf("  [%d] %v (%T)\n", i, arg, arg)
	}

	// Also demonstrate using the new specific event tag builders
	purchaseExpr := &tag.Expression{
		Intersect: []*tag.Expression{
			{
				Tag: &tag.Tag{
					BuilderID: "event_purchase", // PurchaseEventTagBuilder returns this ID
					Params: map[string]any{
						"accumulation":  string(bg.Accumulation(bg.AccumulationDays)),
						"countOperator": string(bg.NumberOperatorBetween),
						"countMin":      float64(1),
						"countMax":      float64(10),
						"timeRange":     string(bg.TimeRange30Days),
					},
				},
			},
		},
	}

	// Also process this expression
	purchaseSQL, err := processor.Process(context.Background(), purchaseExpr)
	if err != nil {
		fmt.Println("Error processing purchase expression:", err)
		return
	}
	fmt.Println("Purchase specific event query:\n", purchaseSQL.Query)
	fmt.Println("Purchase specific event args:\n", purchaseSQL.Args)

	// Output:
	// Generated complex SQL query successfully!
	// (SELECT user_id FROM users WHERE gender = ?)
	// INTERSECT DISTINCT
	// (SELECT user_id FROM users WHERE age BETWEEN ? AND ?)
	// INTERSECT DISTINCT
	// (SELECT user_id FROM users WHERE city IN (?, ?))
	// INTERSECT DISTINCT
	// ((SELECT user_id FROM users WHERE signupSource = ?)
	// UNION DISTINCT
	// (SELECT user_id FROM users WHERE signupSource = ?))
	// INTERSECT DISTINCT
	// (SELECT user_id FROM events
	// WHERE event_name = 'PURCHASE'
	// AND created_at >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
	// GROUP BY user_id
	// HAVING COUNT(DISTINCT DATE(created_at)) >= ?)
	// Arguments:
	//   [0] FEMALE (string)
	//   [1] 25 (float64)
	//   [2] 35 (float64)
	//   [3] TOKYO (string)
	//   [4] OSAKA (string)
	//   [5] WEBSITE (string)
	//   [6] MOBILE_APP (string)
	//   [7] 2 (float64)
	// Purchase specific event query:
	//  (SELECT user_id FROM events
	// WHERE event_name = 'PURCHASE'
	// AND created_at >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
	// GROUP BY user_id
	// HAVING COUNT(DISTINCT DATE(created_at)) BETWEEN ? AND ?)
	// Purchase specific event args:
	//  [1 10]
}
