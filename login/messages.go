package login

import "github.com/goplaid/x/i18n"

const I18nLoginKey i18n.ModuleKey = "I18nLoginKey"

type Messages struct {
	// common
	Confirm string
	Verify  string
	// login page
	AccountLabel        string
	AccountPlaceholder  string
	PasswordLabel       string
	PasswordPlaceholder string
	SignInBtn           string
	ForgetPasswordLink  string
	// forget password page
	ForgotMyPasswordTitle          string
	ForgetPasswordEmailLabel       string
	ForgetPasswordEmailPlaceholder string
	SendResetPasswordEmailBtn      string
	ResendResetPasswordEmailBtn    string
	SendEmailTooFrequentlyNotice   string
	// reset password link sent page
	ResetPasswordLinkWasSentTo  string
	ResetPasswordLinkSentPrompt string
	// reset password page
	ResetYourPasswordTitle          string
	ResetPasswordLabel              string
	ResetPasswordPlaceholder        string
	ResetPasswordConfirmLabel       string
	ResetPasswordConfirmPlaceholder string
	// change password page
	ChangePasswordTitle                 string
	ChangePasswordOldLabel              string
	ChangePasswordOldPlaceholder        string
	ChangePasswordNewLabel              string
	ChangePasswordNewPlaceholder        string
	ChangePasswordNewConfirmLabel       string
	ChangePasswordNewConfirmPlaceholder string
	// TOTP setup page
	TOTPSetupTitle           string
	TOTPSetupScanPrompt      string
	TOTPSetupSecretPrompt    string
	TOTPSetupEnterCodePrompt string
	TOTPSetupCodePlaceholder string
	// TOTP validate page
	TOTPValidateTitle           string
	TOTPValidateEnterCodePrompt string
	TOTPValidateCodeLabel       string
	TOTPValidateCodePlaceholder string
	// Error Messages
	ErrorSystemError                    string
	ErrorCompleteUserAuthFailed         string
	ErrorUserNotFound                   string
	ErrorIncorrectAccountNameOrPassword string
	ErrorUserLocked                     string
	ErrorAccountIsRequired              string
	ErrorPasswordCannotBeEmpty          string
	ErrorPasswordNotMatch               string
	ErrorIncorrectPassword              string
	ErrorInvalidToken                   string
	ErrorTokenExpired                   string
	ErrorIncorrectTOTP                  string
	ErrorIncorrectRecaptchaToken        string
	// Warn Messages
	WarnPasswordHasBeenChanged string
	// Info Messages
	InfoPasswordSuccessfullyReset   string
	InfoPasswordSuccessfullyChanged string
}

var Messages_en_US = &Messages{
	Confirm:                             "Confirm",
	Verify:                              "Verify",
	AccountLabel:                        "Email",
	AccountPlaceholder:                  "Email",
	PasswordLabel:                       "Password",
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
	ErrorIncorrectTOTP:                  "Incorrect passcode",
	ErrorIncorrectRecaptchaToken:        "Incorrect reCAPTCHA token",
	WarnPasswordHasBeenChanged:          "Password has been changed, please sign-in again",
	InfoPasswordSuccessfullyReset:       "Password successfully reset, please sign-in again",
	InfoPasswordSuccessfullyChanged:     "Password successfully changed, please sign-in again",
}

var Messages_zh_CN = &Messages{
	Confirm:                             "确认",
	Verify:                              "验证",
	AccountLabel:                        "邮箱",
	AccountPlaceholder:                  "邮箱",
	PasswordLabel:                       "密码",
	PasswordPlaceholder:                 "密码",
	SignInBtn:                           "登录",
	ForgetPasswordLink:                  "忘记密码？",
	ForgotMyPasswordTitle:               "我忘记密码了",
	ForgetPasswordEmailLabel:            "输入您的电子邮箱",
	ForgetPasswordEmailPlaceholder:      "电子邮箱",
	SendResetPasswordEmailBtn:           "发送重置密码电子邮件",
	ResendResetPasswordEmailBtn:         "重新发送重置密码电子邮件",
	SendEmailTooFrequentlyNotice:        "邮件发送过于频繁，请稍后再试",
	ResetPasswordLinkWasSentTo:          "已将重置密码链接发送到",
	ResetPasswordLinkSentPrompt:         "您可以关闭此页面并从此链接重置密码。",
	ResetYourPasswordTitle:              "重置您的密码",
	ResetPasswordLabel:                  "改变您的密码",
	ResetPasswordPlaceholder:            "新密码",
	ResetPasswordConfirmLabel:           "再次输入新密码",
	ResetPasswordConfirmPlaceholder:     "新密码",
	ChangePasswordTitle:                 "修改您的密码",
	ChangePasswordOldLabel:              "旧密码",
	ChangePasswordOldPlaceholder:        "旧密码",
	ChangePasswordNewLabel:              "新密码",
	ChangePasswordNewPlaceholder:        "新密码",
	ChangePasswordNewConfirmLabel:       "再次输入新密码",
	ChangePasswordNewConfirmPlaceholder: "新密码",
	TOTPSetupTitle:                      "双重认证",
	TOTPSetupScanPrompt:                 "使用Google Authenticator（或类似）应用程序扫描此二维码",
	TOTPSetupSecretPrompt:               "或者将以下代码手动输入到您首选的验证器应用程序中",
	TOTPSetupEnterCodePrompt:            "然后在下面输入提供的一次性代码",
	TOTPSetupCodePlaceholder:            "passcode",
	TOTPValidateTitle:                   "双重认证",
	TOTPValidateEnterCodePrompt:         "在下面输入提供的一次性代码",
	TOTPValidateCodeLabel:               "Authenticator验证码",
	TOTPValidateCodePlaceholder:         "passcode",
	ErrorSystemError:                    "系统错误",
	ErrorCompleteUserAuthFailed:         "用户认证失败",
	ErrorUserNotFound:                   "找不到该用户",
	ErrorIncorrectAccountNameOrPassword: "邮箱或密码错误",
	ErrorUserLocked:                     "用户已锁定",
	ErrorAccountIsRequired:              "邮箱是必须的",
	ErrorPasswordCannotBeEmpty:          "密码不能为空",
	ErrorPasswordNotMatch:               "确认密码不匹配",
	ErrorIncorrectPassword:              "密码错误",
	ErrorInvalidToken:                   "token无效",
	ErrorTokenExpired:                   "token过期",
	ErrorIncorrectTOTP:                  "passcode错误",
	ErrorIncorrectRecaptchaToken:        "reCAPTCHA token错误",
	WarnPasswordHasBeenChanged:          "密码被修改了，请重新登录",
	InfoPasswordSuccessfullyReset:       "密码重置成功，请重新登录",
	InfoPasswordSuccessfullyChanged:     "密码修改成功，请重新登录",
}
