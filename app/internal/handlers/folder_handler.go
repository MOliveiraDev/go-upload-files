package handlers

import "net/http"

type FolderHandler struct {
}

func NewFolderHandler() *FolderHandler {
	return &FolderHandler{}
}

// POST /folders
func (h *FolderHandler) CreateRootFolder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Não implementado: Criar pasta raiz"))
}

// POST /folders/{folderId}/children
func (h *FolderHandler) CreateSubfolder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// POST /folders/{folderId}/copy
func (h *FolderHandler) CopyFolder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GET /folders (e /folders?name={name})
func (h *FolderHandler) ListFolders(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GET /folders/{folderId}
func (h *FolderHandler) GetFolder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GET /folders/{folderId}/children
func (h *FolderHandler) ListSubfolders(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GET /folders/{folderId}/path
func (h *FolderHandler) GetFolderPath(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GET /folders/{folderId}/items
func (h *FolderHandler) GetFolderItems(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// PATCH /folders/{folderId}/name
func (h *FolderHandler) RenameFolder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// PATCH /folders/{folderId}/metadata
func (h *FolderHandler) EditFolderMetadata(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// PATCH /folders/{folderId}/parent
func (h *FolderHandler) MoveFolder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// DELETE /folders/{folderId} e DELETE /folders
func (h *FolderHandler) DeleteFolders(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
