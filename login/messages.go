package login

type Messages struct {
	Email                       string
	Password                    string
	SignIn                      string
	ForgetPass                  string
	TwoFactorAuthentication     string
	TwoFactorPrompt             string
	Verify                      string
	ChangeYourPassword          string
	ResetYourPassword           string
	ForgotMyPassword            string
	EnterYourEmail              string
	AResetPasswordLinkWasSentTo string
	ResetPasswordPrompt         string
	ReenterNewPassword          string
	Confirm                     string
	OldPassword                 string
	NewPassword                 string
	TOTPScanQRCodeWithApp       string
	TOTPEnterCodeWithApp        string
	TOTPThenEnterPassCode       string
	TOTPEnterPassCode           string
	SendResetPasswordEmail      string
	ResendResetPasswordEmail    string
}

var Messages_en_US = &Messages{
	Email:                       "Email",
	Password:                    "Password",
	SignIn:                      "Sign In",
	ForgetPass:                  "Forget your password?",
	TwoFactorAuthentication:     "Two Factor Authentication",
	TwoFactorPrompt:             "Enter the provided one-time code below",
	Verify:                      "Verify",
	ChangeYourPassword:          "Change your password",
	ResetYourPassword:           "Reset your password",
	ForgotMyPassword:            "I forgot my password",
	EnterYourEmail:              "Enter your email",
	AResetPasswordLinkWasSentTo: "A reset password link was sent to",
	ResetPasswordPrompt:         "You can close this page and reset your password from this link.",
	ReenterNewPassword:          "Re-enter new password",
	Confirm:                     "Confirm",
	OldPassword:                 "Old password",
	NewPassword:                 "New password",
	TOTPScanQRCodeWithApp:       "Scan this QR code with Google Authenticator (or similar) app",
	TOTPEnterCodeWithApp:        "Or manually enter the following code into your preferred authenticator app",
	TOTPThenEnterPassCode:       "Then enter the provided one-time code below",
	TOTPEnterPassCode:           "Enter your passcode here",
	SendResetPasswordEmail:      "Send reset password email",
	ResendResetPasswordEmail:    "Resend reset password email",
}

var Messages_zh_CN = &Messages{
	Email:                       "邮箱",
	Password:                    "密码",
	SignIn:                      "登录",
	ForgetPass:                  "忘记密码？",
	TwoFactorAuthentication:     "双重认证",
	TwoFactorPrompt:             "在下面输入提供的一次性代码",
	Verify:                      "验证",
	ChangeYourPassword:          "改变您的密码",
	ResetYourPassword:           "重置您的密码",
	ForgotMyPassword:            "我忘记密码了",
	EnterYourEmail:              "输入您的电子邮箱",
	AResetPasswordLinkWasSentTo: "已将重置密码链接发送到",
	ResetPasswordPrompt:         "您可以关闭此页面并从此链接重置密码。",
	ReenterNewPassword:          "重新输入新密码",
	Confirm:                     "确认",
	OldPassword:                 "旧密码",
	NewPassword:                 "新密码",
	TOTPScanQRCodeWithApp:       "使用Google Authenticator（或类似）应用程序扫描此二维码",
	TOTPEnterCodeWithApp:        "或者将以下代码手动输入到您首选的验证器应用程序中",
	TOTPThenEnterPassCode:       "然后在下面输入提供的一次性代码",
	TOTPEnterPassCode:           "在此处输入您的一次性代码",
	SendResetPasswordEmail:      "发送重置密码电子邮件",
	ResendResetPasswordEmail:    "重新发送重置密码电子邮件",
}
