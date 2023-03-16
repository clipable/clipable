package transcoder

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// Resolution is a struct that contains the width and height of a video
type Quality struct {
	Width     int
	Height    int
	Bitrate   float32
	Framerate int
}

type VideoInfo struct {
	Streams []StreamInfo `json:"streams"`
	Format  FormatInfo   `json:"format"`
}

type FormatInfo struct {
	Duration string `json:"duration"`
}

type SideData struct {
	Rotation int `json:"rotation"`
}

type StreamInfo struct {
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Index        int        `json:"index"`
	CodecType    string     `json:"codec_type"`
	RFrameRate   string     `json:"r_frame_rate"`
	SideDataList []SideData `json:"side_data_list"`
}

func bitString(bitrate float32) string {
	return strconv.FormatFloat(float64(bitrate), 'f', 1, 64) + "M"
}

func GetVideoStats(file string) (width int, height int, fps int, duration time.Duration, audioStreams int, err error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration:stream=width,height,r_frame_rate,index,codec_type:stream_side_data=rotation", "-sexagesimal", "-of", "json", file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, 0, 0, 0, errors.Wrap(err, fmt.Sprintf("ffprobe failed: %s", out))
	}

	var info VideoInfo

	err = json.Unmarshal(out, &info)

	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	videoStream, ok := lo.Find(info.Streams, func(s StreamInfo) bool { return s.CodecType == "video" })

	if !ok {
		return 0, 0, 0, 0, 0, errors.New("no video stream found")
	}

	fps, err = strconv.Atoi(strings.Split(videoStream.RFrameRate, "/")[0])

	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	audioStreams = lo.CountBy(info.Streams, func(s StreamInfo) bool { return s.CodecType == "audio" })

	dur, err := ParseSexagesimal(info.Format.Duration)

	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	// If the video is rotated 90 or 270 degrees swap the width and height
	if len(videoStream.SideDataList) > 0 {
		rotation := math.Abs(float64(videoStream.SideDataList[0].Rotation))

		if rotation == 90 || rotation == 270 {
			videoStream.Width, videoStream.Height = videoStream.Height, videoStream.Width
		}
	}

	return videoStream.Width, videoStream.Height, fps, dur, audioStreams, nil
}

func ParseSexagesimal(duration string) (time.Duration, error) {
	parts := strings.Split(strings.TrimSpace(duration), ":")

	if len(parts) != 3 {
		return 0, errors.New("invalid duration format")
	}

	hours, err := strconv.Atoi(strings.TrimSpace(parts[0]))

	if err != nil {
		return 0, errors.Wrap(err, "failed to parse hours")
	}

	minutes, err := strconv.Atoi(strings.TrimSpace(parts[1]))

	if err != nil {
		return 0, errors.Wrap(err, "failed to parse minutes")
	}

	seconds, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)

	if err != nil {
		return 0, errors.Wrap(err, "failed to parse seconds")
	}

	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds*float64(time.Second)), nil
}

func (t *transcoder) GetPresets(width int, height int, fps int, audioStreams int) []string {
	if fps < 30 {
		fps = 30
	}

	vertical := height > width
	aspectRatio := "16:9"

	if vertical {
		aspectRatio = "9:16"
		width, height = height, width
	}

	var presets []Quality
	for _, preset := range t.qualityPresets {
		if preset.Width <= width && preset.Height <= height && preset.Framerate <= fps {
			targetPreset := preset
			if vertical {
				targetPreset.Width, targetPreset.Height = preset.Height, preset.Width
			}
			presets = append(presets, targetPreset)
		}
	}

	if len(presets) == 0 {
		presets = []Quality{t.qualityPresets[0]} // If no quality was select use the lowest one
	}

	ffmpegArgs := []string{"-aspect", aspectRatio}

	// ffmpeg -hwaccel nvdec -hwaccel_output_format cuda -i http://127.0.0.1:12786/s3/36/raw 
	// -filter_complex "
	// [0:v]hwupload_cuda, scale_npp=w=360:h=640:interp_algo=lanczos [v0];
	// [0:v]hwupload_cuda, scale_npp=w=480:h=854:interp_algo=lanczos [v1];
	// [0:v]hwupload_cuda, scale_npp=w=720:h=1280:interp_algo=lanczos [v2];
	// [0:v]hwupload_cuda, scale_npp=w=1080:h=1920:interp_algo=lanczos [v3];
	// [0:v]hwupload_cuda, scale_npp=w=1080:h=1920:interp_algo=lanczos [v4]" -preset medium -tune film -keyint_min 60 -threads 0 -hls_playlist_type vod -g 60 -seg_duration 2 -sc_threshold 0 -c:v h264_nvenc -pix_fmt yuv420p -c:a aac -b:a 128k -ac 1 -ar 96000 -use_template 1 -use_timeline 1 -single_file 1 -x264opts no-scenecut -streaming 0 -movflags +faststart+dash+global_sidx -global_sidx 1 -utc_timing_url https://time.akamai.com/?iso -progress http://127.0.0.1:12786/progress/36 -aspect 9:16 -map "[v0]" -b:v:0 1.0M -maxrate:0 1.2M -bufsize:0 2.0M -r:v:0 30 -map "[v1]" -b:v:1 


	for i, preset := range presets {
		ffmpegArgs = append(ffmpegArgs,
			"-map",
			"v:0",
			"-s:v:"+strconv.Itoa(i),
			fmt.Sprintf("%dx%d", preset.Width, preset.Height),
			"-b:v:"+strconv.Itoa(i),
			bitString(preset.Bitrate),
			"-maxrate:"+strconv.Itoa(i),
			bitString(preset.Bitrate*1.2),
			"-bufsize:"+strconv.Itoa(i),
			bitString(preset.Bitrate*2),
			"-r:v:"+strconv.Itoa(i),
			strconv.Itoa(preset.Framerate),
		)
	}

	return ffmpegArgs
}
