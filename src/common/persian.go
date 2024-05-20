package common

import (
	"log"
	"regexp"
)

const indianMobileNumberPattern string = `^[6-9]\d{9}$`

func IndianMobileNumberValidate(mobileNumber string) bool {
	res, err := regexp.MatchString(indianMobileNumberPattern, mobileNumber)
	if err != nil {
		log.Print(err.Error())
	}
	return res
}
