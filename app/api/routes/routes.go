package routes

import (
	"net/http"

	"github.com/MOliveiraDev/go-upload-files/internal/handlers"
	"github.com/MOliveiraDev/go-upload-files/internal/middleware"
)

// SetupAuthRoutes registra as rotas de autenticacao publica
func SetupAuthRoutes(mux *http.ServeMux, authHandler *handlers.AuthHandler) {
	mux.Handle("POST /auth/register", middleware.WrapErrorHandler(authHandler.Register))
	mux.Handle("POST /auth/login", middleware.WrapErrorHandler(authHandler.Login))
}

// SetupFileRoutes registra todas as rotas relacionadas a arquivos
func SetupFileRoutes(mux *http.ServeMux, fileHandler *handlers.FileHandler) {
	// Padrão do próprio Go a partir da versão 1.22: "MÉTODO /caminho"
	mux.Handle("POST /folders/{folderId}/files", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.UploadFile)))

	mux.Handle("GET /files", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.ListFiles)))
	mux.Handle("GET /folders/{folderId}/files", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.ListFilesInFolder)))

	mux.Handle("PATCH /files/{fileId}/name", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.RenameFile)))
	mux.Handle("PATCH /files/{fileId}/metadata", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.EditMetadata)))
	mux.Handle("PATCH /files/{fileId}/folder", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.MoveFile)))

	mux.Handle("DELETE /files", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.DeleteFiles)))
}

// SetupFolderRoutes registra todas as rotas relacionadas a pastas
func SetupFolderRoutes(mux *http.ServeMux, folderHandler *handlers.FolderHandler) {
	mux.Handle("POST /folders", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.CreateRootFolder)))
	mux.Handle("POST /folders/{folderId}/children", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.CreateSubfolder)))
	mux.Handle("POST /folders/{folderId}/copy", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.CopyFolder)))

	mux.Handle("GET /folders", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.ListFolders)))
	mux.Handle("GET /folders/{folderId}", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.GetFolder)))
	mux.Handle("GET /folders/{folderId}/children", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.ListSubfolders)))
	mux.Handle("GET /folders/{folderId}/path", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.GetFolderPath)))
	mux.Handle("GET /folders/{folderId}/items", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.GetFolderItems)))

	mux.Handle("PATCH /folders/{folderId}/name", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.RenameFolder)))
	mux.Handle("PATCH /folders/{folderId}/metadata", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.EditFolderMetadata)))
	mux.Handle("PATCH /folders/{folderId}/parent", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.MoveFolder)))

	mux.Handle("DELETE /folders/{folderId}", middleware.AuthMiddleware(http.HandlerFunc(folderHandler.DeleteFolders)))
}
