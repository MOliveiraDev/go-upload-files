package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MOliveiraDev/go-upload-files/api/routes"
	"github.com/MOliveiraDev/go-upload-files/internal/database"
	"github.com/MOliveiraDev/go-upload-files/internal/handlers"
	"github.com/MOliveiraDev/go-upload-files/internal/services"
	"github.com/MOliveiraDev/go-upload-files/internal/storage/aws"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Carregando configurações...")

	err := godotenv.Load()
	if err != nil {
		log.Println("Variáveis não encontradas.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("PORT não definida, usando padrão: %s\n", port)
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Falha ao inicializar o banco de dados: %v", err)
	}
	defer db.Close() 

	storageClient, err := aws.NewS3Storage()
	if err != nil {
		log.Fatalf("Falha ao inicializar AWS S3: %v", err)
	}

	fileService := services.NewFileService(storageClient, nil)
	fileHandler := handlers.NewFileHandler(fileService)
	folderHandler := handlers.NewFolderHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	routes.SetupFileRoutes(mux, fileHandler)
	routes.SetupFolderRoutes(mux, folderHandler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Iniciando API Server na porta %s...\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v\n", err)
		}
	}()

	<-stop
	log.Println("Sinal de parada recebido... Desligando servidor (Graceful Shutdown)")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erro durante o desligamento: %v\n", err)
	}

	log.Println("Servidor finalizado com segurança.")
}
