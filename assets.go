package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetName(videoId uuid.UUID, contentType string) string {
	extension := getFileType(contentType)
	return fmt.Sprintf("%s%s", videoId, extension)
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
