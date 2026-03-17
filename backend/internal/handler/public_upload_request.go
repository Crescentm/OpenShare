package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/model"
	"openshare/backend/internal/service"
	"openshare/backend/internal/session"
)

var (
	errUploadFormParse        = errors.New("failed to parse upload form")
	errUploadManifestInvalid  = errors.New("invalid upload manifest")
	errUploadManifestMismatch = errors.New("upload files do not match manifest")
	errUploadFileRequired     = errors.New("file is required")
	errUploadFileRead         = errors.New("failed to read uploaded file")
)

type uploadManifestEntry struct {
	RelativePath string `json:"relative_path"`
}

type parsedPublicUploadRequest struct {
	input   service.PublicUploadInput
	closers []io.Closer
}

func (r *parsedPublicUploadRequest) Close() {
	for _, closer := range r.closers {
		_ = closer.Close()
	}
}

func (h *PublicUploadHandler) parseSubmissionRequest(ctx *gin.Context) (*parsedPublicUploadRequest, error) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, h.currentMaxRequestBytes(ctx.Request.Context()))

	form, err := ctx.MultipartForm()
	if err != nil {
		return nil, errUploadFormParse
	}

	fileHeaders, manifest, err := extractUploadFiles(form, ctx.PostForm("manifest"))
	if err != nil {
		return nil, err
	}

	files := make([]service.PublicUploadFileInput, 0, len(fileHeaders))
	closers := make([]io.Closer, 0, len(fileHeaders))
	for index, fileHeader := range fileHeaders {
		file, openErr := fileHeader.Open()
		if openErr != nil {
			closeClosers(closers)
			return nil, errUploadFileRead
		}
		closers = append(closers, file)
		files = append(files, service.PublicUploadFileInput{
			OriginalName: fileHeader.Filename,
			RelativePath: manifest[index].RelativePath,
			DeclaredMIME: fileHeader.Header.Get("Content-Type"),
			File:         file,
		})
	}

	return &parsedPublicUploadRequest{
		input: service.PublicUploadInput{
			Description:   ctx.PostForm("description"),
			ReceiptCode:   readPublicReceiptCode(ctx),
			FolderID:      ctx.PostForm("folder_id"),
			UploaderIP:    ctx.ClientIP(),
			DirectPublish: canDirectPublish(ctx),
			Files:         files,
		},
		closers: closers,
	}, nil
}

func extractUploadFiles(form *multipart.Form, manifestRaw string) ([]*multipart.FileHeader, []uploadManifestEntry, error) {
	if manifestRaw == "" {
		fileHeaders := form.File["file"]
		if len(fileHeaders) == 0 {
			return nil, nil, errUploadFileRequired
		}

		manifest := make([]uploadManifestEntry, 0, len(fileHeaders))
		for _, fileHeader := range fileHeaders {
			manifest = append(manifest, uploadManifestEntry{RelativePath: fileHeader.Filename})
		}
		return fileHeaders, manifest, nil
	}

	var manifest []uploadManifestEntry
	if err := json.Unmarshal([]byte(manifestRaw), &manifest); err != nil {
		return nil, nil, errUploadManifestInvalid
	}

	fileHeaders := form.File["files"]
	if len(fileHeaders) == 0 || len(fileHeaders) != len(manifest) {
		return nil, nil, errUploadManifestMismatch
	}
	return fileHeaders, manifest, nil
}

func canDirectPublish(ctx *gin.Context) bool {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		return false
	}
	return identity.HasPermission(model.AdminPermissionSubmissionModeration)
}

func closeClosers(closers []io.Closer) {
	for _, closer := range closers {
		_ = closer.Close()
	}
}
