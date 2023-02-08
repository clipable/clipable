package transcoder

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"

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

func (t *transcoder) process(ctx context.Context, clip *models.Clip) {
	// Maybe just use https://stackoverflow.com/questions/53352348/mpeg-dash-output-generated-by-ffmpeg-not-working ?
	// Example of variables in ffmpeg https://ottverse.com/hls-packaging-using-ffmpeg-live-vod/
	// Example of using ffmpeg map to pipes https://stackoverflow.com/questions/71041370/separate-video-from-audio-from-ffmpeg-stream
	// syscall pipe: https://www.codeflict.com/go/syscall-pipe
	// https://support.google.com/youtube/answer/1722171?hl=en#zippy=%2Cbitrate

	log.Infoln("Transcoding video", clip.ID)

	cmd := exec.Command("ffmpeg",
		"-i", "http://127.0.0.1:12786/s3/"+clip.ID+"/raw",
		"-ss", "00:00:01",
		"-s", "1280x720",
		"-qscale:v", "5",
		"-frames:v", "1",
		"http://127.0.0.1:12786/s3/"+clip.ID+"/thumbnail.jpg",
	)

	_, err := cmd.CombinedOutput()

	if err != nil {
		log.WithError(err).
			Error(cmd.String())
		return
	}

	audioStreams, err := CountAudioStreams("http://127.0.0.1:12786/s3/" + clip.ID + "/raw")

	if err != nil {
		log.WithError(err).
			Error("Error counting audio streams")
		return
	}

	ffmpegArgs := []string{
		"-i", "http://127.0.0.1:12786/s3/" + clip.ID + "/raw",
		"-preset", "veryslow",
		"-keyint_min", "30",
		"-hls_playlist_type", "vod",
		"-g", "30",
		"-sc_threshold", "0",
		"-seg_duration", "1",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "1",
		"-ar", "96000",
		"-use_template", "1",
		"-use_timeline", "1",
		"-single_file", "1",
		"-tune", "film",
		"-x264opts", "no-scenecut",
		"-utc_timing_url", "https://time.akamai.com/?iso",
	}

	ffmpegArgs = append(ffmpegArgs, GetPresetsForVideo("http://127.0.0.1:12786/s3/"+clip.ID+"/raw")...)

	if audioStreams > 0 {
		ffmpegArgs = append(ffmpegArgs, "-map", "0:a")
	}

	fmt.Println(audioStreams)

	if audioStreams > 1 {
		ffmpegArgs = append(
			ffmpegArgs,
			"-filter_complex",
			"amerge=inputs="+strconv.Itoa(audioStreams),
		)
	}

	if audioStreams > 0 {
		ffmpegArgs = append(ffmpegArgs, "-adaptation_sets", "id=0,streams=v id=1,streams=a")
	} else {
		ffmpegArgs = append(ffmpegArgs, "-adaptation_sets", "id=0,streams=v")
	}

	ffmpegArgs = append(ffmpegArgs,
		"-f", "dash",
		"http://127.0.0.1:12786/s3/"+clip.ID+"/dash.mpd",
	)

	cmd = exec.Command("ffmpeg", ffmpegArgs...)

	_, err = cmd.CombinedOutput()

	if err != nil {
		log.WithError(err).
			Error(cmd.String())
		return
	}

	log.Infoln("Finished transcoding video", clip.ID)

	if err := t.obj.DeleteObject(ctx, clip.ID+"/raw"); err != nil {
		log.WithError(err).
			Error("Error deleting raw video")
		return
	}
}
