package service_errors

const (
	// Token
	UnExpectedError = "Expected error"
	ClaimsNotFound  = "Claims not found"
	TokenRequired   = "token required"
	TokenExpired    = "token expired"
	TokenInvalid    = "token invalid"

	// OTP
	OptExists   = "Otp exists"
	OtpUsed     = "Otp used"
	OtpNotValid = "Otp invalid"

	// User
	EmailExists      = "Email exists"
	UsernameExists   = "Username exists"
	PermissionDenied = "Permission denied"

	// DB
	RecordNotFound = "record not found"

	// GST
	GstNotFound = "GST %s does not exists in the system"
	GstExists   = "GST %s already exists in the system"
)
