package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

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

	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "medietype parse failed", err)
		return
	}

	if mediatype != "image/png" && mediatype != "image/jpeg" {
		respondWithError(w, http.StatusBadRequest, "content not allowed", nil)
		return
	}

	dbVideo, err := cfg.db.GetVideo(videoID)

	if userID != dbVideo.UserID {
		respondWithError(w, http.StatusUnauthorized, "not authorized to modify video", nil)
		return
	}

	fileName := getAssetName(dbVideo.ID, contentType)
	assetPath := cfg.getAssetPath(fileName)

	f, err := os.Create(assetPath)
	defer f.Close()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "asset creation failed", err)
		return
	}

	_, err = io.Copy(f, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to save asset", err)
	}

	thumbnailUrl := cfg.getAssetUrl(fileName)
	dbVideo.ThumbnailURL = &thumbnailUrl

	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "thumbnail url creation failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideo)
}
