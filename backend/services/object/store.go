package object

import (
	"context"
	"io"
	"webserver/services"

	"github.com/minio/minio-go/v7"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type store struct {
	s3         *minio.Client
	bucketName string
	contexts   cmap.ConcurrentMap[string, context.CancelFunc]
}

func NewStore(c *minio.Client, bucketName string) services.ObjectStore {
	return &store{c, bucketName, cmap.New[context.CancelFunc]()}
}

func (s *store) PutObject(ctx context.Context, id string, r io.Reader, size int64) (int64, error) {
	if cancel, ok := s.contexts.Get(id); ok {
		cancel()
	}

	ctx, cancel := context.WithCancel(ctx)

	s.contexts.Set(id, cancel)

	inf, err := s.s3.PutObject(ctx, s.bucketName, id, r, size, minio.PutObjectOptions{})

	return inf.Size, err
}

func (s *store) GetObject(ctx context.Context, id string) (io.ReadSeekCloser, int64, error) {
	obj, err := s.s3.GetObject(ctx, s.bucketName, id, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, err
	}

	inf, err := obj.Stat()
	if err != nil {
		return nil, 0, err
	}

	return obj, inf.Size, err
}

func (s *store) DeleteObject(ctx context.Context, id string) error {
	return s.s3.RemoveObject(ctx, s.bucketName, id, minio.RemoveObjectOptions{})
}

func (s *store) HasObject(ctx context.Context, id string) bool {
	_, err := s.s3.StatObject(ctx, s.bucketName, id, minio.GetObjectOptions{})
	return err == nil
}
