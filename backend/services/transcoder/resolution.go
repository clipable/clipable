package transcoder

import (
	"encoding/json"
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
}

type StreamInfo struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Index      int    `json:"index"`
	CodecType  string `json:"codec_type"`
	RFrameRate string `json:"r_frame_rate"`
	Duration   string `json:"duration"`
}

// https://support.google.com/youtube/answer/1722171?hl=en#zippy=%2Cbitrate
var QualityPresets = []Quality{
	// 360p
	{Width: 640, Height: 360, Bitrate: 1, Framerate: 30},
	// 480p
	{Width: 854, Height: 480, Bitrate: 2.5, Framerate: 30},
	// 720p
	{Width: 1280, Height: 720, Bitrate: 5, Framerate: 30},
	// 1080p
	{Width: 1920, Height: 1080, Bitrate: 8, Framerate: 30},
	{Width: 1920, Height: 1080, Bitrate: 12, Framerate: 60},
	// 1440p
	{Width: 2560, Height: 1440, Bitrate: 16, Framerate: 30},
	{Width: 2560, Height: 1440, Bitrate: 24, Framerate: 60},
	// 2160p
	{Width: 3840, Height: 2160, Bitrate: 45, Framerate: 30},
	{Width: 3840, Height: 2160, Bitrate: 68, Framerate: 60},
	// 4320p
	{Width: 7680, Height: 4320, Bitrate: 160, Framerate: 30},
	{Width: 7680, Height: 4320, Bitrate: 240, Framerate: 60},
}

func bitString(bitrate float32) string {
	return strconv.FormatFloat(float64(bitrate), 'f', 1, 64) + "M"
}

func GetVideoStats(file string) (width int, height int, fps int, duration time.Duration, audioStreams int, err error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=width,height,r_frame_rate,index,codec_type,duration", "-sexagesimal", "-of", "json", file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, 0, 0, 0, err
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

	dur, err := ParseSexagesimal(videoStream.Duration)

	if err != nil {
		return 0, 0, 0, 0, 0, err
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

func GetPresets(width int, height int, fps int, audioStreams int) []string {
	if fps < 30 {
		fps = 30
	}

	var presets []Quality
	for _, preset := range QualityPresets {
		if preset.Width <= width && preset.Height <= height && preset.Framerate <= fps {
			presets = append(presets, preset)
		}
	}

	// TODO: What happens when someone uploads vertical video?
	if len(presets) == 0 {
		presets = []Quality{QualityPresets[0]} // If no quality was select use the lowest one
	}

	ffmpegArgs := []string{}

	for i, preset := range presets {
		ffmpegArgs = append(ffmpegArgs,
			"-map",
			"v:0",
			"-s:"+strconv.Itoa(i),
			strconv.Itoa(preset.Width)+"x"+strconv.Itoa(preset.Height),
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
