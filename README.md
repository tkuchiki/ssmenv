# ssmenv
Expand environment variables from AWS EC2 Systems Manager Parameter Store

# Installation

Download from https://github.com/tkuchiki/ssmenv/releases

# Usage

```console
$ ./ssmenv --help
usage: ssmenv [<flags>]

Expand environment variables from AWS EC2 Systems Manager Parameter Store

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --names=PARAM_NAME,...         Names of the parameters (comma separated)
  --paths=PARAM_PATH,...         The hierarchy for the parameter names (comma separated)
  --tags=KEY=VALUE,...           Filter by tags (comma separated)
  --types=String,SecureString    The type of parameters (comma separated)
  --multi-values=PARAM_NAME,...  Names or paths with multiple values (comma separated)
  --without-export               Without export
  --hide-exists                  Hide environment variables if it already exists
  --fail-exists                  Fail if environment variables alerady exists
  --access-key=ACCESS-KEY        The AWS access key ID
  --secret-key=SECRET-KEY        The AWS secret access key
  --assume-role-arn=ASSUME-ROLE-ARN
                                 The AWS assume role ARN
  --token=TOKEN                  The AWS access token
  --region=REGION                The AWS region
  --profile=PROFILE              The AWS CLI profile
  --aws-config=AWS-CONFIG        The AWS CLI Config file
  --credentials=CREDENTIALS      The AWS CLI Credential file
  --version                      Show application version.
```

# Examples

```console
$ aws ssm put-parameter --name /test/env1 --value "foo" --type String --region us-east-1
$ aws ssm put-parameter --name /test/env2 --value "bar" --type String --region us-east-1

$ ./ssmenv --paths /test
export ENV1="foo"
export ENV2="bar"

$ eval $(./ssmenv --paths /test)
$ echo $ENV1
foo
$ echo $ENV2
bar
```

```console
$ aws ssm put-parameter --name test.env1 --value "foo" --type String --region us-east-1
$ aws ssm put-parameter --name test.env2 --value "bar" --type String --region us-east-1

$ ./ssmenv --names test.env1,test.env2
export ENV1="foo"
export ENV2="bar"

$ eval $(./ssmenv --names test.env1,test.env2)
$ echo $ENV1
foo
$ echo $ENV2
bar
```

```console
$ aws ssm put-parameter --name /test/env1 --value "foo" --type String --region us-east-1
$ aws ssm put-parameter --name /test/env2 --value "bar" --type String --region us-east-1
$ aws ssm add-tags-to-resource --resource-type Parameter --resource-id "/test/env1 resource id(n.b. invalid resource id)" --tags "Key=Env,Value=Production"

$ ./ssmenv --paths /test/env1,/test/env2 --tags "Env=Production"
export ENV1="foo"

$ eval $(./ssmenv --paths /test/env1,/test/env2 --tags "Env=Production")
$ echo $ENV1
foo
```

```console
$ aws ssm put-parameter --name /test/env1 --value "foo" --type String --region us-east-1
$ aws ssm put-parameter --name /test/env2 --value "bar" --type String --region us-east-1

$ ./ssmenv --paths /test
export ENV1="foo"
export ENV2="bar"

$ ./ssmenv --paths /test --without-export
ENV1="foo"
ENV2="bar"
```

```console
$ aws ssm put-parameter --name /test/env1 --value "foo" --type String --region us-east-1
$ aws ssm put-parameter --name /test/env2 --value "bar" --type String --region us-east-1

$ ./ssmenv --paths /test
export ENV1="foo"
export ENV2="bar"

$ export ENV2=""
$ ./ssmenv --paths /test --hide-exists
export ENV1="foo"

$ ./ssmenv --paths /test --fail-exists
2018/07/02 12:31:49 ENV2 already exists
$ echo $1
1
```

```console
$ aws ssm put-parameter --name /test/multienv --value "$(echo ENV3=bar; echo ENV4=baz)" --type String --region us-east-1

$ ./ssmenv --paths /test --multi-values /test/multienv
export ENV1="foo"
export ENV2="bar"
export ENV3="bar"
export ENV4="baz"

$ aws ssm put-parameter --name /test/secure_multienv --value "$(echo ENV5=foobar; echo ENV6=foobaz)" --type SecureString --key-id alias/aws/ssm --region us-east-1

$ ./ssmenv --paths /test --multi-values /test/multienv,/test/secure_multienv
export ENV1="foo"
export ENV2="bar"
export ENV3="bar"
export ENV4="baz"
export ENV5="foobar"
export ENV6="foobaz"

$ ./ssmenv --paths /test --multi-values /test/multienv,/test/secure_multienv --without-export
ENV1="foo"
ENV2="bar"
ENV3="bar"
ENV4="baz"
ENV5="foobar"
ENV6="foobaz"
```

# Known Issues

- `--paths` does not work absolute path

```console
$ aws ssm put-parameter --name /test/env1 --value "foo" --type String --region us-east-1
$ aws ssm put-parameter --name /test/env2 --value "bar" --type String --region us-east-1

# work
$ ./ssmenv --paths /test/
export ENV1="foo"
export ENV2="bar"

# does not work
$ ./ssmenv --paths /test/env1
2018/07/02 13:37:45 InvalidParameter: 1 validation error(s) found.
- minimum field size of 1, GetParametersInput.Names.

# work
$ ./ssmenv --names /test/env1,/test/env2
export ENV1="foo"
export ENV2="bar"
```
