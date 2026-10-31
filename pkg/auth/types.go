package auth

type createOwnerAccountReq struct {
	Email       string `json:"email" binding:"required,email"`
	FirstName   string `json:"fname" binding:"required"`
	LastName    string `json:"lname" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

type passwordLoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Type     string `json:"accountType" binding:"required,oneof=tenant admin"`
}

type setPasswordReq struct {
	Password string `json:"password" binding:"required,min=8"`
	Token    string `json:"token" binding:"required"`
}

type refreshAuthTokenReq struct {
	AccessToken  string `json:"accessToken" binding:"required"`
	RefreshToken string `json:"refreshToken" binding:"required"`
}
