package dto

import (
	"mime/multipart"
)

// File
type FileFormRequest struct {
	File *multipart.FileHeader `json:"file" form:"file" binding:"required" swaggerignoer:"true"`
}

type UploadFileRequest struct {
	FileFormRequest
	Description string `json:"description" form:"description" binding:"required"`
}

type CreateFileRequest struct {
	Name        string `json:"name"`
	Directory   string `json:"directory"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

type UpdateFileRequest struct {
	Description string `json:"description"`
}

type FileResponse struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Directory   string `json:"directory"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// HttpRequestConfig represents the configuration for each request
type HttpRequestConfig struct {
	Method       string
	URL          string
	Body         any
	ResponseType any
	RequestID    string
}

// Response represents the structure of the HTTP response
type HttpResponseWrapper struct {
	StatusCode   int
	Body         any
	Err          error
	ResponseType any
	RequestID    string
}

type HttpResponseResult[T any] struct {
	Result T `json:"result"`
}
