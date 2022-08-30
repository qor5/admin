package login

type Messages struct {
	Email                   string
	Password                string
	SignIn                  string
	ForgetPass              string
	TwoFactorAuthentication string
	TwoFactorPrompt         string
}

var Messages_en_US = &Messages{
	Email:                   "Email",
	Password:                "Password",
	SignIn:                  "Sign In",
	ForgetPass:              "Forget your password?",
	TwoFactorAuthentication: "Two Factor Authentication",
	TwoFactorPrompt:         "Enter the provided one-time code below",
}

var Messages_zh_CN = &Messages{
	Email:                   "邮箱",
	Password:                "密码",
	SignIn:                  "登录",
	ForgetPass:              "忘记密码？",
	TwoFactorAuthentication: "双重认证",
	TwoFactorPrompt:         "在下面输入提供的一次性代码",
}
