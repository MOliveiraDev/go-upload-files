package routes

import (
	"net/http"

	"github.com/MOliveiraDev/go-upload-files/internal/handlers"
	"github.com/MOliveiraDev/go-upload-files/internal/middleware"
)

// SetupAuthRoutes registra as rotas sem necessidade de autenticação
func SetupAuthRoutes(mux *http.ServeMux, authHandler *handlers.AuthHandler) {
	mux.Handle("POST /auth/register", middleware.WrapErrorHandler(authHandler.Register))
	mux.Handle("POST /auth/login", middleware.WrapErrorHandler(authHandler.Login))
	mux.Handle("POST /auth/logout", middleware.WrapErrorHandler(authHandler.Logout))
}

// SetupFileRoutes registra todas as rotas relacionadas a arquivos
func SetupFileRoutes(mux *http.ServeMux, fileHandler *handlers.FileHandler) {
	mux.Handle("POST /files", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.UploadFile)))
	mux.Handle("POST /folders/{folderId}/files", middleware.AuthMiddleware(http.HandlerFunc(fileHandler.UploadFile)))

	mux.Handle("GET /files", middleware.AuthMiddleware(middleware.WrapErrorHandler(fileHandler.ListFiles)))
	mux.Handle("GET /files/{fileId}", middleware.AuthMiddleware(middleware.WrapErrorHandler(fileHandler.GetFile)))
	mux.Handle("GET /files/{fileId}/download", middleware.AuthMiddleware(middleware.WrapErrorHandler(fileHandler.GetDownloadURL)))
	mux.Handle("GET /folders/{folderId}/files", middleware.AuthMiddleware(middleware.WrapErrorHandler(fileHandler.ListFilesInFolder)))

	mux.Handle("PATCH /files/{fileId}/name", middleware.AuthMiddleware(middleware.WrapErrorHandler(fileHandler.RenameFile)))
	mux.Handle("PATCH /files/{fileId}/metadata", middleware.AuthMiddleware(middleware.WrapErrorHandler(fileHandler.EditMetadata)))
	mux.Handle("PATCH /files/{fileId}/folder", middleware.AuthMiddleware(middleware.WrapErrorHandler(fileHandler.MoveFile)))

	mux.Handle("DELETE /files", middleware.AuthMiddleware(middleware.WrapErrorHandler(fileHandler.DeleteFiles)))
}

// SetupFolderRoutes registra todas as rotas relacionadas a pastas
func SetupFolderRoutes(mux *http.ServeMux, folderHandler *handlers.FolderHandler) {
	mux.Handle("POST /folders", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.CreateRootFolder)))
	mux.Handle("POST /folders/{folderId}/children", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.CreateSubfolder)))

	mux.Handle("GET /folders", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.ListFolders)))
	mux.Handle("GET /folders/{folderId}", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.GetFolder)))
	mux.Handle("GET /folders/{folderId}/children", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.ListSubfolders)))
	mux.Handle("GET /folders/{folderId}/path", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.GetFolderPath)))
	mux.Handle("GET /folders/{folderId}/items", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.GetFolderItems)))

	mux.Handle("PATCH /folders/{folderId}/name", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.RenameFolder)))
	mux.Handle("PATCH /folders/{folderId}/parent", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.MoveFolder)))

	mux.Handle("DELETE /folders/{folderId}", middleware.AuthMiddleware(middleware.WrapErrorHandler(folderHandler.DeleteFolder)))
}
