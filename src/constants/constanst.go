package constants

const (
	DOF    = "02-01-2006"
	TAXPRD = "012006"
)

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

	// API
	Version uint = 1
)

type GstReturnType string

const (
	GSTR1  GstReturnType = "GSTR1"
	GSTR3B GstReturnType = "GSTR3B"
	GSTR9  GstReturnType = "GSTR9"
)

type GstReturnStatus string

const (
	InvoiceCall     = "InvoiceCall"
	InvoiceReceived = "InvoiceReceived"
	InvoiceEntry    = "InvoiceEntry"
	Filed           = "Filed"
)

func (d GstReturnType) String() string {
	return string(d)
}

func (d GstReturnStatus) String() string {
	return string(d)
}
