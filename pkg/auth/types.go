package auth

type createOwnerAccountReq struct {
	Email       string `json:"email"`
	FirstName   string `json:"fname"`
	LastName    string `json:"lname"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber"`
}

type passwordLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"accountType"`
}

type refreshAuthTokenReq struct {
	AccessToken  string `json:"accessToken" binding:"required"`
	RefreshToken string `json:"refreshToken" binding:"required"`
}
