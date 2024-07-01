package basics

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var Login = Doc(
	Markdown(`
Login package provides comprehensive login authentication logic and related UI interfaces. It is designed to simplify the process of adding user authentication to QOR5-based backend development project.   
In QOR5 admin development, we recommend using [github.com/qor5/admin/login](https://github.com/qor5/admin/tree/main/login), which wraps [github.com/qor5/x/login](https://github.com/qor5/x/tree/master/login) to keep the theme of login UI consistent with Presets and provide more powerful features.
## Basic Usage
The example shows how to enable both username/password login and OAuth login.
    `),
	ch.Code(generated.LoginBasicUsage).Language("go"),
	Markdown(`
## Username/Password Login
To enable Username/Password login, the ~UserModel~ needs to implement the [UserPasser](https://github.com/qor5/x/blob/8f986dddfeaf235fd42bb3361717551d06695517/login/user_pass.go#L13) interface. There is a default implementation - [UserPass](https://github.com/qor5/x/blob/8f986dddfeaf235fd42bb3361717551d06695517/login/user_pass.go#L44).  
    `),
	ch.Code(generated.LoginEnableUserPass).Language("go"),
	Markdown(`
### Change Password
There are three ways to change the password:

1\. Visit the default change password page.

2\. Call the ~OpenChangePasswordDialogEvent~ event to change it in dialog.
    `),
	ch.Code(generated.LoginOpenChangePasswordDialog).Language("go"),
	Markdown(`

3\. Change the password directly in Editing.
    `),
	ch.Code(generated.LoginChangePasswordInEditing).Language("go"),
	Markdown(`
### MaxRetryCount
By default, it allows 5 login attempts with incorrect credentials, and if the limit is exceeded, the user will be locked for 1 hour. This helps to prevent brute-force attacks on the login system. You can call ~MaxRetryCount~ to set the maximum retry count. If you set MaxRetryCount to a value less than or equal to 0, it means there is no limit of login attempts, and the user will not be locked after a certain number of failed login attempts. 
    `),
	ch.Code(generated.LoginSetMaxRetryCount).Language("go"),
	Markdown(`
### TOTP
There is TOTP (Time-based One-time Password) functionality out of the box, which is enabled by default.
    `),
	ch.Code(generated.LoginSetTOTP).Language("go"),
	Markdown(`
### Google reCAPTCHA
Google reCAPTCHA is disabled by default.
    `),
	ch.Code(generated.LoginSetRecaptcha).Language("go"),
	Markdown(`
## OAuth Login
OAuth login is based on [goth](https://github.com/markbates/goth).   
OAuth login does not require a ~UserModel~. If there is a ~UserModel~, it needs to implement the [OAuthUser](https://github.com/qor5/x/blob/8f986dddfeaf235fd42bb3361717551d06695517/login/oauth_user.go#L5) interface. There is a default implementation - [OAuthInfo](https://github.com/qor5/x/blob/8f986dddfeaf235fd42bb3361717551d06695517/login/oauth_user.go#L13).  
    `),
	ch.Code(generated.LoginEnableOAuth).Language("go"),
	Markdown(`
## Session Secure
The [SessionSecurer](https://github.com/qor5/x/blob/8f986dddfeaf235fd42bb3361717551d06695517/login/session_secure.go#L11) provides a way to manage unique salt for a user record. There is a default implementation - [SessionSecure](https://github.com/qor5/x/blob/8f986dddfeaf235fd42bb3361717551d06695517/login/session_secure.go#L16).  
    `),
	ch.Code(generated.LoginEnableSessionSecure).Language("go"),
	Markdown(`
~SessionSecurer~ helps to ensure user security even in the event of secret leakage. When a user logs in, ~SessionSecurer~ generates a random salt and associates it with the user's record. This salt is then used to sign the user's session token. When the user makes requests to the server, the server verifies that the session token has been signed with the correct salt. If the salt has been changed, the session token is considered invalid and the user is logged out.
    `),
	Markdown(`
## Hooks
[Hooks](https://github.com/qor5/x/blob/8f986dddfeaf235fd42bb3361717551d06695517/login/log_builder.go#L39) are functions that are called before or after certain events.   
The following hooks are available:
### BeforeSetPassword
#### Extra Values
- password

This hook is called before resetting or changing a password. The hook can be used to validate password formats.
### AfterLogin
This hook is called after a successful login.
### AfterFailedToLogin
#### Extra Values
- login error

This hook is called after a failed login. Note that the ~user~ parameter may be nil.
### AfterUserLocked
This hook is called after a user is locked.
### AfterLogout
This hook is called after a logout.
### AfterConfirmSendResetPasswordLink
#### Extra Values
- reset link

This hook is called after confirming the sending of a password reset link. This is where the code to send the reset link to the user should be written.
### AfterResetPassword
This hook is called after a password is reset.
### AfterChangePassword
This hook is called after a password is changed.
### AfterExtendSession
#### Extra Values
- old session token

This hook is called after a session is extended.
### AfterTOTPCodeReused
This hook is called after a TOTP code has been reused.
### AfterOAuthComplete
This hook is called after an OAuth authentication is completed.
    `),
	Markdown(`
## Customize Pages
To customize pages, there are two ways:

1\. Each page has a corresponding ~xxxPageFunc~ to rewrite the page content. You can easily customize a page by copying the [default page func](https://github.com/qor5/admin/blob/main/login/views.go) and modifying it according to your needs.
    `),
	ch.Code(generated.LoginCustomizePage).Language("go"),
	Markdown(`
2\. Only mount the API and serve the login pages manually.  
When you want to embed the login form into an existing page, this way can be very useful.
    `),
	ch.Code(generated.LoginCustomizePage2).Language("go"),
).Slug("basics/login").Title("Login")
