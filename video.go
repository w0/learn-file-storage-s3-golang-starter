package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

func getVideoAspectRatio(filepath string) (string, error) {
	eCmd := exec.Command("/usr/bin/ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)

	// cmd result sent to buffer
	buf := bytes.Buffer{}
	eCmd.Stdout = &buf
	err := eCmd.Run()
	if err != nil {
		return "", err
	}

	type ffprobe struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	probeJson := ffprobe{}

	err = json.Unmarshal(buf.Bytes(), &probeJson)
	if err != nil {
		return "", err
	}

	ratio := CalculateAspectRatio(probeJson.Streams[0].Width, probeJson.Streams[0].Height)

	return ratio, nil
}

func CalculateAspectRatio(width int, height int) string {
	ratio := float64(width) / float64(height)

	if ratio >= 1.7 && ratio <= 1.8 {
		return "16:9"
	} else if ratio >= 0.5 && ratio <= 0.6 {
		return "9:16"
	}
	return "other"
}

func addRatioPrefix(filename string, ratio string) string {
	if ratio == "16:9" {
		return "landscape/" + filename
	}

	if ratio == "9:16" {
		return "portrait/" + filename
	}

	return "other/" + filename
}

func processVideoForFastStart(filePath string) (string, error) {
	outFile := filePath + ".processing"

	cmd := exec.Command("/usr/bin/ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outFile)
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return outFile, nil
}
