package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"freelance-market/internal/jobs"
	_ "github.com/lib/pq"
	// "github.com/nats-io/nats.go" // Uncomment when internet is available
)

func main() {
	// DB connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// NATS connection (Ready for integration)
	/*
	nc, _ := nats.Connect(os.Getenv("NATS_URL"))
	if nc != nil {
		defer nc.Close()
		log.Println("Connected to NATS")
	}
	*/

	svc := jobs.NewService(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				ClientID, Title, Desc string
				Budget                float64
			}
			json.NewDecoder(r.Body).Decode(&req)
			id, err := svc.Create(r.Context(), req.ClientID, req.Title, req.Desc, req.Budget)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			// Usage of Message Queue (Ready for integration)
			/*
			if nc != nil {
				msg := map[string]string{"job_id": id, "title": req.Title}
				data, _ := json.Marshal(msg)
				nc.Publish("jobs.created", data)
				log.Println("Published jobs.created to NATS")
			}
			*/

			json.NewEncoder(w).Encode(map[string]string{"id": id})
		} else {
			list, _ := svc.List(r.Context())
			json.NewEncoder(w).Encode(list)
		}
	})

	log.Println("Job Service listening on :8082")
	http.ListenAndServe(":8082", mux)
}
