package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MOliveiraDev/go-upload-files/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Carregando configurações...")

	// Carrega as variáveis de ambiente do arquivo .env (se existir)
	err := godotenv.Load()
	if err != nil {
		log.Println("Variáveis não encontradas.")
	}

	// Resgata a porta do servidor (padrão: 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("PORT não definida, usando padrão: %s\n", port)
	}

	// Inicializar Conexões (Banco de Dados)
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Falha ao inicializar o banco de dados: %v", err)
	}
	defer db.Close() // Fecha o banco suavemente quando a main() acabar

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Configurações do servidor HTTP
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Função assincrona para iniciar o servidor, permitindo que a main() continue e escute sinais de parada
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
