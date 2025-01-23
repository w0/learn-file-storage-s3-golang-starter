package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetName(contentType string) string {
	extension := getFileType(contentType)
	// help prevent caching of updated assets by giving each image a random filename
	bSlice := make([]byte, 32)
	rand.Read(bSlice)
	return fmt.Sprintf("%s%s", base64.RawURLEncoding.EncodeToString(bSlice), extension)
}

func getFileType(contentType string) string {
	split := strings.Split(contentType, "/")
	if len(split) != 2 {
		return ".bin"
	}
	return "." + split[1]
}

func (cfg *apiConfig) testAssetDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func (cfg *apiConfig) getAssetPath(filename string) string {
	return filepath.Join(cfg.assetsRoot, filename)
}

func (cfg *apiConfig) getAssetUrl(filename string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, filename)
}

func (cfg *apiConfig) getS3Url(filename string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, filename)
}

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
