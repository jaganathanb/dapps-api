package dto

type GetOtpRequest struct {
	MobileNumber string `json:"mobileNumber" binding:"required,mobile,min=11,max=11"`
}

type TokenDetail struct {
	AccessToken            string `json:"accessToken"`
	RefreshToken           string `json:"refreshToken"`
	AccessTokenExpireTime  int64  `json:"accessTokenExpireTime"`
	RefreshTokenExpireTime int64  `json:"refreshTokenExpireTime"`
}

type RegisterUserByUsernameRequest struct {
	FirstName string `json:"firstName" binding:"required,min=3"`
	LastName  string `json:"lastName" binding:"required,min=6"`
	Username  string `json:"username" binding:"required,min=5"`
	Email     string `json:"email" binding:"min=6,email"`
	Password  string `json:"password" binding:"required,password,min=6"`
}

type RegisterLoginByMobileRequest struct {
	MobileNumber string `json:"mobileNumber" binding:"required,mobile,min=11,max=11"`
	Otp          string `json:"otp" binding:"required,min=6,max=6"`
}

type LoginByUsernameRequest struct {
	Username string `json:"username" binding:"required,min=5"`
	Password string `json:"password" binding:"required,min=6"`
}

type LogoutByUsernameRequest struct {
	Username string `json:"username" binding:"required,min=5"`
}

type User struct {
	BaseDto
	Username     string      `json:"userName"`
	FirstName    string      `json:"firstName"`
	LastName     string      `json:"lastName"`
	MobileNumber string      `json:"mobileNumber"`
	Email        string      `json:"email"`
	Enabled      bool        `json:"enabled"`
	UserRoles    *[]UserRole `json:"userRoles"`
}

type Role struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type UserRole struct {
	Role Role `json:"role"`
}
