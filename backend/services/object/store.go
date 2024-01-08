package object

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"sync/atomic"
	"webserver/config"
	"webserver/services"

	. "github.com/docker/go-units"
	"github.com/friendsofgo/errors"
	"github.com/minio/minio-go/v7"
	cmap "github.com/orcaman/concurrent-map/v2"
	log "github.com/sirupsen/logrus"
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

	defer s.DeleteObjects(ctx, cid, "raw")

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
			break
		}

		nid := fmt.Sprintf("%s-%d", objectPath, i)

		parts[i] = minio.CopySrcOptions{
			Bucket: s.cfg.S3.Bucket,
			Object: nid,
			Start:  0,
			End:    int64(n),
		}

		if _, err = s.s3.PutObject(ctx, s.cfg.S3.Bucket, nid, bytes.NewReader(buffer[:n]), int64(n), minio.PutObjectOptions{}); err != nil {
			return 0, err
		}
	}

	final, err := s.s3.ComposeObject(ctx, minio.CopyDestOptions{
		Bucket: s.cfg.S3.Bucket,
		Object: objectPath,
	}, parts...)

	if err != nil {
		return 0, err
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

func (s *store) DeleteObjects(ctx context.Context, cid int64, path string) error {
	objectsCh := make(chan minio.ObjectInfo)

	// Send object names that are needed to be removed to objectsCh
	// Todo: make sure we aren't leaking go-routines using this terrible pattern
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		opts := minio.ListObjectsOptions{Prefix: strconv.FormatInt(cid, 10) + path, Recursive: true}
		for object := range s.s3.ListObjects(context.Background(), s.cfg.S3.Bucket, opts) {
			if object.Err != nil {
				log.WithError(object.Err).Error("failed to list object")
			}
			objectsCh <- object
		}
	}()

	// Call RemoveObjects API
	errorCh := s.s3.RemoveObjects(context.Background(), s.cfg.S3.Bucket, objectsCh, minio.RemoveObjectsOptions{})

	// Print errors received from RemoveObjects API
	for e := range errorCh {
		return errors.Wrap(e.Err, "failed to delete object")
	}

	return nil
}

func (s *store) HasObject(ctx context.Context, cid int64, filename string) bool {
	_, err := s.s3.StatObject(ctx, s.cfg.S3.Bucket, fmt.Sprintf("%d/%s", cid, filename), minio.GetObjectOptions{})
	return err == nil
}
