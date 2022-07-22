package flags

import (
	"errors"
	"flag"
)

type Flags struct {
	Profile      string
	MfaProfile   string
	SerialNumber string
	TokenCode    string
}

func ParseFlags() (*Flags, error) {
	profilePtr := flag.String("profile", "default", "Profile to authenticate against")
	mfaProfile := flag.String("mfa-profile", "", "Profile to assign MFA credentials")
	serialNumber := flag.String("serial-number", "", "arn of the mfa device")
	tokenCode := flag.String("token-code", "", "MFA code")

	flag.Parse()

	if len(*mfaProfile) == 0 {
		return nil, errors.New("you must spesify a mfa-profile")
	}

	if len(*serialNumber) == 0 {
		return nil, errors.New("you must spesify a serial-number")
	}

	if len(*tokenCode) == 0 {
		return nil, errors.New("you must spesify a token-code")
	}

	flags := &Flags{
		Profile:      *profilePtr,
		MfaProfile:   *mfaProfile,
		SerialNumber: *serialNumber,
		TokenCode:    *tokenCode,
	}

	return flags, nil
}
