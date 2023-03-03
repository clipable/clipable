package object

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"sync/atomic"
	"webserver/config"
	"webserver/services"

	. "github.com/docker/go-units"
	"github.com/minio/minio-go/v7"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type store struct {
	s3            *minio.Client
	cfg           *config.Config
	contexts      cmap.ConcurrentMap[string, context.CancelFunc]
	activeUploads cmap.ConcurrentMap[int64, *int64]
}

func NewStore(c *minio.Client, cfg *config.Config) services.ObjectStore {
	return &store{c, cfg, cmap.New[context.CancelFunc](), cmap.NewWithCustomShardingFunction[int64, *int64](func(key int64) uint32 {
		// Copilot recommended this i have no idea if its correct
		return uint32(key % 10)
	})}
}

func (s *store) HasActiveUploads(ctx context.Context, cid int64) bool {
	return s.activeUploads.Has(cid)
}

func (s *store) AddActiveUpload(ctx context.Context, cid int64) {
	if s.activeUploads.Has(cid) {
		count, _ := s.activeUploads.Get(cid)

		// Atomically increment and set the value
		atomic.AddInt64(count, 1)
		return
	}

	count := int64(1)

	s.activeUploads.Set(cid, &count)
}

func (s *store) RemoveActiveUpload(ctx context.Context, cid int64) {
	if !s.activeUploads.Has(cid) {
		return
	}

	count, _ := s.activeUploads.Get(cid)

	// Atomically decrement and set the value
	atomic.AddInt64(count, -1)

	if atomic.LoadInt64(count) == 0 {
		s.activeUploads.Remove(cid)
	}
}

func (s *store) PutObject(ctx context.Context, cid int64, filename string, r io.Reader) (int64, error) {
	objectPath := fmt.Sprintf("%d/%s", cid, filename)

	s.AddActiveUpload(ctx, cid)
	defer s.RemoveActiveUpload(ctx, cid)

	if cancel, ok := s.contexts.Get(objectPath); ok {
		cancel()
	}

	ctx, cancel := context.WithCancel(ctx)

	s.contexts.Set(objectPath, cancel)
	defer s.contexts.Remove(objectPath)

	buffer := make([]byte, 16*MB)
	parts := make([]minio.CopySrcOptions, int(math.Round(float64(s.cfg.MaxUploadSizeBytes)/float64(16*MB)))+1)
	deletionObjects := make([]minio.ObjectInfo, len(parts))

	for i := 0; i < len(parts); i++ {
		if ctx.Err() != nil {
			return 0, errors.Wrap(ctx.Err(), "context error")
		}
		n, err := io.ReadFull(r, buffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return 0, err
		}

		if n == 0 {
			parts = parts[:i]
			deletionObjects = deletionObjects[:i]
			break
		}

		nid := fmt.Sprintf("%s-%d", objectPath, i)

		parts[i] = minio.CopySrcOptions{
			Bucket: s.cfg.S3.Bucket,
			Object: nid,
			Start:  0,
			End:    int64(n),
		}

		upinf, err := s.s3.PutObject(ctx, s.cfg.S3.Bucket, nid, bytes.NewReader(buffer[:n]), int64(n), minio.PutObjectOptions{})
		if err != nil {
			return 0, err
		}

		deletionObjects[i] = minio.ObjectInfo{
			Key:       upinf.Key,
			VersionID: upinf.VersionID,
		}
	}

	final, err := s.s3.ComposeObject(ctx, minio.CopyDestOptions{
		Bucket: s.cfg.S3.Bucket,
		Object: objectPath,
	}, parts...)

	if err != nil {
		return 0, err
	}

	deletionQueue := make(chan minio.ObjectInfo)

	go func() {
		for _, obj := range deletionObjects {
			deletionQueue <- obj
		}

		close(deletionQueue)
	}()

	for r := range s.s3.RemoveObjects(ctx, s.cfg.S3.Bucket, deletionQueue, minio.RemoveObjectsOptions{}) {
		if r.Err != nil {
			return 0, r.Err
		}
	}

	return final.Size, err
}

func (s *store) GetObject(ctx context.Context, cid int64, filename string) (io.ReadSeekCloser, int64, string, error) {
	obj, err := s.s3.GetObject(ctx, s.cfg.S3.Bucket, fmt.Sprintf("%d/%s", cid, filename), minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, "", err
	}

	inf, err := obj.Stat()
	if err != nil {
		return nil, 0, "", err
	}

	return obj, inf.Size, inf.ETag, nil
}

func (s *store) DeleteObject(ctx context.Context, cid int64, filename string) error {
	return s.s3.RemoveObject(ctx, s.cfg.S3.Bucket, fmt.Sprintf("%d/%s", cid, filename), minio.RemoveObjectOptions{})
}

func (s *store) HasObject(ctx context.Context, cid int64, filename string) bool {
	_, err := s.s3.StatObject(ctx, s.cfg.S3.Bucket, fmt.Sprintf("%d/%s", cid, filename), minio.GetObjectOptions{})
	return err == nil
}
