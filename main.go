package main

import (
	"fmt"
	"log"
	"os"

	"github.com/JohnstonCode/aws-2fa/config"
	"github.com/JohnstonCode/aws-2fa/credentials"
	"github.com/JohnstonCode/aws-2fa/flags"
)

func main() {
	flags, err := flags.ParseFlags()
	if err != nil {
		fmt.Println(err)
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

	if !config.SectionExists(flags.Profile) {
		log.Printf("credentails file doesnt have %s section", flags.Profile)
		os.Exit(1)
	}

	credentials, err := credentials.GetCredentials(flags.Profile, flags.SerialNumber, flags.TokenCode)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config.SetValue(flags.MfaProfile, "aws_access_key_id", *credentials.Credentials.AccessKeyId)
	config.SetValue(flags.MfaProfile, "aws_secret_access_key", *credentials.Credentials.SecretAccessKey)
	config.SetValue(flags.MfaProfile, "aws_session_token", *credentials.Credentials.SessionToken)

	if err := config.Save(); err != nil {
		fmt.Println("Unable to save credential file changes")
		os.Exit(1)
	}
}
