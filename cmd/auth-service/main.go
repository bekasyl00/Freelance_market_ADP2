package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"freelance-market/internal/auth"
	_ "github.com/lib/pq"
	// "github.com/nats-io/nats.go" // Uncomment when internet is available
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// NATS connection (Ready for integration)
	/*
	nc, _ := nats.Connect(os.Getenv("NATS_URL"))
	if nc != nil {
		defer nc.Close()
		nc.Subscribe("jobs.created", func(m *nats.Msg) {
			log.Printf("Received NATS message: %s", string(m.Data))
		})
	}
	*/

	svc := auth.NewService(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Email, Password string }
		json.NewDecoder(r.Body).Decode(&req)
		id, role, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id": id, "role": role})
	})

	log.Println("Auth Service listening on :8081")
	http.ListenAndServe(":8081", mux)
}
