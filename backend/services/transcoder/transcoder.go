package transcoder

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"time"

	"webserver/config"
	"webserver/models"
	"webserver/services"

	"github.com/alitto/pond"
	cmap "github.com/orcaman/concurrent-map/v2"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// A go package that implements a worker pool to process files in minio using ffmpeg into mpeg-dash format
// and stores the output in minio.
type transcoder struct {
	*services.Group
	cfg  *config.Config
	pool *pond.WorkerPool

	progress cmap.ConcurrentMap[int64, *clipProgress]
	started  bool
}

type clipProgress struct {
	maxFrames    int
	currentFrame int
}

func New(cfg *config.Config, grp *services.Group) (services.Transcoder, error) {
	t := &transcoder{
		pool:  pond.New(cfg.FFmpeg.Concurrency, 1000),
		cfg:   cfg,
		Group: grp,
		progress: cmap.NewWithCustomShardingFunction[int64, *clipProgress](func(key int64) uint32 {
			// Copilot recommended this i have no idea if its correct
			return uint32(key % 10)
		}),
		started: false,
	}

	// Find all clips that are marked as processing while starting to resume their processing
	orphanedClips, err := t.Clips.FindMany(context.Background(), models.ClipWhere.Processing.EQ(true))

	if err != nil {
		return nil, err
	}

	for _, clip := range orphanedClips {
		if err := t.Queue(context.Background(), clip); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (t *transcoder) Start() error {
	if t.started {
		return fmt.Errorf("transcoder already started")
	}

	t.started = true

	return nil
}

func (t *transcoder) Queue(ctx context.Context, clip *models.Clip) error {
	t.progress.Set(clip.ID, &clipProgress{
		maxFrames:    0,
		currentFrame: -1,
	})

	t.pool.Submit(func() {
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

func (t *transcoder) GetProgress(clipID int64) (int, bool) {
	prog, ok := t.progress.Get(clipID)

	if !ok {
		return 0, false
	}

	if prog.currentFrame == -1 {
		return -1, true
	}

	// divide current frame by max frames to get progress percentage
	return int(math.Round(float64(prog.currentFrame) / float64(prog.maxFrames) * 100.0)), true
}

func (t *transcoder) ReportProgress(clipID int64, currentFrame int) {
	prog, ok := t.progress.Get(clipID)

	if !ok {
		return
	}

	prog.currentFrame = currentFrame
}

func (t *transcoder) process(ctx context.Context, clip *models.Clip) {
	// Maybe just use https://stackoverflow.com/questions/53352348/mpeg-dash-output-generated-by-ffmpeg-not-working ?
	// Example of variables in ffmpeg https://ottverse.com/hls-packaging-using-ffmpeg-live-vod/
	// Example of using ffmpeg map to pipes https://stackoverflow.com/questions/71041370/separate-video-from-audio-from-ffmpeg-stream
	// syscall pipe: https://www.codeflict.com/go/syscall-pipe
	// https://support.google.com/youtube/answer/1722171?hl=en#zippy=%2Cbitrate

	// Wait until we're started listen for requests before attempting to launch ffmpeg
	for !t.started {
		time.Sleep(1 * time.Second)
	}

	log.Infoln("Transcoding video", clip.ID)

	rawURL := fmt.Sprintf("http://127.0.0.1:12786/s3/%d/raw", clip.ID)

	cmd := exec.Command("ffmpeg",
		"-i", rawURL,
		"-ss", "00:00:01",
		"-s", "1280x720",
		"-qscale:v", "5",
		"-frames:v", "1",
		fmt.Sprintf("http://127.0.0.1:12786/s3/%d/thumbnail.jpg", clip.ID),
	)

	_, err := cmd.CombinedOutput()

	if err != nil {
		log.WithError(err).
			Error(cmd.String())
		return
	}

	width, height, fps, duration, audioStreams, err := GetVideoStats(rawURL)

	if err != nil {
		log.WithError(err).
			Error("Error getting video stats")
		return
	}

	fmt.Println("Width", width, "Height", height, "FPS", fps, "Duration", duration, "AudioStreams", audioStreams)
	start := time.Now()

	ffmpegArgs := []string{
		"-i", rawURL,
		"-preset", t.cfg.FFmpeg.Preset,
		"-tune", t.cfg.FFmpeg.Tune,
		"-keyint_min", strconv.Itoa(fps),
		"-hls_playlist_type", "vod",
		"-g", strconv.Itoa(fps),
		"-seg_duration", "2",
		"-sc_threshold", "0",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p", // TODO: overwrite pix_fmt parameters
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "1",
		"-ar", "96000",
		"-use_template", "1",
		"-use_timeline", "1",
		"-single_file", "1",
		"-x264opts", "no-scenecut",
		"-streaming", "0",
		"-movflags", "frag_keyframe+empty_moov",
		"-utc_timing_url", "https://time.akamai.com/?iso",
		"-progress", fmt.Sprintf("http://127.0.0.1:12786/progress/%d", clip.ID),
	}

	ffmpegArgs = append(ffmpegArgs, GetPresets(width, height, fps, audioStreams)...)

	if audioStreams > 0 {
		ffmpegArgs = append(ffmpegArgs, "-map", "0:a")
	}

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
		fmt.Sprintf("http://127.0.0.1:12786/s3/%d/dash.mpd", clip.ID),
	)

	cmd = exec.Command("ffmpeg", ffmpegArgs...)

	prog, ok := t.progress.Get(clip.ID)

	if !ok {
		log.WithField("clip", clip.ID).
			Error("Error getting progress")
		return
	}

	prog.maxFrames = int(math.Round(duration.Seconds() * 30))

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.WithError(err).
			WithField("output", string(output)).
			Error(cmd.String())
		return
	}

	log.Infoln("Finished transcoding video", clip.ID, "in", time.Since(start))

	if err := t.ObjectStore.DeleteObject(ctx, fmt.Sprintf("%d/raw", clip.ID)); err != nil {
		log.WithError(err).
			Error("Error deleting raw video")
		return
	}

	clip.Processing = false

	if err := t.Clips.Update(ctx, clip, boil.Whitelist(models.ClipColumns.Processing)); err != nil {
		log.WithError(err).
			Error("Error updating clip")
		return
	}
}
