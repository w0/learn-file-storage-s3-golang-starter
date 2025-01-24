package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
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

func (cfg *apiConfig) createBucketKey(filename string) string {
	return fmt.Sprintf("%s,%s", cfg.s3Bucket, filename)
}
