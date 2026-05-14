package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type app struct {
	db *sql.DB
}

var jwtSecret = []byte("your_secret_key_here")

// Websocket hub
type client struct {
	id   string
	conn *websocket.Conn
	send chan []byte
}

type hub struct {
	clients    map[string]*client
	register   chan *client
	unregister chan *client
	broadcast  chan struct {
		to  string
		msg []byte
	}
}

func newHub() *hub {
	return &hub{
		clients:    map[string]*client{},
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast: make(chan struct {
			to  string
			msg []byte
		}),
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c.id] = c
		case c := <-h.unregister:
			if _, ok := h.clients[c.id]; ok {
				delete(h.clients, c.id)
				close(c.send)
			}
		case b := <-h.broadcast:
			if c, ok := h.clients[b.to]; ok {
				select {
				case c.send <- b.msg:
				default:
				}
			}
		}
	}
}

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var globalHub = newHub()

type apiError struct {
	Error string `json:"error"`
}

type summaryResponse struct {
	ActiveJobs    int64   `json:"activeJobs"`
	EscrowBalance float64 `json:"escrowBalance"`
	Proposals     int64   `json:"proposals"`
	Rating        float64 `json:"rating"`
}

type jobResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Client      string   `json:"client"`
	ClientID    string   `json:"clientId"`
	Budget      float64  `json:"budget"`
	Deadline    string   `json:"deadline"`
	Status      string   `json:"status"`
	Skills      []string `json:"skills"`
	Description string   `json:"description"`
	Proposals   int64    `json:"proposals"`
}

type profileResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	Rating        float64  `json:"rating"`
	CompletedJobs int64    `json:"completedJobs"`
	Skills        []string `json:"skills"`
	Avatar        string   `json:"avatar,omitempty"`
}

type paymentsResponse struct {
	Available float64               `json:"available"`
	Escrowed  float64               `json:"escrowed"`
	History   []transactionResponse `json:"history"`
}

type transactionResponse struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
	Date   string  `json:"date"`
}

func main() {
	databaseURL := env("DATABASE_URL", "postgresql://freelance:freelance@localhost:5433/freelance_market?sslmode=disable")
	port := env("PORT", "8080")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := waitForDB(ctx, db); err != nil {
		log.Fatal(err)
	}
	if err := ensureSeedData(ctx, db); err != nil {
		log.Fatal(err)
	}

	a := &app{db: db}
	mux := http.NewServeMux()
	// auth endpoints
	mux.HandleFunc("/api/register", a.handleRegister)
	mux.HandleFunc("/api/login", a.handleLogin)
	// websocket chat
	mux.HandleFunc("/ws/chat", a.handleWSChat)
	// profile update and avatar upload
	mux.HandleFunc("/api/profile/photo", a.handleUploadAvatar)
	mux.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			a.profile(w, r)
			return
		}
		if r.Method == http.MethodPut {
			a.handleUpdateProfile(w, r)
			return
		}
		http.NotFound(w, r)
	})
	// serve uploaded files
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	mux.HandleFunc("/health", a.health)
	mux.HandleFunc("/api/summary", a.summary)
	// jobs: list and create
	mux.HandleFunc("/api/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			a.listJobs(w, r)
			return
		}
		if r.Method == http.MethodPost {
			a.createJob(w, r)
			return
		}
		http.NotFound(w, r)
	})
	// job apply (prefix to capture /api/jobs/{id}/apply)
	mux.HandleFunc("/api/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/apply") {
			a.applyToJob(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/profile/skills", a.updateSkills)
	mux.HandleFunc("/api/payments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			a.payments(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/payments/deposit", a.deposit)
	mux.HandleFunc("/api/payments/escrows", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && !strings.HasSuffix(r.URL.Path, "/release") {
			a.createEscrow(w, r)
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/release") {
			a.releaseEscrow(w, r)
			return
		}
		http.NotFound(w, r)
	})

	handler := cors(logging(mux))
	log.Printf("gateway listening on :%s", port)
	// ensure messages and profile columns
	if err := ensureMessagesTable(db); err != nil {
		log.Fatalf("ensure messages table: %v", err)
	}
	if err := ensureProfileColumns(db); err != nil {
		log.Fatalf("ensure profile columns: %v", err)
	}
	go globalHub.run()
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func ensureMessagesTable(db *sql.DB) error {
	_, err := db.Exec(`
create table if not exists messages (
  id bigserial primary key,
  from_user_id text not null,
  to_user_id text not null,
  content text not null,
  created_at timestamptz default now()
)
`)
	return err
}

func ensureProfileColumns(db *sql.DB) error {
	// add avatar_url column if missing
	_, err := db.Exec(`alter table users add column if not exists avatar_url text`)
	return err
}

func (a *app) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	if req.Role != "client" && req.Role != "freelancer" {
		writeErrorText(w, http.StatusBadRequest, "role must be 'client' or 'freelancer'")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeErrorText(w, http.StatusBadRequest, "email and password required")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var id string
	err = a.db.QueryRowContext(r.Context(), `insert into users (email, password_hash, full_name, role, is_verified) values ($1, $2, $3, $4, true) returning id`, req.Email, string(hash), req.Name, req.Role).Scan(&id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	token, err := createJWT(id, req.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"token": token, "id": id})
}

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	var id, hash, role string
	if err := a.db.QueryRowContext(r.Context(), `select id, password_hash, role from users where email = $1`, req.Email).Scan(&id, &hash, &role); err != nil {
		writeErrorText(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		writeErrorText(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := createJWT(id, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "id": id})
}

func createJWT(id, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func parseJWT(tokenStr string) (string, string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) { return jwtSecret, nil })
	if err != nil || !token.Valid {
		return "", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}
	uid, _ := claims["user_id"].(string)
	role, _ := claims["role"].(string)
	return uid, role, nil
}

func (a *app) handleWSChat(w http.ResponseWriter, r *http.Request) {
	// auth: token in query ?token= or header Authorization: Bearer
	token := r.URL.Query().Get("token")
	if token == "" {
		h := r.Header.Get("Authorization")
		if strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		}
	}
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	uid, _, err := parseJWT(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &client{id: uid, conn: conn, send: make(chan []byte, 256)}
	globalHub.register <- c

	// read loop
	go func() {
		defer func() {
			globalHub.unregister <- c
			c.conn.Close()
		}()
		for {
			var msg struct {
				To      string `json:"to"`
				Content string `json:"content"`
			}
			if err := c.conn.ReadJSON(&msg); err != nil {
				return
			}
			// persist
			if _, err := a.db.ExecContext(r.Context(), `insert into messages (from_user_id, to_user_id, content) values ($1, $2, $3)`, c.id, msg.To, msg.Content); err != nil {
				log.Printf("insert message: %v", err)
			}
			// forward
			b, _ := json.Marshal(map[string]string{"from": c.id, "content": msg.Content})
			globalHub.broadcast <- struct {
				to  string
				msg []byte
			}{to: msg.To, msg: b}
		}
	}()

	// write loop
	go func() {
		for m := range c.send {
			c.conn.WriteMessage(websocket.TextMessage, m)
		}
	}()
}

func (a *app) handleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	// token auth
	token := r.Header.Get("Authorization")
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	uid, _, err := parseJWT(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()
	// ensure dir
	_ = os.MkdirAll("uploads/avatars", 0755)
	fname := fmt.Sprintf("uploads/avatars/%s-%s", uid, header.Filename)
	out, err := os.Create(fname)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	url := "/" + strings.TrimPrefix(fname, "./")
	if strings.HasPrefix(url, "uploads/") {
		url = "/" + url
	}
	// store in db
	if _, err := a.db.ExecContext(r.Context(), `update users set avatar_url = $1 where id = $2`, "/"+fname, uid); err != nil {
		log.Printf("update avatar url: %v", err)
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": "/" + fname})
}

func (a *app) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.UserID == "" {
		writeErrorText(w, http.StatusBadRequest, "userId required")
		return
	}
	if _, err := a.db.ExecContext(r.Context(), `update users set full_name = $1, avatar_url = coalesce(nullif($2, ''), avatar_url) where id = $3`, req.Name, req.Avatar, req.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	profile, err := a.getProfile(r.Context(), req.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (a *app) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *app) summary(w http.ResponseWriter, r *http.Request) {
	var out summaryResponse
	if err := a.db.QueryRowContext(r.Context(), `
select
  (select count(*) from jobs where status in ('open', 'in_progress')) as active_jobs,
  (select coalesce(sum(escrow_cents), 0) from payment_accounts) as escrow_cents,
  (select count(*) from proposals) as proposals,
  (select coalesce(round(avg(rating)::numeric, 2), 0) from users where role = 'freelancer') as rating`).Scan(
		&out.ActiveJobs, centsScanner(&out.EscrowBalance), &out.Proposals, &out.Rating,
	); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *app) listJobs(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), `
select
  j.id,
  j.title,
  j.description,
  j.budget_cents,
  j.currency,
  j.status,
  coalesce(j.deadline::text, ''),
  u.id,
  u.full_name,
  coalesce(array_agg(distinct js.skill) filter (where js.skill is not null), '{}') as skills,
  count(distinct p.id) as proposals
from jobs j
join users u on u.id = j.client_id
left join job_skills js on js.job_id = j.id
left join proposals p on p.job_id = j.id
group by j.id, u.id, u.full_name
order by j.created_at desc`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	jobs := make([]jobResponse, 0)
	for rows.Next() {
		var budgetCents int64
		var currency string
		var skills pqStringArray
		job := jobResponse{}
		if err := rows.Scan(
			&job.ID, &job.Title, &job.Description, &budgetCents, &currency, &job.Status,
			&job.Deadline, &job.ClientID, &job.Client, &skills, &job.Proposals,
		); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		job.Budget = centsToMoney(budgetCents)
		job.Skills = []string(skills)
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (a *app) createJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientID    string   `json:"clientId"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Budget      float64  `json:"budget"`
		Deadline    string   `json:"deadline"`
		Skills      []string `json:"skills"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.ClientID == "" {
		req.ClientID = mustDemoID(r.Context(), a.db, "client@demo.local")
	}
	if req.Title == "" || req.Budget <= 0 {
		writeErrorText(w, http.StatusBadRequest, "title and budget are required")
		return
	}
	if req.Description == "" {
		req.Description = "Client is ready to discuss scope, timeline, and expected deliverables."
	}
	if req.Deadline == "" {
		req.Deadline = time.Now().AddDate(0, 0, 21).Format("2006-01-02")
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rollback(tx)

	var jobID string
	if err := tx.QueryRowContext(r.Context(), `
insert into jobs (client_id, title, description, budget_cents, deadline, status)
values ($1, $2, $3, $4, $5, 'open')
returning id`, req.ClientID, req.Title, req.Description, moneyToCents(req.Budget), req.Deadline).Scan(&jobID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if len(req.Skills) == 0 {
		req.Skills = []string{"Web Design", "Frontend", "Communication"}
	}
	for _, skill := range cleanSkills(req.Skills) {
		if _, err := tx.ExecContext(r.Context(), `insert into job_skills (job_id, skill) values ($1, $2) on conflict do nothing`, jobID, skill); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	job, err := a.getJob(r.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

func (a *app) applyToJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FreelancerID  string  `json:"freelancerId"`
		CoverLetter   string  `json:"coverLetter"`
		Bid           float64 `json:"bid"`
		EstimatedDays int     `json:"estimatedDays"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.FreelancerID == "" {
		req.FreelancerID = mustDemoID(r.Context(), a.db, "freelancer@demo.local")
	}
	if strings.TrimSpace(req.CoverLetter) == "" {
		req.CoverLetter = "I can help with this project and deliver clear updates along the way."
	}
	if req.EstimatedDays <= 0 {
		req.EstimatedDays = 7
	}
	if req.Bid <= 0 {
		var budgetCents int64
		_ = a.db.QueryRowContext(r.Context(), `select budget_cents from jobs where id = $1`, pathValue(r, "id")).Scan(&budgetCents)
		req.Bid = centsToMoney(budgetCents)
	}
	_, err := a.db.ExecContext(r.Context(), `
insert into proposals (job_id, freelancer_id, cover_letter, bid_cents, estimated_days)
values ($1, $2, $3, $4, $5)
on conflict (job_id, freelancer_id) do update
set cover_letter = excluded.cover_letter,
    bid_cents = excluded.bid_cents,
    estimated_days = excluded.estimated_days,
    updated_at = now()`,
		pathValue(r, "id"), req.FreelancerID, req.CoverLetter, moneyToCents(req.Bid), req.EstimatedDays)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "proposal_sent"})
}

func (a *app) profile(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = mustDemoID(r.Context(), a.db, "freelancer@demo.local")
	}
	profile, err := a.getProfile(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (a *app) updateSkills(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string   `json:"userId"`
		Skills []string `json:"skills"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.UserID == "" {
		req.UserID = mustDemoID(r.Context(), a.db, "freelancer@demo.local")
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rollback(tx)
	if _, err := tx.ExecContext(r.Context(), `delete from user_skills where user_id = $1`, req.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	for _, skill := range cleanSkills(req.Skills) {
		if _, err := tx.ExecContext(r.Context(), `insert into user_skills (user_id, skill) values ($1, $2)`, req.UserID, skill); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	profile, err := a.getProfile(r.Context(), req.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (a *app) payments(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = mustDemoID(r.Context(), a.db, "client@demo.local")
	}
	out := paymentsResponse{History: []transactionResponse{}}
	if err := a.db.QueryRowContext(r.Context(), `
select available_cents, escrow_cents
from payment_accounts
where user_id = $1`, userID).Scan(centsScanner(&out.Available), centsScanner(&out.Escrowed)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = a.db.ExecContext(r.Context(), `insert into payment_accounts (user_id) values ($1)`, userID)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	rows, err := a.db.QueryContext(r.Context(), `
select id, type, amount_cents, status, created_at
from transactions
where user_id = $1
order by created_at desc
limit 20`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cents int64
		var created time.Time
		item := transactionResponse{}
		if err := rows.Scan(&item.ID, &item.Type, &cents, &item.Status, &created); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		item.Amount = centsToMoney(cents)
		item.Date = created.Format("2006-01-02")
		out.History = append(out.History, item)
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *app) deposit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string  `json:"userId"`
		Amount float64 `json:"amount"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.UserID == "" {
		req.UserID = mustDemoID(r.Context(), a.db, "client@demo.local")
	}
	if req.Amount <= 0 {
		writeErrorText(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rollback(tx)
	cents := moneyToCents(req.Amount)
	if _, err := tx.ExecContext(r.Context(), `
insert into payment_accounts (user_id, available_cents)
values ($1, $2)
on conflict (user_id) do update
set available_cents = payment_accounts.available_cents + excluded.available_cents`, req.UserID, cents); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `
insert into transactions (user_id, type, amount_cents, status, provider)
values ($1, 'deposit', $2, 'completed', 'demo')`, req.UserID, cents); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.payments(w, r)
}

func (a *app) createEscrow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JobID        string  `json:"jobId"`
		ClientID     string  `json:"clientId"`
		FreelancerID string  `json:"freelancerId"`
		Amount       float64 `json:"amount"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.ClientID == "" {
		req.ClientID = mustDemoID(r.Context(), a.db, "client@demo.local")
	}
	if req.FreelancerID == "" {
		req.FreelancerID = mustDemoID(r.Context(), a.db, "freelancer@demo.local")
	}
	if req.JobID == "" || req.Amount <= 0 {
		writeErrorText(w, http.StatusBadRequest, "jobId and amount are required")
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rollback(tx)
	cents := moneyToCents(req.Amount)
	res, err := tx.ExecContext(r.Context(), `
update payment_accounts
set available_cents = available_cents - $1,
    escrow_cents = escrow_cents + $1
where user_id = $2 and available_cents >= $1`, cents, req.ClientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeErrorText(w, http.StatusConflict, "insufficient funds")
		return
	}
	var escrowID string
	if err := tx.QueryRowContext(r.Context(), `
insert into escrows (job_id, client_id, freelancer_id, amount_cents)
values ($1, $2, $3, $4)
on conflict (job_id) do update set updated_at = now()
returning id`, req.JobID, req.ClientID, req.FreelancerID, cents).Scan(&escrowID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `
insert into transactions (user_id, job_id, escrow_id, type, amount_cents, status)
values ($1, $2, $3, 'escrow_hold', $4, 'completed')`, req.ClientID, req.JobID, escrowID, cents); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": escrowID, "status": "held"})
}

func (a *app) releaseEscrow(w http.ResponseWriter, r *http.Request) {
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rollback(tx)
	var jobID, clientID, freelancerID string
	var cents int64
	var status string
	if err := tx.QueryRowContext(r.Context(), `
select job_id, client_id, freelancer_id, amount_cents, status
from escrows
where id = $1
for update`, pathValue(r, "id")).Scan(&jobID, &clientID, &freelancerID, &cents, &status); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if status != "held" {
		writeErrorText(w, http.StatusConflict, "escrow is not held")
		return
	}
	if _, err := tx.ExecContext(r.Context(), `update payment_accounts set escrow_cents = escrow_cents - $1 where user_id = $2`, cents, clientID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `
insert into payment_accounts (user_id, available_cents)
values ($1, $2)
on conflict (user_id) do update set available_cents = payment_accounts.available_cents + excluded.available_cents`, freelancerID, cents); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `update escrows set status = 'released', released_at = now() where id = $1`, pathValue(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `
insert into transactions (user_id, job_id, escrow_id, type, amount_cents, status)
values ($1, $2, $3, 'escrow_release', $4, 'completed')`, freelancerID, jobID, pathValue(r, "id"), cents); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `update jobs set status = 'completed', completed_at = now() where id = $1`, jobID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "released"})
}

func (a *app) getJob(ctx context.Context, id string) (*jobResponse, error) {
	row := a.db.QueryRowContext(ctx, `
select
  j.id, j.title, j.description, j.budget_cents, j.currency, j.status,
  coalesce(j.deadline::text, ''), u.id, u.full_name,
  coalesce(array_agg(distinct js.skill) filter (where js.skill is not null), '{}') as skills,
  count(distinct p.id) as proposals
from jobs j
join users u on u.id = j.client_id
left join job_skills js on js.job_id = j.id
left join proposals p on p.job_id = j.id
where j.id = $1
group by j.id, u.id, u.full_name`, id)
	var budgetCents int64
	var currency string
	var skills pqStringArray
	job := jobResponse{}
	if err := row.Scan(&job.ID, &job.Title, &job.Description, &budgetCents, &currency, &job.Status, &job.Deadline, &job.ClientID, &job.Client, &skills, &job.Proposals); err != nil {
		return nil, err
	}
	job.Budget = centsToMoney(budgetCents)
	job.Skills = []string(skills)
	return &job, nil
}

func (a *app) getProfile(ctx context.Context, userID string) (*profileResponse, error) {
	row := a.db.QueryRowContext(ctx, `
select u.id, u.full_name, u.role, u.rating, u.completed_jobs, u.avatar_url,
	   coalesce(array_agg(us.skill order by us.skill) filter (where us.skill is not null), '{}') as skills
from users u
left join user_skills us on us.user_id = u.id
where u.id = $1
group by u.id`, userID)
	var skills pqStringArray
	out := profileResponse{}
	if err := row.Scan(&out.ID, &out.Name, &out.Role, &out.Rating, &out.CompletedJobs, &out.Avatar, &skills); err != nil {
		return nil, err
	}
	out.Skills = []string(skills)
	return &out, nil
}

func ensureSeedData(ctx context.Context, db *sql.DB) error {
	var count int
	if err := db.QueryRowContext(ctx, `select count(*) from users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	var clientID, freelancerID, designerID string
	if err := tx.QueryRowContext(ctx, `
insert into users (email, password_hash, full_name, role, bio, rating, completed_jobs, is_verified)
values ('client@demo.local', 'demo', 'Apex Studio', 'client', 'Product team hiring creative experts.', 0, 0, true)
returning id`).Scan(&clientID); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `
insert into users (email, password_hash, full_name, role, bio, rating, completed_jobs, is_verified)
values ('freelancer@demo.local', 'demo', 'Aruzhan Karimova', 'freelancer', 'Frontend and product designer for growing teams.', 4.8, 37, true)
returning id`).Scan(&freelancerID); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `
insert into users (email, password_hash, full_name, role, bio, rating, completed_jobs, is_verified)
values ('designer@demo.local', 'demo', 'Daniyar Saken', 'freelancer', 'Brand designer focused on launch visuals.', 4.6, 21, true)
returning id`).Scan(&designerID); err != nil {
		return err
	}
	for _, item := range []struct {
		userID string
		skills []string
	}{
		{freelancerID, []string{"Web Design", "Frontend", "Branding", "SEO", "Payments"}},
		{designerID, []string{"Branding", "Social Media", "Figma", "Illustration"}},
	} {
		for _, skill := range item.skills {
			if _, err := tx.ExecContext(ctx, `insert into user_skills (user_id, skill) values ($1, $2)`, item.userID, skill); err != nil {
				return err
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `
insert into payment_accounts (user_id, available_cents, escrow_cents)
values ($1, 420000, 310000), ($2, 125000, 0), ($3, 78000, 0)`, clientID, freelancerID, designerID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
insert into transactions (user_id, type, amount_cents, status, provider, created_at)
values
($1, 'deposit', 200000, 'completed', 'demo', now() - interval '4 days'),
($1, 'escrow_hold', 90000, 'completed', 'demo', now() - interval '2 days'),
($2, 'escrow_release', 125000, 'completed', 'demo', now() - interval '1 day')`, clientID, freelancerID); err != nil {
		return err
	}

	jobs := []struct {
		title       string
		description string
		budget      int64
		deadline    string
		status      string
		skills      []string
	}{
		{"Create a modern landing page for a new online course", "Design a clear page that explains the offer, collects leads, and works well on mobile.", 120000, "2026-06-04", "open", []string{"Web Design", "Copywriting", "SEO", "Analytics"}},
		{"Build a booking flow for a small fitness studio", "Improve the appointment flow and make it easier for customers to reserve sessions.", 90000, "2026-05-29", "in_progress", []string{"UX", "Frontend", "Payments", "Testing"}},
		{"Prepare brand visuals for a product launch", "Create campaign visuals for Instagram, presentation slides, and product announcements.", 65000, "2026-06-11", "open", []string{"Branding", "Social Media", "Figma"}},
	}
	for i, item := range jobs {
		var jobID string
		if err := tx.QueryRowContext(ctx, `
insert into jobs (client_id, title, description, budget_cents, deadline, status, selected_freelancer_id)
values ($1, $2, $3, $4, $5, $6, $7)
returning id`, clientID, item.title, item.description, item.budget, item.deadline, item.status, nullableFreelancer(i, freelancerID)).Scan(&jobID); err != nil {
			return err
		}
		for _, skill := range item.skills {
			if _, err := tx.ExecContext(ctx, `insert into job_skills (job_id, skill) values ($1, $2)`, jobID, skill); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `
insert into proposals (job_id, freelancer_id, cover_letter, bid_cents, estimated_days, status)
values ($1, $2, 'I can deliver this with clear milestones and weekly updates.', $3, 7, 'pending')`, jobID, freelancerID, item.budget); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func nullableFreelancer(index int, id string) sql.NullString {
	if index == 1 {
		return sql.NullString{String: id, Valid: true}
	}
	return sql.NullString{}
}

func mustDemoID(ctx context.Context, db *sql.DB, email string) string {
	var id string
	if err := db.QueryRowContext(ctx, `select id from users where email = $1`, email).Scan(&id); err != nil {
		panic(err)
	}
	return id
}

func waitForDB(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// pathValue extracts an ID-like path segment for routes like
// /api/jobs/{id}/apply or /api/payments/escrows/{id}/release
func pathValue(r *http.Request, key string) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if last == "apply" || last == "release" || last == "photo" {
		if len(parts) >= 2 {
			return parts[len(parts)-2]
		}
		return ""
	}
	return last
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeErrorText(w, status, err.Error())
}

func writeErrorText(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiError{Error: message})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func rollback(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.Printf("rollback: %v", err)
	}
}

func centsToMoney(cents int64) float64 {
	return float64(cents) / 100
}

func moneyToCents(amount float64) int64 {
	return int64(amount*100 + 0.5)
}

func cleanSkills(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[strings.ToLower(value)] {
			continue
		}
		seen[strings.ToLower(value)] = true
		out = append(out, value)
	}
	return out
}

type centsScannerTarget struct {
	target *float64
}

func centsScanner(target *float64) *centsScannerTarget {
	return &centsScannerTarget{target: target}
}

func (s *centsScannerTarget) Scan(value any) error {
	switch v := value.(type) {
	case int64:
		*s.target = centsToMoney(v)
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*s.target = centsToMoney(n)
	default:
		return fmt.Errorf("unsupported cents value %T", value)
	}
	return nil
}

type pqStringArray []string

func (a *pqStringArray) Scan(src any) error {
	if src == nil {
		*a = []string{}
		return nil
	}
	raw, ok := src.(string)
	if !ok {
		if bytes, ok := src.([]byte); ok {
			raw = string(bytes)
		} else {
			return fmt.Errorf("unsupported array value %T", src)
		}
	}
	raw = strings.Trim(raw, "{}")
	if raw == "" {
		*a = []string{}
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, strings.Trim(part, `"`))
	}
	*a = out
	return nil
}
