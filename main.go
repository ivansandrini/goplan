package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

func FetchIssuesByRepository(client *github.Client, owner string, repo string, options *github.IssueListByRepoOptions) ([]*github.Issue, error) {
	issues, _, err := client.Issues.ListByRepo(context.Background(), owner, repo, options)
	return issues, err
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/metrics/{owner}/{repository}", GetMetrics).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
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
