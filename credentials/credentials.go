package credentials

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func GetCredentials(profile string, serialNumber string, tokenCode string) (*sts.GetSessionTokenOutput, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", profile),
	})
	if err != nil {
		return nil, err
	}

	svc := sts.New(sess)
	input := &sts.GetSessionTokenInput{
		SerialNumber: aws.String(serialNumber),
		TokenCode:    aws.String(tokenCode),
	}

	var awsErr error
	result, err := svc.GetSessionToken(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sts.ErrCodeRegionDisabledException:
				awsErr = errors.New(fmt.Sprint(sts.ErrCodeRegionDisabledException, aerr.Error()))
			default:
				awsErr = errors.New(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			awsErr = errors.New(err.Error())
		}

		return nil, awsErr
	}

	return result, nil
}
