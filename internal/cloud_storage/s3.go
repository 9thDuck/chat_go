package cloudStorage

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Presigner struct {
	Presigner *s3.PresignClient
}

func (p Presigner) Get(
	ctx context.Context, bucketName string, objectKey string, lifetimeSecs int64) (string, error) {
	request, err := p.Presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		return "", err
	}
	return request.URL, err
}

func (presigner Presigner) Create(
	ctx context.Context, bucketName string, objectKey string, lifetimeSecs int64) (string, error) {
	request, err := presigner.Presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		return "", err
	}
	return request.URL, nil
}

func NewS3Presigner(s3PresignClient *s3.PresignClient) *Presigner {
	return &Presigner{
		Presigner: s3PresignClient,
	}
}

func NewS3Client(cfg *aws.Config) *s3.Client {
	s3Client := s3.NewFromConfig(*cfg, func(o *s3.Options) {
		o.Region = cfg.Region
	})
	return s3Client
}

func NewAWSCredentialsProvider(accessKeyID string, secretAccessKey string) aws.CredentialsProvider {
	return aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}, nil
	})
}

func NewAWSConfig(region string, credentialsProvider aws.CredentialsProvider) *aws.Config {
	return &aws.Config{
		Region:      region,
		Credentials: credentialsProvider,
	}
}

func NewS3CloudStorage(cfg *aws.Config) *CloudStorage {
	s3Client := NewS3Client(cfg)
	s3PresignerClient := s3.NewPresignClient(s3Client)
	s3Presigner := NewS3Presigner(s3PresignerClient)
	return &CloudStorage{
		PreSigner: s3Presigner,
	}
}
