package agentfilestorage

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3AgentFileStore implements the Storage interface and expects data in S3 at /<S3_BUCKET>/<email>/AGENT.md
type S3AgentFileStore struct {
	client *s3.Client
	bucket string
}

func NewS3AgentFileStorage(ctx context.Context, bucket string) (*S3AgentFileStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}
	return &S3AgentFileStore{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

func (s *S3AgentFileStore) RetrieveFile(name string) (string, error) {
	key := path.Join(name, filename)

	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("getting s3://%s/%s: %w", s.bucket, key, err)
	}
	defer out.Body.Close()

	b, err := io.ReadAll(out.Body)
	if err != nil {
		return "", fmt.Errorf("reading s3://%s/%s: %w", s.bucket, key, err)
	}
	return string(b), nil
}
