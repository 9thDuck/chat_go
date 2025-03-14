package cloudStorage

import (
	"context"
)

type PreSigner interface {
	Create(ctx context.Context, bucketName, key string, lifetimeInSeconds int64) (string, error)
	Get(ctx context.Context, bucketName string, objectKey string, lifetimeSecs int64) (string, error)
}
type CloudStorage struct {
	PreSigner PreSigner
}
