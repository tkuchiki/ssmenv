package ssmenv

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/ssm"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type CSV []string

func (c *CSV) Set(value string) error {
	vals := strings.Split(value, ",")

	for _, v := range vals {
		*c = append(*c, strings.Trim(v, " "))
	}

	return nil
}

func (c *CSV) String() string {
	return ""
}

func ParseCSV(s kingpin.Settings) (target *[]string) {
	target = &[]string{}
	s.SetValue((*CSV)(target))
	return

}

func Run() error {
	var app = kingpin.New("ssmenv", "Expand environment variables from AWS EC2 Systems Manager Parameter Store")
	var names = ParseCSV(app.Flag("names", "Names of the parameters (comma separated)").PlaceHolder("PARAM_NAME,..."))
	var paths = ParseCSV(app.Flag("paths", "The hierarchy for the parameter names (comma separated)").PlaceHolder("PARAM_PATH,..."))
	var tags = ParseCSV(app.Flag("tags", "Filter by tags (comma separated)").PlaceHolder("KEY=VALUE,..."))
	var types = ParseCSV(app.Flag("types", "The type of parameters (comma separated)").Default("String,SecureString"))
	var multiValues = ParseCSV(app.Flag("multi-values", "Names or paths with multiple values (comma separated)").PlaceHolder("PARAM_NAME,..."))
	var nonRecursive = app.Flag("non-recursive", "Describes one level paths").Bool()
	var withoutExport = app.Flag("without-export", "Without export").Bool()
	var hideExists = app.Flag("hide-exists", "Hide environment variables if it already exists").Bool()
	var failExists = app.Flag("fail-exists", "Fail if environment variables alerady exists").Bool()
	var awsAccessKeyID = app.Flag("access-key", "The AWS access key ID").String()
	var awsSecretAccessKey = app.Flag("secret-key", "The AWS secret access key").String()
	var awsArn = app.Flag("assume-role-arn", "The AWS assume role ARN").String()
	var awsToken = app.Flag("token", "The AWS access token").String()
	var awsRegion = app.Flag("region", "The AWS region").String()
	var awsProfile = app.Flag("profile", "The AWS CLI profile").String()
	var awsConfig = app.Flag("aws-config", "The AWS CLI Config file").String()
	var awsCreds = app.Flag("credentials", "The AWS CLI Credential file").String()

	app.Version("0.1.1")

	kingpin.MustParse(app.Parse(os.Args[1:]))

	recursive := !*nonRecursive

	ssmenv, err := NewSSMEnv(*awsAccessKeyID, *awsSecretAccessKey, *awsArn, *awsToken, *awsRegion, *awsProfile, *awsConfig, *awsCreds)
	if err != nil {
		return err
	}

	metadata, err := ssmenv.DescribeParametersFilterByPaths(*paths, *tags, *types, recursive)
	if err != nil {
		return err
	}

	envs := make(map[string]string)

	var params *ssm.GetParametersOutput
	nameFilters := make([]string, 0, len(metadata))
	if len(*paths) > 0 {
		for _, m := range metadata {
			nameFilters = append(nameFilters, *(m.Name))

		}
	}

	metadata, err = ssmenv.DescribeParametersFilterByNames(*names, *tags, *types)
	if err != nil {
		return err
	}

	if len(*names) > 0 {
		for _, m := range metadata {
			nameFilters = append(nameFilters, *(m.Name))
		}
	}

	params, err = ssmenv.GetParameters(nameFilters)
	if err != nil {
		return err
	}

	rawValues := make([]string, 0, len(*multiValues))
LABEL:
	for _, p := range params.Parameters {
		for _, name := range *multiValues {
			if *(p.Name) == name {
				rawValues = append(rawValues, *(p.Value))
				continue LABEL
			}
		}

		envName := strings.ToUpper(ssmenv.GetSplitedName(*(p.Name)))
		if *failExists && ssmenv.EnvIsExists(envName) {
			return fmt.Errorf("%s already exists", envName)
		}

		if *hideExists && ssmenv.EnvIsExists(envName) {
			continue
		}

		envs[envName] = *(p.Value)
	}

	envNames := make([]string, 0)
	for ename, _ := range envs {
		envNames = append(envNames, ename)
	}

	for _, val := range rawValues {
		scanner := bufio.NewScanner(strings.NewReader(val))
		for scanner.Scan() {
			if *withoutExport {
				fmt.Println(scanner.Text())
			} else {
				fmt.Println(fmt.Sprintf("export %s", scanner.Text()))
			}
		}
	}

	sort.SliceStable(envNames, func(i, j int) bool { return envNames[i] < envNames[j] })
	for _, ename := range envNames {
		if *withoutExport {
			fmt.Println(fmt.Sprintf(`%s="%s"`, ename, envs[ename]))
		} else {
			fmt.Println(fmt.Sprintf(`export %s="%s"`, ename, envs[ename]))
		}
	}

	return nil
}
