package constants

const (
	// User
	AdminRoleName      string = "admin"
	DefaultRoleName    string = "default"
	DefaultUserName    string = "admin"
	RedisOtpDefaultKey string = "otp"

	// Claims
	AuthorizationHeaderKey string = "Authorization"
	UserIdKey              string = "UserId"
	FirstNameKey           string = "FirstName"
	LastNameKey            string = "LastName"
	UsernameKey            string = "Username"
	EmailKey               string = "Email"
	MobileNumberKey        string = "MobileNumber"
	RolesKey               string = "Roles"
	ExpireTimeKey          string = "Exp"
)

type GstReturnType int

const (
	GSTR1  GstReturnType = 0
	GSTR2B GstReturnType = 1
	GSTR9  GstReturnType = 2
)

type GstReturnStatus int

const (
	INVOICE_CALL     = 0
	INVOICE_RECEIVED = 1
	INVOICE_ENTRY    = 2
	FILED            = 3
)
