package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MOliveiraDev/go-upload-files/internal/middleware"
	"github.com/MOliveiraDev/go-upload-files/internal/services"
	"github.com/google/uuid"
)

var ErrNewBadRequest = errors.New("Corpo da requisição inválido")

type FolderHandler struct {
	folderService *services.FolderService
}

type createFolderRequest struct {
	Name string `json:"name"`
}

type renameFolderRequest struct {
	Name string `json:"name"`
}

type moveFolderRequest struct {
	ParentID *string `json:"parentId"`
}

func NewFolderHandler(folderService *services.FolderService) *FolderHandler {
	return &FolderHandler{folderService: folderService}
}

func (h *FolderHandler) CreateRootFolder(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var req createFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}

	folder, err := h.folderService.CreateFolder(r.Context(), userID, req.Name, nil)
	if err != nil {
		return mapFolderServiceError(err)
	}

	writeJSONResponse(w, http.StatusCreated, map[string]any{"data": folder})
	return nil
}

func (h *FolderHandler) CreateSubfolder(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	parentID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	var req createFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}

	folder, err := h.folderService.CreateFolder(r.Context(), userID, req.Name, &parentID)
	if err != nil {
		return mapFolderServiceError(err)
	}

	writeJSONResponse(w, http.StatusCreated, map[string]any{"data": folder})
	return nil
}

func (h *FolderHandler) ListFolders(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folders, err := h.folderService.ListFolders(r.Context(), userID, r.URL.Query().Get("name"))
	if err != nil {
		return middleware.NewInternalError("falha ao listar pastas", err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": folders})
	return nil
}

func (h *FolderHandler) GetFolder(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folderID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	folder, err := h.folderService.GetFolder(r.Context(), userID, folderID)
	if err != nil {
		return mapFolderServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": folder})
	return nil
}

func (h *FolderHandler) ListSubfolders(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folderID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	folders, err := h.folderService.ListSubfolders(r.Context(), userID, folderID)
	if err != nil {
		return mapFolderServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": folders})
	return nil
}

func (h *FolderHandler) GetFolderPath(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folderID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	path, err := h.folderService.GetFolderPath(r.Context(), userID, folderID)
	if err != nil {
		return mapFolderServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": path})
	return nil
}

func (h *FolderHandler) GetFolderItems(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folderID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	items, err := h.folderService.GetFolderItems(r.Context(), userID, folderID)
	if err != nil {
		return mapFolderServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": items})
	return nil
}

func (h *FolderHandler) RenameFolder(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folderID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	var req renameFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}

	folder, err := h.folderService.RenameFolder(r.Context(), userID, folderID, req.Name)
	if err != nil {
		return mapFolderServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": folder})
	return nil
}

func (h *FolderHandler) MoveFolder(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folderID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	var req moveFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError(ErrNewBadRequest.Error(), err)
	}

	parentID, err := parseOptionalUUID(req.ParentID)
	if err != nil {
		return middleware.NewBadRequestError("parentId inválido", err)
	}

	folder, err := h.folderService.MoveFolder(r.Context(), userID, folderID, parentID)
	if err != nil {
		return mapFolderServiceError(err)
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"data": folder})
	return nil
}

func (h *FolderHandler) DeleteFolder(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	folderID, err := parseUUIDPathValue(r, "folderId")
	if err != nil {
		return err
	}

	if err := h.folderService.DeleteFolder(r.Context(), userID, folderID); err != nil {
		return mapFolderServiceError(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func authenticatedUserID(r *http.Request) (uuid.UUID, error) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, middleware.NewUnauthorizedError("token de autenticação obrigatório", errors.New("id do usuário não encontrado no contexto"))
	}

	return userID, nil
}

func parseUUIDPathValue(r *http.Request, key string) (uuid.UUID, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return uuid.Nil, middleware.NewBadRequestError("parâmetro de rota ausente", errors.New(key+" é obrigatório"))
	}

	parsedValue, err := uuid.Parse(pathValue)
	if err != nil {
		return uuid.Nil, middleware.NewBadRequestError("parâmetro UUID de rota inválido", err)
	}

	return parsedValue, nil
}

func parseOptionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	if *value == "" {
		return nil, nil
	}

	parsedValue, err := uuid.Parse(*value)
	if err != nil {
		return nil, err
	}

	return &parsedValue, nil
}

func mapFolderServiceError(err error) error {
	switch {
	case errors.Is(err, services.ErrFolderNotFound):
		return middleware.NewNotFoundError("pasta não encontrada", err)
	case errors.Is(err, services.ErrFolderForbidden):
		return middleware.NewForbiddenError("a pasta não pertence ao usuário autenticado", err)
	case errors.Is(err, services.ErrFolderNameAlreadyExists):
		return middleware.NewConflictError("já existe uma pasta com esse nome no destino", err)
	case errors.Is(err, services.ErrInvalidFolderName):
		return middleware.NewBadRequestError(err.Error(), err)
	case errors.Is(err, services.ErrInvalidFolderMove):
		return middleware.NewConflictError(err.Error(), err)
	default:
		return middleware.NewInternalError("falha ao processar requisição de pasta", err)
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
