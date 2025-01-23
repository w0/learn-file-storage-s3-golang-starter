package main

import (
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	http.MaxBytesReader(w, r.Body, 1<<30)
	defer r.Body.Close()

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Id", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT", err)
		return
	}

	dbVideo, err := cfg.db.GetVideo(videoID)
	if userId != dbVideo.UserID {
		respondWithError(w, http.StatusUnauthorized, "Not video owner", nil)
		return
	}

	video, header, err := r.FormFile("video")
	defer video.Close()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "video upload failed", err)
		return
	}

	contentType := header.Header.Get("Content-Type")

	mediaType, _, err := mime.ParseMediaType(contentType)
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "invalid video format", nil)
		return
	}

	tmp, err := os.CreateTemp("", "tubely-upload.mp4")
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create temp file", err)
		return
	}

	_, err = io.Copy(tmp, video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to save video", err)
		return
	}

	ratio, err := getVideoAspectRatio(tmp.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get aspect ratio", err)
		return
	}

	tmp.Seek(0, io.SeekStart)

	procVideo, err := processVideoForFastStart(tmp.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "video process error", err)
		return
	}

	videoFile, err := os.Open(procVideo)
	defer videoFile.Close()
	defer os.Remove(videoFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to open processed video", err)
		return
	}

	fileKey := getAssetName(mediaType)

	fileKey = addRatioPrefix(fileKey, ratio)

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &fileKey,
		Body:        videoFile,
		ContentType: &mediaType,
	}, func(o *s3.Options) {
		o.Region = cfg.s3Region
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "s3 upload failed", err)
	}

	videoUrl := cfg.getS3Url(fileKey)
	dbVideo.VideoURL = &videoUrl
	cfg.db.UpdateVideo(dbVideo)

}
