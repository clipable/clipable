package transcoder

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"sort"
	"strconv"
	"time"

	"webserver/config"
	"webserver/models"
	"webserver/services"

	"github.com/alitto/pond"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// A go package that implements a worker pool to process files in minio using ffmpeg into mpeg-dash format
// and stores the output in minio.
type transcoder struct {
	*services.Group
	cfg  *config.Config
	pool *pond.WorkerPool

	qualityPresets []Quality

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

	// Parse the quality presets
	for _, preset := range cfg.FFmpeg.QualityPresets {
		var q Quality
		if _, err := fmt.Sscanf(preset, "%dx%d-%d@%f", &q.Width, &q.Height, &q.Framerate, &q.Bitrate); err != nil {
			return nil, errors.Wrap(err, "failed to parse quality preset")
		}

		// Assert the width and height are 16:9 or very close to it
		if math.Abs(float64(q.Width)/float64(q.Height)-16.0/9.0) > 0.1 {
			return nil, fmt.Errorf("invalid aspect ratio for quality preset %s. should be 16:9", preset)
		}

		t.qualityPresets = append(t.qualityPresets, q)
	}

	// Assert we have at least one preset
	if len(t.qualityPresets) == 0 {
		return nil, fmt.Errorf("no quality presets defined")
	}

	// Sort the quality presets by bitrate
	sort.Slice(t.qualityPresets, func(i, j int) bool {
		return t.qualityPresets[i].Bitrate < t.qualityPresets[j].Bitrate
	})

	return t, nil
}

func (t *transcoder) Start() error {
	// Find all clips that are marked as processing while starting to resume their processing
	orphanedClips, err := t.Clips.FindMany(context.Background(), &models.User{ID: -1}, models.ClipWhere.Processing.EQ(true))

	if err != nil {
		return err
	}

	for _, clip := range orphanedClips {
		if err := t.Queue(context.Background(), clip); err != nil {
			return err
		}
	}

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

	if prog.currentFrame < 0 {
		return prog.currentFrame, true
	}

	// divide current frame by max frames to get progress percentage
	return int(math.Round(float64(prog.currentFrame) / float64(prog.maxFrames) * 100.0)), true
}

func (t *transcoder) ReportProgress(clipID int64, currentFrame int) {
	prog, ok := t.progress.Get(clipID)

	if !ok {
		return
	}

	if prog.currentFrame == -2 {
		// If the clip is marked as failed, don't update the progress
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

	log.Infoln("Transcoding video", clip.ID)
	success := false

	defer func() {
		// If the clip is not marked as successful when we return, mark it as failed (-2) and then remove it from the progress map after 1 minute
		// This should give ample time for the client to call progress and notice the failure
		if !success {
			t.ReportProgress(clip.ID, -2)

			if err := t.Clips.Delete(ctx, clip); err != nil {
				log.WithError(err).Error("Failed to delete clip")
			}

			go func() {
				// Sleep for 5 minutes before marking the clip as failed
				time.Sleep(1 * time.Minute)
				t.progress.Remove(clip.ID)
			}()
		}
	}()

	rawURL := fmt.Sprintf("http://127.0.0.1:12786/s3/%d/raw", clip.ID)

	cmd := exec.Command("ffmpeg",
		"-i", rawURL,
		"-vf", `scale='if(gt(dar,1280/720),720*dar,1280)':'if(gt(dar,1280/720),720,1280/dar)',setsar=1,crop=1280:720`, // Scale to 1280 width, then crop image height to 720
		"-frames:v", "1",
		fmt.Sprintf("http://127.0.0.1:12786/s3/%d/thumbnail.jpg", clip.ID),
	)

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.WithError(err).
			WithField("output", string(output)).
			WithField("command", cmd.String()).
			Error("Failed to create thumbnail for video, we'd appreciate it if you'd report this issue to us on GitHub with a sample clip that causes the issue: https://github.com/clipable/clipable/issues/new")
		return
	}

	width, height, fps, duration, audioStreams, err := GetVideoStats(rawURL)

	if err != nil {
		log.WithError(err).
			Error("Error getting video stats")
		return
	}

	log.Infoln("Width", width, "Height", height, "FPS", fps, "Duration", duration, "AudioStreams", audioStreams)
	start := time.Now()

	ffmpegArgs := []string{
		"-i", rawURL,
		"-preset", t.cfg.FFmpeg.Preset,
		"-tune", t.cfg.FFmpeg.Tune,
		"-keyint_min", strconv.Itoa(fps),
		"-threads", strconv.Itoa(t.cfg.FFmpeg.Threads),
		"-hls_playlist_type", "vod",
		"-g", strconv.Itoa(fps),
		"-seg_duration", "2",
		"-sc_threshold", "0",
		"-c:v", t.cfg.FFmpeg.Codec,
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "1",
		"-ar", "96000",
		"-use_template", "1",
		"-use_timeline", "1",
		"-single_file", "1",
		"-x264opts", "no-scenecut",
		"-streaming", "0",
		"-movflags", "+faststart+dash+global_sidx",
		"-global_sidx", "1",
		"-utc_timing_url", "https://time.akamai.com/?iso",
		"-progress", fmt.Sprintf("http://127.0.0.1:12786/progress/%d", clip.ID),
	}

	ffmpegArgs = append(ffmpegArgs, t.GetPresets(width, height, fps, audioStreams)...)

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

	output, err = cmd.CombinedOutput()

	if err != nil {
		log.WithError(err).
			WithField("output", string(output)).
			WithField("args", ffmpegArgs).
			Error("Failed to transcode video, we'd appreciate it if you'd report this issue to us on GitHub with a sample clip that causes the issue: https://github.com/clipable/clipable/issues/new")
		return
	}

	log.Infoln("Finished transcoding video", clip.ID, "in", time.Since(start))

	if err := t.ObjectStore.DeleteObject(ctx, clip.ID, "raw"); err != nil {
		log.WithError(err).
			Error("Error deleting raw video")
		return
	}

	// Wait until all uploads are flushed and available in S3
	for t.ObjectStore.HasActiveUploads(ctx, clip.ID) {
		time.Sleep(500 * time.Millisecond)
	}

	clip.Processing = false

	if err := t.Clips.Update(ctx, clip, boil.Whitelist(models.ClipColumns.Processing)); err != nil {
		log.WithError(err).
			Error("Error updating clip")
		return
	}

	t.progress.Remove(clip.ID)

	success = true
}
