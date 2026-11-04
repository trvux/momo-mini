package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var pool *pgxpool.Pool

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", checkHealth)
	mux.HandleFunc("GET /accounts/{id}", checkWallet)
	mux.HandleFunc("POST /accounts", createWallet)

	ctx := context.Background()

	pool, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", mux)
}

type account struct {
	Owner string `json:"owner"`
}

func checkHealth(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func checkWallet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, ("id must be a number"), http.StatusBadRequest)
		return
	}

	var owner string
	var balance int64
	err = pool.QueryRow(r.Context(), "select owner, balance from account where id=$1", idInt).Scan(&owner, &balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}
		log.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("your balance is " + strconv.FormatInt(balance, 10)))
}

func createWallet(w http.ResponseWriter, r *http.Request) {
	var req account
	var idInt int64

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Owner == "" {
		http.Error(w, "owner is required", http.StatusBadRequest)
		return
	}

	err = pool.QueryRow(r.Context(), "insert into account (owner) values ($1) returning id", req.Owner).Scan(&idInt)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%d", idInt)
}
