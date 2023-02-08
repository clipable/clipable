package transcoder

import (
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Resolution is a struct that contains the width and height of a video
type Quality struct {
	Width     int
	Height    int
	Bitrate   float32
	Framerate int
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

func getVideoStats(file string) (int, int, int, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height,r_frame_rate", "-of", "csv=s=x:p=0", file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, 0, err
	}

	resolution := strings.Split(string(out), "x")
	width, err := strconv.Atoi(resolution[0])
	if err != nil {
		return 0, 0, 0, err
	}

	height, err := strconv.Atoi(resolution[1])
	if err != nil {
		return 0, 0, 0, err
	}

	fps, err := strconv.Atoi(strings.Split(resolution[2], "/")[0])

	if err != nil {
		return 0, 0, 0, err
	}

	return width, height, fps, nil
}

func CountAudioStreams(file string) (int, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "a", "-show_entries", "stream=index", "-of", "csv=s=x:p=0", file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	return len(strings.Split(string(out), "\n")) - 1, nil
}

func GetPresetsForVideo(file string) []string {
	width, height, fps, err := getVideoStats(file)

	if err != nil {
		log.Error(err)
		return []string{}
	}

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
