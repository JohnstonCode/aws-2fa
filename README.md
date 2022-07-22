# Usage

This will run `aws sts get-session-token` will the required flags and then update `~/.aws/credentials`

By default the default profile in `~/.aws/credentials` is used. This can be changed by using the `-profile` flag

Example usage:

```
aws-2fa -serial-number arn:aws:iam::123456789:mfa/test -token-code 12345 -mfa-profile mfa_profile
```

See all possible options
```
aws-2fa --help
```