package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Job struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Budget      float64  `json:"budget"`
	Status      string   `json:"status"`
	Client      string   `json:"client"`
}

type Service struct {
	db    *sql.DB
	cache map[string]string // Simple in-memory cache simulating Redis for offline environment
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:    db,
		cache: make(map[string]string),
	}
}

func (s *Service) Create(ctx context.Context, clientID, title, desc string, budget float64) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, 
		"INSERT INTO jobs (client_id, title, description, budget_cents) VALUES ($1, $2, $3, $4) RETURNING id",
		clientID, title, desc, int64(budget*100)).Scan(&id)
	return id, err
}

func (s *Service) List(ctx context.Context) ([]Job, error) {
	// Check Cache (Simulating Redis GET)
	if val, ok := s.cache["all_jobs"]; ok {
		var list []Job
		if err := json.Unmarshal([]byte(val), &list); err == nil {
			return list, nil
		}
	}

	rows, err := s.db.QueryContext(ctx, 
		"SELECT j.id, j.title, j.description, j.budget_cents, j.status, u.full_name FROM jobs j JOIN users u ON u.id = j.client_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Job
	for rows.Next() {
		var j Job
		var cents int64
		if err := rows.Scan(&j.ID, &j.Title, &j.Description, &cents, &j.Status, &j.Client); err == nil {
			j.Budget = float64(cents) / 100
			list = append(list, j)
		}
	}

	// Update Cache (Simulating Redis SET with TTL)
	if data, err := json.Marshal(list); err == nil {
		s.cache["all_jobs"] = string(data)
	}

	return list, nil
}
