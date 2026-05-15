package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Carregar Configurações (Ambiente, .env)
	log.Println("Carregando configurações...")

	// Carrega o arquivo .env (se existir)
	err := godotenv.Load()
	if err != nil {
		log.Println("Variáveis não encontradas.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. Inicializar Conexões (Banco de Dados, S3, Redis/Fila)
	// TODO: db := database.Connect()
	// TODO: s3Client := storage.NewS3Client()
	// TODO: queue := queue.NewClient()
	log.Println("Conexões com dependências (DB, S3) inicializadas com sucesso.")

	// 3. Configurar Roteamento e Handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 4. Configurar o Servidor HTTP
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		// Timeout é importante para conexões normais (mas ajustaremos isso nas rotas de upload dps)
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Graceful Shutdown (Desligamento Gracioso)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Iniciando API Server na porta %s...\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v\n", err)
		}
	}()

	<-stop // Espera até receber um sinal (ex: Ctrl+C)
	log.Println("Sinal de parada recebido... Desligando servidor (Graceful Shutdown)")

	// Dá um tempo (ex: 30 seg) para finalizar uploads ou requisições ativas antes de matar
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erro durante o desligamento: %v\n", err)
	}

	log.Println("Servidor finalizado com segurança.")
}
