package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}

	splits := strings.Split(*video.VideoURL, ",")

	if len(splits) < 2 {
		return video, nil
	}

	presignedUrl, err := generatePresignedURL(cfg.s3Client, splits[0], splits[1], time.Minute*5)
	if err != nil {
		return video, err
	}

	video.VideoURL = &presignedUrl

	return video, nil
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	preSign := s3.NewPresignClient(s3Client)

	req, err := preSign.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expireTime))

	if err != nil {
		return "", err
	}

	return req.URL, nil
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
