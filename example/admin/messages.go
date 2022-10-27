package admin

import "github.com/qor/qor5/login"

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

var Messages_ja_JP = &Messages{
	QOR5Example: "QOR5 Example",
	Roles:       "Roles",
	Users:       "Users",
}

var Messages_login_ja_JP = &login.Messages{
	Confirm:                             "Confirm",
	Verify:                              "Verify",
	AccountLabel:                        "Email JP",
	AccountPlaceholder:                  "Email",
	PasswordLabel:                       "Password JP",
	PasswordPlaceholder:                 "Password",
	SignInBtn:                           "Sign In",
	ForgetPasswordLink:                  "Forget your password?",
	ForgotMyPasswordTitle:               "I forgot my password",
	ForgetPasswordEmailLabel:            "Enter your email",
	ForgetPasswordEmailPlaceholder:      "Email",
	SendResetPasswordEmailBtn:           "Send reset password email",
	ResendResetPasswordEmailBtn:         "Resend reset password email",
	SendEmailTooFrequentlyNotice:        "Sending emails too frequently, please try again later",
	ResetPasswordLinkWasSentTo:          "A reset password link was sent to",
	ResetPasswordLinkSentPrompt:         "You can close this page and reset your password from this link.",
	ResetYourPasswordTitle:              "Reset your password",
	ResetPasswordLabel:                  "Change your password",
	ResetPasswordPlaceholder:            "New password",
	ResetPasswordConfirmLabel:           "Re-enter new password",
	ResetPasswordConfirmPlaceholder:     "Confirm new password",
	ChangePasswordTitle:                 "Change your password",
	ChangePasswordOldLabel:              "Old password",
	ChangePasswordOldPlaceholder:        "Old Password",
	ChangePasswordNewLabel:              "New password",
	ChangePasswordNewPlaceholder:        "New Password",
	ChangePasswordNewConfirmLabel:       "Re-enter new password",
	ChangePasswordNewConfirmPlaceholder: "New Password",
	TOTPSetupTitle:                      "Two Factor Authentication",
	TOTPSetupScanPrompt:                 "Scan this QR code with Google Authenticator (or similar) app",
	TOTPSetupSecretPrompt:               "Or manually enter the following code into your preferred authenticator app",
	TOTPSetupEnterCodePrompt:            "Then enter the provided one-time code below",
	TOTPSetupCodePlaceholder:            "Passcode",
	TOTPValidateTitle:                   "Two Factor Authentication",
	TOTPValidateEnterCodePrompt:         "Enter the provided one-time code below",
	TOTPValidateCodeLabel:               "Authenticator passcode",
	TOTPValidateCodePlaceholder:         "Passcode",
	ErrorSystemError:                    "System Error",
	ErrorCompleteUserAuthFailed:         "Complete User Auth Failed",
	ErrorUserNotFound:                   "User Not Found",
	ErrorIncorrectAccountNameOrPassword: "Incorrect email or password",
	ErrorUserLocked:                     "User Locked",
	ErrorAccountIsRequired:              "Email is required",
	ErrorPasswordCannotBeEmpty:          "Password cannot be empty",
	ErrorPasswordNotMatch:               "Password do not match",
	ErrorIncorrectPassword:              "Old password is incorrect",
	ErrorInvalidToken:                   "Invalid token",
	ErrorTokenExpired:                   "Token expired",
	ErrorIncorrectTOTPCode:              "Incorrect passcode",
	ErrorTOTPCodeReused:                 "This passcode has been used",
	ErrorIncorrectRecaptchaToken:        "Incorrect reCAPTCHA token",
	WarnPasswordHasBeenChanged:          "Password has been changed, please sign-in again",
	InfoPasswordSuccessfullyReset:       "Password successfully reset, please sign-in again",
	InfoPasswordSuccessfullyChanged:     "Password successfully changed, please sign-in again",
}
