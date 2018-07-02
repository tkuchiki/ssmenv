package ssmenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/tkuchiki/aws-sdk-go-config"
)

type SSMEnv struct {
	ssmClient *ssm.SSM
	replacer  *strings.Replacer
	envs      map[string]struct{}
}

func newAWSSession(accessKey, secretKey, arn, token, region, profile, config, creds string) (*session.Session, error) {
	conf := awsconfig.Option{
		Arn:         arn,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Region:      region,
		Token:       token,
		Profile:     profile,
		Config:      config,
		Credentials: creds,
	}

	return awsconfig.NewSession(conf)
}

func NewSSMEnv(accessKey, secretKey, arn, token, region, profile, config, creds string) (*SSMEnv, error) {
	sess, err := newAWSSession(accessKey, secretKey, arn, token, region, profile, config, creds)

	if err != nil {
		return &SSMEnv{}, err
	}

	envs := make(map[string]struct{})

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		envs[pair[0]] = struct{}{}
	}

	return &SSMEnv{
		ssmClient: ssm.New(sess),
		replacer:  strings.NewReplacer("-", "/", ".", "/"),
		envs:      envs,
	}, nil
}

func (se *SSMEnv) describeParametersFilterByNames(names, types []string) ([]*ssm.ParameterMetadata, error) {
	strFilters := []*ssm.ParameterStringFilter{
		&ssm.ParameterStringFilter{
			Key:    aws.String("Name"),
			Values: aws.StringSlice(names),
		},
	}

	strFilters = append(strFilters,
		&ssm.ParameterStringFilter{
			Key:    aws.String("Type"),
			Values: aws.StringSlice(types),
		},
	)

	return se.describeParameters(strFilters)
}

func (se *SSMEnv) describeParametersFilterByPaths(paths, types []string, recursive bool) ([]*ssm.ParameterMetadata, error) {
	strFilter := &ssm.ParameterStringFilter{
		Key:    aws.String("Path"),
		Option: aws.String("OneLevel"),
		Values: aws.StringSlice(paths),
	}
	if recursive {
		strFilter.SetOption("Recursive")
	}

	strFilters := []*ssm.ParameterStringFilter{
		strFilter,
		&ssm.ParameterStringFilter{
			Key:    aws.String("Type"),
			Values: aws.StringSlice(types),
		},
	}

	return se.describeParameters(strFilters)
}

func (se *SSMEnv) describeParametersFilterByTags(tags, types []string) ([]*ssm.ParameterMetadata, error) {
	strFilters := make([]*ssm.ParameterStringFilter, 0, len(tags)+len(types))

	for _, tag := range tags {
		splited := strings.Split(tag, "=")
		if len(splited) != 2 {
			return []*ssm.ParameterMetadata{}, fmt.Errorf("Invalid tag format: %s", tag)
		}
		values := strings.Split(splited[1], ",")
		filter := &ssm.ParameterStringFilter{
			Key:    aws.String(fmt.Sprintf("tag:%s", splited[0])),
			Values: aws.StringSlice(values),
		}
		strFilters = append(strFilters, filter)
	}

	strFilters = append(strFilters,
		&ssm.ParameterStringFilter{
			Key:    aws.String("Type"),
			Values: aws.StringSlice(types),
		},
	)

	return se.describeParameters(strFilters)
}

func (se *SSMEnv) describeParameters(strFilters []*ssm.ParameterStringFilter) ([]*ssm.ParameterMetadata, error) {

	var nextToken string
	metadata := make([]*ssm.ParameterMetadata, 0)
	input := &ssm.DescribeParametersInput{}

	if len(strFilters) > 0 {
		input.SetParameterFilters(strFilters)
	}

	for {
		if nextToken != "" {
			input.SetNextToken(nextToken)
		}

		out, err := se.ssmClient.DescribeParameters(input)
		if err != nil {
			return []*ssm.ParameterMetadata{}, err
		}

		metadata = append(metadata, out.Parameters...)

		if out.NextToken == nil {
			break
		}
		nextToken = *out.NextToken
	}

	return metadata, nil
}

func (se *SSMEnv) DescribeParametersFilterByPaths(paths, tags, types []string, recursive bool) ([]*ssm.ParameterMetadata, error) {
	metadata := make([]*ssm.ParameterMetadata, 0)

	if len(tags) > 0 {
		md, err := se.describeParametersFilterByTags(tags, types)
		if err != nil {
			return []*ssm.ParameterMetadata{}, err
		}
		metadata = append(metadata, md...)
	}

	if len(paths) > 0 {
		md, err := se.describeParametersFilterByPaths(paths, types, recursive)
		if err != nil {
			return []*ssm.ParameterMetadata{}, err
		}
		metadata = append(metadata, md...)
	}

	return metadata, nil
}

func (se *SSMEnv) DescribeParametersFilterByNames(names, tags, types []string) ([]*ssm.ParameterMetadata, error) {
	metadata := make([]*ssm.ParameterMetadata, 0)

	if len(tags) > 0 {
		md, err := se.describeParametersFilterByTags(tags, types)
		if err != nil {
			return []*ssm.ParameterMetadata{}, err
		}
		metadata = append(metadata, md...)
	}

	if len(names) > 0 {
		md, err := se.describeParametersFilterByNames(names, types)
		if err != nil {
			return []*ssm.ParameterMetadata{}, err
		}
		metadata = append(metadata, md...)
	}

	return metadata, nil
}

func (se *SSMEnv) GetParametersByPath(path string, recursive bool) ([]*ssm.GetParametersByPathOutput, error) {
	var nextToken string
	outs := make([]*ssm.GetParametersByPathOutput, 0)
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
		Recursive:      aws.Bool(recursive),
	}

	for {

		if nextToken != "" {
			input.SetNextToken(nextToken)
		}

		out, err := se.ssmClient.GetParametersByPath(input)
		if err != nil {
			return []*ssm.GetParametersByPathOutput{}, err
		}

		outs = append(outs, out)

		if out.NextToken == nil {
			break
		}
		nextToken = *out.NextToken
	}

	return outs, nil
}

func (se *SSMEnv) GetParameters(names []string) (*ssm.GetParametersOutput, error) {
	input := &ssm.GetParametersInput{
		Names:          aws.StringSlice(names),
		WithDecryption: aws.Bool(true),
	}

	return se.ssmClient.GetParameters(input)
}

func (se *SSMEnv) SliceHasPrefix(values []string, key string) bool {
	for _, v := range values {
		if strings.HasPrefix(v, key) {
			return true
		}
	}

	return false
}

func (se *SSMEnv) SliceContains(values []string, key string) bool {
	for _, v := range values {
		if v == key {
			return true
		}
	}

	return false
}

func (se *SSMEnv) GetParamNames(parameters []*ssm.ParameterMetadata) []string {
	names := make([]string, 0, len(parameters))

	for _, p := range parameters {
		names = append(names, *(p.Name))
	}

	return names
}

func (se *SSMEnv) GetSplitedName(name string) string {
	splitNames := strings.Split(se.replacer.Replace(name), "/")
	return splitNames[len(splitNames)-1]
}

func (se *SSMEnv) EnvIsExists(name string) bool {
	_, ok := se.envs[name]

	return ok
}
