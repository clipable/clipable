package transcoder

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"webserver/models"
	"webserver/services"

	log "github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc/pool"
)

// A go package that implements a worker pool to process files in minio using ffmpeg into mpeg-dash format
// and stores the output in minio.
type transcoder struct {
	pool *pool.Pool
	obj  services.ObjectStore
}

func New(obj services.ObjectStore, workers int) *transcoder {
	return &transcoder{
		pool: pool.New().WithMaxGoroutines(workers),
		obj:  obj,
	}
}

func (t *transcoder) Queue(ctx context.Context, clip *models.Clip) error {
	t.pool.Go(func() {
		defer func() {
			if v := recover(); v != nil {
				log.WithField("clip", clip.ID).
					WithField("panic", v).
					Error("Panic in transcoder")
			}
		}()
		t.process(ctx, clip)
	})

	return nil
}

func (t *transcoder) uploadFile(id string, file string, obj string) error {
	f, err := os.Open(id + "/" + file)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = t.obj.PutObject(id+"/"+obj, f, -1)

	return err
}

func (t *transcoder) process(ctx context.Context, clip *models.Clip) {
	// Maybe just use https://stackoverflow.com/questions/53352348/mpeg-dash-output-generated-by-ffmpeg-not-working ?
	// Example of variables in ffmpeg https://ottverse.com/hls-packaging-using-ffmpeg-live-vod/
	// Example of using ffmpeg map to pipes https://stackoverflow.com/questions/71041370/separate-video-from-audio-from-ffmpeg-stream
	// syscall pipe: https://www.codeflict.com/go/syscall-pipe
	// https://support.google.com/youtube/answer/1722171?hl=en#zippy=%2Cbitrate

	if err := os.Mkdir(clip.ID, 0755); err != nil {
		log.WithError(err).
			Error("Error creating directory for clip")
		return
	}

	defer os.RemoveAll(clip.ID + "/")

	log.Infoln("Transcoding video", clip.ID)

	cmd := exec.Command("ffmpeg",
		"-i", "http://127.0.0.1:12786/read/"+clip.ID+"/raw",
		"-preset", "slow",
		"-keyint_min", "30",
		"-g", "30",
		"-sc_threshold", "0",
		"-seg_duration", "1",
		"-r", "30",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "1",
		"-ar", "96000",
		"-use_template", "1",
		"-use_timeline", "1",
		"-single_file", "1",
		"-map", "v:0", "-s:0", "640x360", "-b:v:0", "1M", "-maxrate:0", "1.2M", "-bufsize:0", "2M",
		"-map", "v:0", "-s:1", "854x480", "-b:v:1", "2.5M", "-maxrate:1", "2.7M", "-bufsize:1", "5M",
		"-map", "v:0", "-s:2", "1280x720", "-b:v:2", "5M", "-maxrate:2", "5.3M", "-bufsize:2", "10M",
		"-map", "v:0", "-s:3", "1920x1080", "-b:v:3", "10M", "-maxrate:3", "10.6M", "-bufsize:3", "20M",
		"-map", "0:a",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		"-f", "dash",
		clip.ID+"/manifest.mpd",
	)

	_, err := cmd.CombinedOutput()

	if err != nil {
		log.WithError(err).
			Error(cmd.String())
		return
	}

	if err := t.uploadFile(clip.ID, "manifest.mpd", "manifest"); err != nil {
		log.WithError(err).
			Error("Error uploading manifest to minio")
		return
	}

	filepath.Walk(clip.ID, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".mpd") {
			return nil
		}

		if err := t.uploadFile(clip.ID, info.Name(), info.Name()); err != nil {
			log.WithError(err).
				Error("Error uploading file to minio")
			return err
		}

		return nil
	})

	log.Infoln("Finished transcoding video", clip.ID)

	if err := t.obj.DeleteObject(clip.ID + "/raw"); err != nil {
		log.WithError(err).
			Error("Error deleting raw video")
		return
	}
}
