package tokenstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// SSMTokenStore implements Storage. Tokens are found in SSM Params at /<SSM_PREFIX>/<email_at_emailhost>
type SSMTokenStore struct {
	client *ssm.Client
	prefix string
}

func NewSSMTokenStore(ctx context.Context, prefix string) (*SSMTokenStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}
	return &SSMTokenStore{
		client: ssm.NewFromConfig(cfg),
		prefix: prefix,
	}, nil
}

// paramName maps an email to a full SSM parameter name. SSM parameter names
// can't contain "@", so it's substituted with "_at_" here and reversed in
// emailFromParamName — safe as long as no email's local part literally
// contains "_at_".
func (s *SSMTokenStore) paramName(email string) string {
	return path.Join(s.prefix, strings.ReplaceAll(email, "@", "_at_"))
}

func emailFromParamName(prefix, name string) string {
	trimmed := strings.TrimPrefix(name, prefix+"/")
	return strings.ReplaceAll(trimmed, "_at_", "@")
}

func (s *SSMTokenStore) StoreToken(name string, token interface{}) error {
	b, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshaling token for %s: %w", name, err)
	}
	_, err = s.client.PutParameter(context.Background(), &ssm.PutParameterInput{
		Name:      aws.String(s.paramName(name)),
		Value:     aws.String(string(b)),
		Type:      ssmtypes.ParameterTypeSecureString,
		Overwrite: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("putting ssm parameter for %s: %w", name, err)
	}
	return nil
}

func (s *SSMTokenStore) RetrieveToken(name string, token interface{}) error {
	out, err := s.client.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           aws.String(s.paramName(name)),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("getting ssm parameter for %s: %w", name, err)
	}
	return json.Unmarshal([]byte(aws.ToString(out.Parameter.Value)), &token)
}

func (s *SSMTokenStore) ListEmails() ([]string, error) {
	var emails []string
	var nextToken *string
	for {
		out, err := s.client.GetParametersByPath(context.Background(), &ssm.GetParametersByPathInput{
			Path:      aws.String(s.prefix),
			Recursive: aws.Bool(true),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("listing ssm parameters under %s: %w", s.prefix, err)
		}
		for _, p := range out.Parameters {
			emails = append(emails, emailFromParamName(s.prefix, aws.ToString(p.Name)))
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	return emails, nil
}
