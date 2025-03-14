package emailbuilder

import (
	"context"

	"github.com/qor5/admin/v3/marketing/tag"
	"github.com/qor5/admin/v3/marketing/tag/bg"
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

func dummyTags(ctx context.Context) []*tag.CategoryWithBuilders {
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
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventLogin), "Logged In", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventViewPDP), "Viewed Products", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventAddToCart), "Added to Cart", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventBeginCheckout), "Began Checkout", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventConfirm), "Confirmed Orders", "activities"))
	registry.MustRegisterBuilder(bg.EventTagBuilder(string(EventPurchase), "Made Purchases", "activities"))

	return registry.GetCategoriesWithBuilders(ctx)
}
