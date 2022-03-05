package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

	f, err := os.OpenFile(*credFilePtr, os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Unable to read %s", *credFilePtr)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(f)

	var lines []string
	profile := "[" + *mfaProfilePtr + "]"
	replace := false
	found := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, profile) {
			replace = true
			found = true
		}

		if replace && strings.Contains(line, "aws_access_key_id") {
			lines = append(lines, fmt.Sprintf("aws_access_key_id = %s", c.Credentials.SecretAccessKey))
			continue
		}

		if replace && strings.Contains(line, "naws_secret_access_key") {
			lines = append(lines, fmt.Sprintf("aws_secret_access_key = %s", c.Credentials.SecretAccessKey))
			continue
		}

		if replace && strings.Contains(line, "naws_session_token") {
			lines = append(lines, fmt.Sprintf("aws_session_token = %s", c.Credentials.SessionToken))
			continue
		}

		if replace && strings.Contains(line, "[") {
			replace = false
		}

		lines = append(lines, line)
	}

	if found == false {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%s", profile))
		lines = append(lines, fmt.Sprintf("aws_access_key_id = %s", c.Credentials.AccessKeyId))
		lines = append(lines, fmt.Sprintf("aws_secret_access_key = %s", c.Credentials.SecretAccessKey))
		lines = append(lines, fmt.Sprintf("aws_session_token = %s", c.Credentials.SessionToken))
	}

	// fmt.Println(strings.Join(lines, "\n"))
	f.Truncate(0)
	f.Seek(0, 0)

	if _, err := f.WriteString(strings.Join(lines, "\n")); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	f.Close()

	fmt.Println("Successfully updated aws MFA credentials")
	os.Exit(0)
}
