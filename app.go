package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "postgres"
	DB_NAME     = "goplan"
)

// App represents the application
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize sets up the database connection and routes for the app
func (a *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

// Run starts the app and serves on the specified addr
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8000", a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/sprints", a.getSprints).Methods("GET")
	a.Router.HandleFunc("/sprint", a.createSprint).Methods("POST")
	a.Router.HandleFunc("/sprint/{id:[0-9]+}", a.getSprint).Methods("GET")
	a.Router.HandleFunc("/sprint/{id:[0-9]+}", a.updateSprint).Methods("PUT")
	a.Router.HandleFunc("/sprint/{id:[0-9]+}", a.deleteSprint).Methods("DELETE")
}

func (a *App) createSprint(w http.ResponseWriter, r *http.Request) {
	var s sprint
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := s.createSprint(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, s)
}

func (a *App) getSprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid sprint ID")
		return
	}

	s := sprint{ID: id}
	if err := s.getSprint(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Sprint not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

func (a *App) updateSprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid sprint ID")
		return
	}

	var s sprint
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	s.ID = id

	if err := s.updateSprint(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

func (a *App) deleteSprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Sprint ID")
		return
	}

	s := sprint{ID: id}
	if err := s.deleteSprint(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func FetchIssuesByRepository(client *github.Client, owner string, repo string, options *github.IssueListByRepoOptions) ([]*github.Issue, error) {
	issues, _, err := client.Issues.ListByRepo(context.Background(), owner, repo, options)
	return issues, err
}

func (a *App) getSprints(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getSprints(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

func authenticateOauth(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return client
}

func GetMetrics(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	owner := params["owner"]
	repository := params["repository"]

	token := r.Header.Get("token")
	client := authenticateOauth(token)

	defer r.Body.Close()

	var filter github.IssueListByRepoOptions
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		log.Fatal(err)
		return
	}

	issues, err := FetchIssuesByRepository(client, owner, repository, &filter)
	if err != nil {
		log.Fatal(err)
		return
	}

	json.NewEncoder(w).Encode(&issues)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
