package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func AwsCliCheck() error {
	if _, err := exec.LookPath("aws"); err != nil {
		return fmt.Errorf("aws cli not available in $PATH")
	}

	return nil
}

type Credentials struct {
	Credentials struct {
		AccessKeyId     string
		SecretAccessKey string
		SessionToken    string
		Expiration      string
	}
}

func GetMfaCredentials(profile string, serialNumber string, tokenCode string) (Credentials, error) {
	cmd := exec.Command("aws", "sts", "get-session-token", "--profile", profile, "--serial-number", serialNumber, "--token-code", tokenCode)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	credentials := &Credentials{}

	err := cmd.Run()
	if err != nil {
		return *credentials, fmt.Errorf("%s", errOut.String())
	}

	if err := json.Unmarshal(out.Bytes(), credentials); err != nil {
		return *credentials, fmt.Errorf("error parsing JSON: %v", err)
	}

	return *credentials, nil
}
