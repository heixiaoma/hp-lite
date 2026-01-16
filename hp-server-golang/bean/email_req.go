package bean

type ReqSendCode struct {
	Email string `json:"email"`
	Type  string `json:"type"` // verify_email, reset_password
}

type ReqVerifyCode struct {
	Email string `json:"email"`
	Code  string `json:"code"`
	Type  string `json:"type"` // verify_email, reset_password
}

type ReqResetPassword struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type ReqSetEmail struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ReqChangePassword struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}
