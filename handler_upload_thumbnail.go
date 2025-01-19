package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload
	const maxMemory = 10 << 20

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "thumbnail upload failed", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	defer file.Close()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to parse formfile", err)
		return
	}

	contentType := header.Header.Get("Content-Type")

	if contentType == "" {
		respondWithError(w, http.StatusBadRequest, "content-type not set", nil)
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error reading file", err)
		return
	}

	dbVideo, err := cfg.db.GetVideo(videoID)

	if userID != dbVideo.UserID {
		respondWithError(w, http.StatusUnauthorized, "not authorized to modify video", nil)
		return
	}

	thumbUrl := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(imageData))

	dbVideo.ThumbnailURL = &thumbUrl

	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "thumbnail url creation failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideo)
}
