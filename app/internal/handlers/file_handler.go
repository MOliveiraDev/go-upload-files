package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/MOliveiraDev/go-upload-files/internal/dto"
	"github.com/MOliveiraDev/go-upload-files/internal/middleware"
	"github.com/MOliveiraDev/go-upload-files/internal/services"
	"github.com/google/uuid"
)

type FileHandler struct {
	fileService *services.FileService
}

type renameFileRequest struct {
	Name string `json:"name"`
}

type editFileMetadataRequest struct {
	ContentType string `json:"contentType"`
}

type moveFileRequest struct {
	FolderID *string `json:"folderId"`
}

type deleteFilesRequest struct {
	FileIDs []string `json:"fileIds"`
}

func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (h *FileHandler) InitUpload(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var req dto.InitUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}

	upload, err := h.fileService.InitUpload(r.Context(), userID, req)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusCreated, map[string]any{"data": upload})
	return nil
}

func (h *FileHandler) UploadPart(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	uploadID, err := parseUUIDPathValue(r, "uploadId")
	if err != nil {
		return err
	}

	partNumber, err := parseInt32PathValue(r, "partNumber")
	if err != nil {
		return err
	}

	chunk, err := io.ReadAll(r.Body)
	if err != nil {
		return middleware.NewBadRequestError("falha ao ler corpo da parte", err)
	}
	if len(chunk) == 0 {
		return middleware.NewBadRequestError("o corpo da parte não pode estar vazio", errors.New("parte vazia"))
	}

	etag, upload, err := h.fileService.UploadPart(r.Context(), userID, uploadID, partNumber, chunk)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"uploadId":   upload.ID,
			"partNumber": partNumber,
			"etag":       etag,
		},
	})
	return nil
}

func (h *FileHandler) CompleteUpload(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	uploadID, err := parseUUIDPathValue(r, "uploadId")
	if err != nil {
		return err
	}

	var req dto.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}
	if err := req.ValidateRequest(); err != nil {
		return middleware.NewBadRequestError("dados de conclusão do upload inválidos", err)
	}

	file, err := h.fileService.CompleteUpload(r.Context(), userID, uploadID, req.Parts)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusCreated, map[string]any{"data": file})
	return nil
}

func (h *FileHandler) AbortUpload(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	uploadID, err := parseUUIDPathValue(r, "uploadId")
	if err != nil {
		return err
	}

	if err := h.fileService.AbortUpload(r.Context(), userID, uploadID); err != nil {
		return mapFileServiceError(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	filter, err := parseFileFilter(r)
	if err != nil {
		return err
	}

	files, err := h.fileService.ListFiles(r.Context(), userID, filter)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": files})
	return nil
}

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	fileID, err := parseUUIDPathValue(r, "fileId")
	if err != nil {
		return err
	}

	file, err := h.fileService.GetFile(r.Context(), userID, fileID)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": file})
	return nil
}

func (h *FileHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	fileID, err := parseUUIDPathValue(r, "fileId")
	if err != nil {
		return err
	}

	downloadURL, file, err := h.fileService.GetDownloadURL(r.Context(), userID, fileID)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"file":        file,
			"downloadUrl": downloadURL,
		},
	})
	return nil
}

func (h *FileHandler) ListFilesInFolder(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folderID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	files, err := h.fileService.ListFilesInFolder(r.Context(), userID, folderID)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": files})
	return nil
}

func (h *FileHandler) RenameFile(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	fileID, err := parseUUIDPathValue(r, "fileId")
	if err != nil {
		return err
	}

	var req renameFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}

	file, err := h.fileService.RenameFile(r.Context(), userID, fileID, req.Name)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": file})
	return nil
}

func (h *FileHandler) EditMetadata(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	fileID, err := parseUUIDPathValue(r, "fileId")
	if err != nil {
		return err
	}

	var req editFileMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}

	file, err := h.fileService.EditMetadata(r.Context(), userID, fileID, req.ContentType)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": file})
	return nil
}

func (h *FileHandler) MoveFile(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	fileID, err := parseUUIDPathValue(r, "fileId")
	if err != nil {
		return err
	}

	var req moveFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}

	folderID, err := parseOptionalUUID(req.FolderID)
	if err != nil {
		return middleware.NewBadRequestError("folderId inválido", err)
	}

	file, err := h.fileService.MoveFile(r.Context(), userID, fileID, folderID)
	if err != nil {
		return mapFileServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": file})
	return nil
}

func (h *FileHandler) DeleteFiles(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var req deleteFilesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError("corpo da requisição inválido", err)
	}

	fileIDs, err := parseUUIDList(req.FileIDs)
	if err != nil {
		return middleware.NewBadRequestError("lista de fileIds inválida", err)
	}

	if err := h.fileService.DeleteFiles(r.Context(), userID, fileIDs); err != nil {
		return mapFileServiceError(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func mapFileServiceError(err error) error {
	switch {
	case errors.Is(err, services.ErrFileNotFound):
		return middleware.NewNotFoundError("arquivo não encontrado", err)
	case errors.Is(err, services.ErrFileForbidden):
		return middleware.NewForbiddenError("o arquivo não pertence ao usuário autenticado", err)
	case errors.Is(err, services.ErrFileNameAlreadyExists):
		return middleware.NewConflictError("já existe um arquivo com esse nome no destino", err)
	case errors.Is(err, services.ErrInvalidFileName):
		return middleware.NewBadRequestError(err.Error(), err)
	case errors.Is(err, services.ErrFolderNotFound):
		return middleware.NewNotFoundError("pasta não encontrada", err)
	case errors.Is(err, services.ErrFolderForbidden):
		return middleware.NewForbiddenError("a pasta não pertence ao usuário autenticado", err)
	case errors.Is(err, services.ErrUploadNotFound):
		return middleware.NewNotFoundError("upload não encontrado", err)
	case errors.Is(err, services.ErrUploadForbidden):
		return middleware.NewForbiddenError("o upload não pertence ao usuário autenticado", err)
	case errors.Is(err, services.ErrUploadAlreadyClosed):
		return middleware.NewConflictError("o upload já foi concluído ou abortado", err)
	default:
		return middleware.NewInternalError("falha ao processar requisição de arquivo", err)
	}
}

func parseFileFilter(r *http.Request) (services.FileFilter, error) {
	query := r.URL.Query()
	filter := services.FileFilter{
		Name:   query.Get("name"),
		Status: query.Get("status"),
	}

	uploadedFrom, err := parseOptionalRFC3339Query(query.Get("uploadedFrom"), "uploadedFrom inválido; use RFC3339")
	if err != nil {
		return services.FileFilter{}, err
	}
	filter.UploadedFrom = uploadedFrom

	uploadedTo, err := parseOptionalRFC3339Query(query.Get("uploadedTo"), "uploadedTo inválido; use RFC3339")
	if err != nil {
		return services.FileFilter{}, err
	}
	filter.UploadedTo = uploadedTo

	folderID, err := parseOptionalUUIDQuery(query.Get("folderId"), "folderId inválido")
	if err != nil {
		return services.FileFilter{}, err
	}
	filter.FolderID = folderID

	page, err := parseOptionalPositiveIntQuery(query.Get("page"), "page inválido")
	if err != nil {
		return services.FileFilter{}, err
	}
	filter.Page = page

	pageSize, err := parseOptionalPositiveIntQuery(query.Get("pageSize"), "pageSize inválido")
	if err != nil {
		return services.FileFilter{}, err
	}
	filter.PageSize = pageSize

	return filter, nil
}

func parseOptionalRFC3339Query(value, errorMessage string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, middleware.NewBadRequestError(errorMessage, err)
	}

	return &parsedTime, nil
}

func parseOptionalUUIDQuery(value, errorMessage string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}

	parsedUUID, err := uuid.Parse(value)
	if err != nil {
		return nil, middleware.NewBadRequestError(errorMessage, err)
	}

	return &parsedUUID, nil
}

func parseOptionalPositiveIntQuery(value, errorMessage string) (int, error) {
	if value == "" {
		return 0, nil
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil || parsedValue < 1 {
		return 0, middleware.NewBadRequestError(errorMessage, err)
	}

	return parsedValue, nil
}

func parseUUIDList(values []string) ([]uuid.UUID, error) {
	fileIDs := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		parsedValue, err := uuid.Parse(value)
		if err != nil {
			return nil, err
		}
		fileIDs = append(fileIDs, parsedValue)
	}

	return fileIDs, nil
}

func parseInt32PathValue(r *http.Request, key string) (int32, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return 0, middleware.NewBadRequestError("parâmetro de rota ausente", errors.New(key+" é obrigatório"))
	}

	parsedValue, err := strconv.ParseInt(pathValue, 10, 32)
	if err != nil || parsedValue < 1 {
		return 0, middleware.NewBadRequestError("parâmetro numérico de rota inválido", err)
	}

	return int32(parsedValue), nil
}
