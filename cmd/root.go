package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var configFilepath string
var serialNumber string
var mfaCode string
var profile string
var mfaProfile string

type Credential struct {
	SecretAccessKey string
	SessionToken    string
	AccessKeyId     string
	Expiration      string
}

type Credentials struct {
	Credentials Credential
}

var rootCmd = &cobra.Command{
	Use:     "aws-2fa [flags]",
	Short:   "Update aws mfa access automatically",
	Long:    "Update aws mfa access automatically",
	Example: "aws-2fa -s arn-of-the-mfa-device -m 12345 --mfa-profile mfa_profile",
	Run: func(cobra *cobra.Command, args []string) {
		if _, err := exec.LookPath("aws"); err != nil {
			fmt.Println("aws cli not available in path")
			os.Exit(1)
		}

		if _, err := os.Stat(configFilepath); err != nil {
			fmt.Printf("Credentials file not found at: %s\n", configFilepath)
			os.Exit(1)
		}

		cmd := exec.Command("aws", "sts", "get-session-token", "--profile", profile, "--serial-number", serialNumber, "--token-code", mfaCode)
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut

		e := cmd.Run()

		if e != nil {
			fmt.Println(errOut.String())
			os.Exit(1)
		}

		c := &Credentials{}

		if err := json.Unmarshal([]byte(out.String()), &c); err != nil {
			fmt.Printf("error unmarshaling JSON: %v\n", err)
			os.Exit(1)
		}

		file, err := ioutil.ReadFile(configFilepath)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		fileString := string(file)

		r, _ := regexp.Compile(fmt.Sprintf(`(?m)^\[%s]`, mfaProfile))
		if !r.MatchString(fileString) {
			fileString = strings.TrimSpace(fileString)
			placeholder := "\n\n[%s]\naws_access_key_id = %s\naws_secret_access_key = %s\naws_session_token = %s"
			fileString = fileString + fmt.Sprintf(placeholder, mfaProfile, c.Credentials.AccessKeyId, c.Credentials.SecretAccessKey, c.Credentials.SessionToken)

			os.WriteFile(configFilepath, []byte(fileString), 0644)
		} else {
			fmt.Println("Replace")
			//Use regex to overwrite
		}

		fmt.Println("Sucessfully updates aws credentials")
	},
}

func Execute() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to find current users home dir")
	}

	rootCmd.Flags().StringVarP(&configFilepath, "config", "c", filepath.Join(homeDir, ".aws", "credentials"), "AWS credentials filepath")
	rootCmd.Flags().StringVarP(&serialNumber, "serial-number", "s", "", "arn of the mfa device")
	rootCmd.Flags().StringVarP(&mfaCode, "mfa-code", "m", "", "MFA code")
	rootCmd.Flags().StringVarP(&profile, "profile", "p", "default", "Profile to authenticate against")
	rootCmd.Flags().StringVarP(&mfaProfile, "mfa-profile", "", "", "Profile to assign MFA credentials")

	rootCmd.MarkFlagRequired("serial-number")
	rootCmd.MarkFlagRequired("mfa-code")
	rootCmd.MarkFlagRequired("mfa-profile")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
