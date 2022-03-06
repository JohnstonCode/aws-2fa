package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	Comment      = "#"
	Splitter     = "="
	ProfileStart = "["
	ProfileEnd   = "]"
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

type Section struct {
	items        map[string]string
	order        []string
	itemComments map[string][]string
	comments     []string
}

type Config struct {
	sections map[string]*Section
	order    []string
}

func (c *Config) Parse(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	var (
		lineNum  int
		line     string
		profile  string
		comments []string
		section  *Section
		scanner  = bufio.NewScanner(f)
		idx      int
		key      string
		value    string
	)

	for scanner.Scan() {
		lineNum++
		line = scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) == 0 || strings.HasPrefix(line, Comment) {
			comments = append(comments, line)
			continue
		}

		if strings.HasPrefix(line, ProfileStart) {
			if !strings.HasSuffix(line, ProfileEnd) {
				return errors.New(fmt.Sprintf("no end profile: %s at %d", ProfileEnd, lineNum))
			}

			profile := line[1 : len(line)-1]
			s, ok := c.sections[profile]
			if !ok {
				s = &Section{items: map[string]string{}, itemComments: map[string][]string{}, comments: comments}
				c.sections[profile] = s
				c.order = append(c.order, profile)
			} else {
				return errors.New(fmt.Sprintf("profile: %s already exists at %d", profile, lineNum))
			}

			section = s
			comments = []string{}
			continue
		}

		idx = strings.Index(line, Splitter)
		if idx > 0 {
			key = strings.TrimSpace(line[:idx])
			if len(line) > idx {
				value = strings.TrimSpace(line[idx+1:])
			}
		} else {
			return errors.New(fmt.Sprintf("no splitter in key: %s at %d", line, lineNum))
		}

		if section == nil {
			return errors.New(fmt.Sprintf("no profile for key: %s at %d", key, lineNum))
		}

		if _, ok := section.items[key]; ok {
			return errors.New(fmt.Sprintf("section: %s already has key: %s at %d", profile, key, lineNum))
		}

		section.order = append(section.order, key)
		section.items[key] = value
		section.itemComments[key] = comments
		comments = []string{}
	}

	return nil
}

func (c *Config) HasSection(section string) bool {
	_, ok := c.sections[section]

	return ok
}

func (c *Config) Save(file string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, order := range c.order {
		section, _ := c.sections[order]

		for _, comment := range section.comments {
			if _, err := f.WriteString(fmt.Sprintf("%s\n", comment)); err != nil {
				return err
			}
		}

		if _, err := f.WriteString(fmt.Sprintf("[%s]\n", order)); err != nil {
			return err
		}

		for _, k := range section.order {
			v, _ := section.items[k]

			for _, comment := range section.itemComments[k] {
				if _, err := f.WriteString(fmt.Sprintf("%s\n", comment)); err != nil {
					return err
				}
			}

			if _, err := f.WriteString(fmt.Sprintf("%s = %s\n", k, v)); err != nil {
				return err
			}
		}
	}

	return nil
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

	config := &Config{sections: map[string]*Section{}}
	if err := config.Parse(*credFilePtr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !config.HasSection(*mfaProfilePtr) {
		config.order = append(config.order, *mfaProfilePtr)
		config.sections[*mfaProfilePtr] = &Section{
			items: map[string]string{
				"aws_access_key_id":     "",
				"aws_secret_access_key": "",
				"aws_session_token":     "",
			},
			comments: []string{""},
			order: []string{
				"aws_access_key_id",
				"aws_secret_access_key",
				"aws_session_token",
			},
		}
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

	config.sections[*mfaProfilePtr].items["aws_access_key_id"] = c.Credentials.AccessKeyId
	config.sections[*mfaProfilePtr].items["aws_secret_access_key"] = c.Credentials.SecretAccessKey
	config.sections[*mfaProfilePtr].items["aws_session_token"] = c.Credentials.SessionToken

	if err := config.Save(*credFilePtr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated aws MFA credentials")
	os.Exit(0)
}
