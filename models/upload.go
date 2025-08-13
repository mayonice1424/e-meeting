package models

type UploadResponse struct {
	Message string `json:"message"`
	Data    struct {
		ImageURL string `json:"imageURL"`
	} `json:"data"`
}