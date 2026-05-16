package routes

import (
	"net/http"

	"github.com/MOliveiraDev/go-upload-files/internal/handlers"
)

// SetupFileRoutes registra todas as rotas relacionadas a arquivos
func SetupFileRoutes(mux *http.ServeMux, fileHandler *handlers.FileHandler) {
	// Padrão do próprio Go a partir da versão 1.22: "MÉTODO /caminho"
	mux.HandleFunc("POST /folders/{folderId}/files", fileHandler.UploadFile)

	mux.HandleFunc("GET /files", fileHandler.ListFiles)
	mux.HandleFunc("GET /folders/{folderId}/files", fileHandler.ListFilesInFolder)

	mux.HandleFunc("PATCH /files/{fileId}/name", fileHandler.RenameFile)
	mux.HandleFunc("PATCH /files/{fileId}/metadata", fileHandler.EditMetadata)
	mux.HandleFunc("PATCH /files/{fileId}/folder", fileHandler.MoveFile)

	mux.HandleFunc("DELETE /files", fileHandler.DeleteFiles)
}

// SetupFolderRoutes registra todas as rotas relacionadas a pastas
func SetupFolderRoutes(mux *http.ServeMux, folderHandler *handlers.FolderHandler) {
	mux.HandleFunc("POST /folders", folderHandler.CreateRootFolder)
	mux.HandleFunc("POST /folders/{folderId}/children", folderHandler.CreateSubfolder)
	mux.HandleFunc("POST /folders/{folderId}/copy", folderHandler.CopyFolder)

	mux.HandleFunc("GET /folders", folderHandler.ListFolders)
	mux.HandleFunc("GET /folders/{folderId}", folderHandler.GetFolder)
	mux.HandleFunc("GET /folders/{folderId}/children", folderHandler.ListSubfolders)
	mux.HandleFunc("GET /folders/{folderId}/path", folderHandler.GetFolderPath)
	mux.HandleFunc("GET /folders/{folderId}/items", folderHandler.GetFolderItems)

	mux.HandleFunc("PATCH /folders/{folderId}/name", folderHandler.RenameFolder)
	mux.HandleFunc("PATCH /folders/{folderId}/metadata", folderHandler.EditFolderMetadata)
	mux.HandleFunc("PATCH /folders/{folderId}/parent", folderHandler.MoveFolder)

	mux.HandleFunc("DELETE /folders/{folderId}", folderHandler.DeleteFolders)
}
