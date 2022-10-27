package admin

import (
	"github.com/goplaid/x/presets"
	l10n_view "github.com/qor/qor5/l10n/views"
	"github.com/qor/qor5/login"
	"github.com/qor/qor5/note"
	publish_view "github.com/qor/qor5/publish/views"
	"github.com/qor/qor5/utils"
)

type Messages struct {
	QOR5Example string
	Roles       string
	Users       string

	Posts          string
	PostsID        string
	PostsTitle     string
	PostsHeroImage string
	PostsBody      string
	Example        string
	Settings       string
	Post           string
	PostsBodyImage string

	SeoPost             string
	SeoVariableTitle    string
	SeoVariableSiteName string

	PageBuilder              string
	Pages                    string
	SharedContainers         string
	DemoContainers           string
	Templates                string
	Categories               string
	ProductManagement        string
	Products                 string
	SiteManagement           string
	SEO                      string
	UserManagement           string
	Profile                  string
	FeaturedModelsManagement string
	InputHarnesses           string
	ListEditorExample        string
	Customers                string
	ListModels               string
	MicrositeModels          string
	Workers                  string
	ActivityLogs             string
	MediaLibrary             string

	PagesID         string
	PagesTitle      string
	PagesSlug       string
	PagesLocale     string
	PagesNotes      string
	PagesDraftCount string
	PagesOnline     string

	Page                   string
	PagesStatus            string
	PagesSchedule          string
	PagesCategoryID        string
	PagesTemplateSelection string
	PagesEditContainer     string
}

var Messages_zh_CN = &Messages{
	Posts:          "帖子",
	PostsID:        "ID",
	PostsTitle:     "标题",
	PostsHeroImage: "主图",
	PostsBody:      "内容",
	Example:        "QOR5演示",
	Settings:       "设置",
	Post:           "帖子",
	PostsBodyImage: "内容图片",

	SeoPost:             "帖子",
	SeoVariableTitle:    "标题",
	SeoVariableSiteName: "站点名称",
}

var Messages_ja_JP_ModelsI18nModuleKey = &Messages{
	QOR5Example: "QOR5 Example JP",
	Roles:       "Roles JP",
	Users:       "Users JP",

	PageBuilder:              "Page Builder JP JP",
	Pages:                    "Pages JP",
	SharedContainers:         "Shared Containers JP",
	DemoContainers:           "Demo Containers JP",
	Templates:                "Templates JP",
	Categories:               "Categories JP",
	ProductManagement:        "Product Management JP",
	Products:                 "Products JP",
	SiteManagement:           "Site Management JP",
	SEO:                      "SEO JP",
	UserManagement:           "User Management JP",
	Profile:                  "Profile JP",
	FeaturedModelsManagement: "Featured Models Management JP",
	InputHarnesses:           "InputHarnesses JP",
	ListEditorExample:        "List Editor Example JP",
	Customers:                "Customers JP",
	ListModels:               "ListModels JP",
	MicrositeModels:          "MicrositeModels JP",
	Workers:                  "Workers JP",
	ActivityLogs:             "ActivityLogs JP",
	MediaLibrary:             "Media Library JP",
	Settings:                 "Settings JP",
	Posts:                    "Posts JP",

	PagesID:         "ID JP",
	PagesTitle:      "Title JP",
	PagesSlug:       "Slug JP",
	PagesLocale:     "Locale JP",
	PagesNotes:      "Notes JP",
	PagesDraftCount: "Draft Count JP",
	PagesOnline:     "Online JP",

	Page:                   "Page JP",
	PagesStatus:            "Status JP",
	PagesSchedule:          "Schedule JP",
	PagesCategoryID:        "Category ID JP",
	PagesTemplateSelection: "Template Selection JP",
	PagesEditContainer:     "Edit Container JP",
}

var Messages_ja_JP_I18nLoginKey = &login.Messages{
	Confirm:                             "Confirm JP",
	Verify:                              "Verify JP",
	AccountLabel:                        "Email JP JP",
	AccountPlaceholder:                  "Email JP",
	PasswordLabel:                       "Password JP JP",
	PasswordPlaceholder:                 "Password JP",
	SignInBtn:                           "Sign In JP",
	ForgetPasswordLink:                  "Forget your password? JP",
	ForgotMyPasswordTitle:               "I forgot my password JP",
	ForgetPasswordEmailLabel:            "Enter your email JP",
	ForgetPasswordEmailPlaceholder:      "Email JP",
	SendResetPasswordEmailBtn:           "Send reset password email JP",
	ResendResetPasswordEmailBtn:         "Resend reset password email JP",
	SendEmailTooFrequentlyNotice:        "Sending emails too frequently, please try again later JP",
	ResetPasswordLinkWasSentTo:          "A reset password link was sent to JP",
	ResetPasswordLinkSentPrompt:         "You can close this page and reset your password from this link. JP",
	ResetYourPasswordTitle:              "Reset your password JP",
	ResetPasswordLabel:                  "Change your password JP",
	ResetPasswordPlaceholder:            "New password JP",
	ResetPasswordConfirmLabel:           "Re-enter new password JP",
	ResetPasswordConfirmPlaceholder:     "Confirm new password JP",
	ChangePasswordTitle:                 "Change your password JP",
	ChangePasswordOldLabel:              "Old password JP",
	ChangePasswordOldPlaceholder:        "Old Password JP",
	ChangePasswordNewLabel:              "New password JP",
	ChangePasswordNewPlaceholder:        "New Password JP",
	ChangePasswordNewConfirmLabel:       "Re-enter new password JP",
	ChangePasswordNewConfirmPlaceholder: "New Password JP",
	TOTPSetupTitle:                      "Two Factor Authentication JP",
	TOTPSetupScanPrompt:                 "Scan this QR code with Google Authenticator (or similar) app JP",
	TOTPSetupSecretPrompt:               "Or manually enter the following code into your preferred authenticator app JP",
	TOTPSetupEnterCodePrompt:            "Then enter the provided one-time code below JP",
	TOTPSetupCodePlaceholder:            "Passcode JP",
	TOTPValidateTitle:                   "Two Factor Authentication JP",
	TOTPValidateEnterCodePrompt:         "Enter the provided one-time code below JP",
	TOTPValidateCodeLabel:               "Authenticator passcode JP",
	TOTPValidateCodePlaceholder:         "Passcode JP",
	ErrorSystemError:                    "System Error JP",
	ErrorCompleteUserAuthFailed:         "Complete User Auth Failed JP",
	ErrorUserNotFound:                   "User Not Found JP",
	ErrorIncorrectAccountNameOrPassword: "Incorrect email or password JP",
	ErrorUserLocked:                     "User Locked JP",
	ErrorAccountIsRequired:              "Email is required JP",
	ErrorPasswordCannotBeEmpty:          "Password cannot be empty JP",
	ErrorPasswordNotMatch:               "Password do not match JP",
	ErrorIncorrectPassword:              "Old password is incorrect JP",
	ErrorInvalidToken:                   "Invalid token JP",
	ErrorTokenExpired:                   "Token expired JP",
	ErrorIncorrectTOTPCode:              "Incorrect passcode JP",
	ErrorTOTPCodeReused:                 "This passcode has been used JP",
	ErrorIncorrectRecaptchaToken:        "Incorrect reCAPTCHA token JP",
	WarnPasswordHasBeenChanged:          "Password has been changed, please sign-in again JP",
	InfoPasswordSuccessfullyReset:       "Password successfully reset, please sign-in again JP",
	InfoPasswordSuccessfullyChanged:     "Password successfully changed, please sign-in again JP",
}

var Messages_ja_JP_I18nUtilsKey = &utils.Messages{
	OK:     "OK JP",
	Cancel: "Cancel JP",
}

var Messages_ja_JP_I18nPublishKey = &publish_view.Messages{
	StatusDraft:             "Draft JP",
	StatusOnline:            "Online JP",
	StatusOffline:           "Offline JP",
	Publish:                 "Publish JP",
	Unpublish:               "Unpublish JP",
	Republish:               "Republish JP",
	Areyousure:              "Are you sure? JP",
	ScheduledStartAt:        "Start at JP",
	ScheduledEndAt:          "End at JP",
	PublishedAt:             "Start at JP",
	UnPublishedAt:           "End at JP",
	ActualPublishTime:       "Actual Publish Time JP",
	SchedulePublishTime:     "Schedule Publish Time JP",
	NotSet:                  "Not set JP",
	WhenDoYouWantToPublish:  "When do you want to publish? JP",
	PublishScheduleTip:      "After you set the {SchedulePublishTime}, the system will automatically publish/unpublish it. JP",
	DateTimePickerClearText: "Clear JP",
	DateTimePickerOkText:    "OK JP",
	SaveAsNewVersion:        "Save As New Version JP",
	SwitchedToNewVersion:    "Switched To New Version JP",
	SuccessfullyCreated:     "Successfully Created JP",
	SuccessfullyRename:      "Successfully Rename JP",
	OnlineVersion:           "Online Version JP",
	VersionsList:            "Versions List JP",
	AllVersions:             "All versions JP",
	NamedVersions:           "Named versions JP",
	RenameVersion:           "Rename Version JP",
}

var Messages_ja_JP_CoreI18nModuleKey = &presets.Messages{
	SuccessfullyUpdated:            "Successfully Updated JP",
	Search:                         "Search JP",
	New:                            "New JP",
	Update:                         "Update JP",
	Delete:                         "Delete JP",
	Edit:                           "Edit JP",
	FormTitle:                      "Form JP",
	OK:                             "OK JP",
	Cancel:                         "Cancel JP",
	Create:                         "Create JP",
	DeleteConfirmationTextTemplate: "Are you sure you want to delete object with id: {id}? JP",
	CreatingObjectTitleTemplate:    "New {modelName} JP",
	EditingObjectTitleTemplate:     "Editing {modelName} {id} JP",
	ListingObjectTitleTemplate:     "Listing {modelName} JP",
	DetailingObjectTitleTemplate:   "{modelName} {id} JP",
	FiltersClear:                   "Clear Filters JP",
	FiltersAdd:                     "Add Filters JP",
	FilterApply:                    "Apply JP",
	FilterByTemplate:               "Filter by {filter} JP",
	FiltersDateInTheLast:           "is in the last JP",
	FiltersDateEquals:              "is equal to JP",
	FiltersDateBetween:             "is between JP",
	FiltersDateIsAfter:             "is after JP",
	FiltersDateIsAfterOrOn:         "is on or after JP",
	FiltersDateIsBefore:            "is before JP",
	FiltersDateIsBeforeOrOn:        "is before or on JP",
	FiltersDateDays:                "days JP",
	FiltersDateMonths:              "months JP",
	FiltersDateAnd:                 "and JP",
	FiltersDateTo:                  "to JP",
	FiltersNumberEquals:            "is equal to JP",
	FiltersNumberBetween:           "between JP",
	FiltersNumberGreaterThan:       "is greater than JP",
	FiltersNumberLessThan:          "is less than JP",
	FiltersNumberAnd:               "and JP",
	FiltersStringEquals:            "is equal to JP",
	FiltersStringContains:          "contains JP",
	FiltersMultipleSelectIn:        "in JP",
	FiltersMultipleSelectNotIn:     "not in JP",
	PaginationRowsPerPage:          "Rows per page:  JP",
	ListingNoRecordToShow:          "No records to show JP",
	ListingSelectedCountNotice:     "{count} records are selected.  JP",
	ListingClearSelection:          "clear selection JP",
	BulkActionNoAvailableRecords:   "None of the selected records can be executed with this action. JP",
	BulkActionSelectedIdsProcessNoticeTemplate: "Partially selected records cannot be executed with this action: {ids}. JP",
	ConfirmationDialogText:                     "Are you sure? JP",
	Language:                                   "Language JP",
	Colon:                                      ":",
}

var Messages_ja_JP_I18nNoteKey = &note.Messages{
	SuccessfullyCreated: "Successfully Created JP",
	Item:                "Item JP",
	Notes:               "Notes JP",
	NewNote:             "New Note JP",
}

var Messages_ja_JP_I10nLocalizeKey = &l10n_view.Messages{
	Localize:              "Localize JP",
	LocalizeFrom:          "From JP",
	LocalizeTo:            "To JP",
	SuccessfullyLocalized: "Successfully Localized JP",
	Location:              "Location JP",
	Colon:                 ":",
	International:         "International JP",
	China:                 "China JP",
	Japan:                 "Japan JP",
}
