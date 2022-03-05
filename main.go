package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type Credential struct {
	SecretAccessKey string
	SessionToken    string
	AccessKeyId     string
	Expiration      string
}

type Credentials struct {
	Credentials Credential
}

//need to check if the --file flag is set and if not try to find credentials file in user root

func main() {
	if _, err := exec.LookPath("aws"); err != nil {
		fmt.Println("aws cli not available in path")
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to find current users home dir")
	}

	credFilePtr := flag.String("f", filepath.Join(homeDir, ".aws", "credentials"), "AWS credentials filepath")
	serialNumPtr := flag.String("s", "", "arn of the mfa device")
	codePtr := flag.Int("c", 0, "MFA code")
	profilePtr := flag.String("p", "default", "Profile to authenticate against")
	mfaProfilePtr := flag.String("mfa-porfile", "sport80_mfa", "Profile to assign MFA credentials")

	flag.Parse()

	if *serialNumPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *codePtr == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if _, err := os.Stat(*credFilePtr); err != nil {
		fmt.Printf("Credentials file not found at: %s\n", *credFilePtr)
		os.Exit(1)
	}

	cmd := exec.Command("aws", "sts", "get-session-token", "--profile", *profilePtr, "--serial-number", *serialNumPtr, "--token-code", strconv.Itoa(*codePtr))
	var out bytes.Buffer
	cmd.Stdout = &out

	e := cmd.Run()

	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}

	c := &Credentials{}

	if err := json.Unmarshal([]byte(out.String()), &c); err != nil {
		fmt.Printf("error unmarshaling JSON: %v\n", err)
	}

	// fmt.Printf("%+v", c)

	input, err := ioutil.ReadFile(*credFilePtr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	profile := "[" + *mfaProfilePtr + "]"
	if bytes.Contains(input, []byte(profile)) == false {
		t := fmt.Sprintf("\n%s\naws_access_key_id = %s\naws_secret_access_key = %s\naws_session_token = %s", profile, c.Credentials.AccessKeyId, c.Credentials.SecretAccessKey, c.Credentials.SessionToken)
		input = append(input, t...)

		if err := ioutil.WriteFile(*credFilePtr, input, 0666); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Sucessfully updates aws credentials")

		os.Exit(0)
	}
}
