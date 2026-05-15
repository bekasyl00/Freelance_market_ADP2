package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"freelance-market/internal/payments"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	svc := payments.NewService(db)

	mux := http.NewServeMux()
	
	// Get Balance
	mux.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		bal, _ := svc.GetBalance(r.Context(), userID)
		json.NewEncoder(w).Encode(map[string]float64{"balance": bal})
	})

	log.Println("Payment Service listening on :8083")
	http.ListenAndServe(":8083", mux)
}
