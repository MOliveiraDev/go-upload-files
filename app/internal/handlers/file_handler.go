package handlers

import (
	"net/http"

	"github.com/MOliveiraDev/go-upload-files/internal/services"
)

// FileHandler gerencia as requisições HTTP relacionadas aos arquivos
type FileHandler struct {
	fileService *services.FileService
}

func NewFileHandler(s *services.FileService) *FileHandler {
	return &FileHandler{fileService: s}
}

// POST /folders/{folderId}/files
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Vai chamar h.fileService.StartUpload(...)
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Não implementado: Iniciar upload do arquivo"))
}

// GET /files
// Opcionalmente lida com query params: ?name={name}, ?status=processing, ?uploadedFrom=...
func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Não implementado: Listar/Buscar arquivos"))
}

// GET /folders/{folderId}/files
func (h *FileHandler) ListFilesInFolder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Não implementado: Listar arquivos na pasta"))
}

// PATCH /files/{fileId}/name
func (h *FileHandler) RenameFile(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Não implementado: Renomear arquivo"))
}

// PATCH /files/{fileId}/metadata
func (h *FileHandler) EditMetadata(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Não implementado: Editar metadados do arquivo"))
}

// PATCH /files/{fileId}/folder
func (h *FileHandler) MoveFile(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Não implementado: Mover arquivo para outra pasta"))
}

// DELETE /files
func (h *FileHandler) DeleteFiles(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Não implementado: Deletar arquivo(s)"))
}
