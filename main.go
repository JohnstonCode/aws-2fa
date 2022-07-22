package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/JohnstonCode/aws-2fa/config"
	"github.com/JohnstonCode/aws-2fa/credentials"
)

func main() {
	profilePtr := flag.String("profile", "default", "Profile to authenticate against")
	mfaProfile := flag.String("mfa-profile", "", "Profile to assign MFA credentials")
	serialNumber := flag.String("serial-number", "", "arn of the mfa device")
	tokenCode := flag.String("token-code", "", "MFA code")

	flag.Parse()

	if len(*mfaProfile) == 0 {
		fmt.Println("You must spesify a mfa-profile")
		os.Exit(1)
	}

	if len(*serialNumber) == 0 {
		fmt.Println("You must spesify a serial-number")
		os.Exit(1)
	}

	if len(*tokenCode) == 0 {
		fmt.Println("You must spesify a token-code")
		os.Exit(1)
	}

	path, err := config.GetConfigPath()
	if err != nil {
		fmt.Println("Unable to get home dir")
		os.Exit(1)
	}

	config, err := config.LoadConfig(path)
	if err != nil {
		fmt.Printf("Error loading config %s\n", err)
		os.Exit(1)
	}

	if !config.SectionExists(*profilePtr) {
		log.Printf("credentails file doesnt have %s section", *profilePtr)
		os.Exit(1)
	}

	credentials, err := credentials.GetCredentials(*profilePtr, *serialNumber, *tokenCode)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config.SetValue(*mfaProfile, "aws_access_key_id", *credentials.Credentials.AccessKeyId)
	config.SetValue(*mfaProfile, "aws_secret_access_key", *credentials.Credentials.SecretAccessKey)
	config.SetValue(*mfaProfile, "aws_session_token", *credentials.Credentials.SessionToken)

	if err := config.Save(); err != nil {
		fmt.Println("Unable to save credential file changes")
		os.Exit(1)
	}
}
