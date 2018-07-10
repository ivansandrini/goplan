package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize(
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/sprints", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentSprint(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/sprint/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Sprint not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Sprint not found'. Got '%s'", m["error"])
	}
}

func TestCreateSprint(t *testing.T) {
	clearTable()

	payload := []byte(`{"name":"test sprint","start_date":"2018-01-01","end_date":"2018-02-01"}`)

	req, _ := http.NewRequest("POST", "/sprint", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test sprint" {
		t.Errorf("Expected sprint name to be 'test sprint'. Got '%v'", m["name"])
	}

	if m["start_date"] != "2018-01-01" {
		t.Errorf("Expected sprint start_date to be '2018-01-01'. Got '%v'", m["start_date"])
	}

	if m["end_date"] != "2018-02-01" {
		t.Errorf("Expected sprint end_date to be '2018-02-01'. Got '%v'", m["end_date"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected sprint ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addSprints(1)

	req, _ := http.NewRequest("GET", "/sprint/1", nil)
	response := executeRequest(req)

	log.Println(response.Body)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addSprints(1)

	req, _ := http.NewRequest("GET", "/sprint/1", nil)
	response := executeRequest(req)
	var originalSprint map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalSprint)

	payload := []byte(`{"name":"test sprint","start_date":"2018-01-01","end_date":"2018-02-01"}`)

	req, _ = http.NewRequest("PUT", "/sprint/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalSprint["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalSprint["id"], m["id"])
	}

	if m["name"] == originalSprint["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalSprint["name"], m["name"], m["name"])
	}

	if m["start_date"] == originalSprint["start_date"] {
		t.Errorf("Expected the start_date to change from '%v' to '%v'. Got '%v'", originalSprint["start_date"], m["start_date"], m["start_date"])
	}

	if m["end_date"] == originalSprint["end_date"] {
		t.Errorf("Expected the end_date to change from '%v' to '%v'. Got '%v'", originalSprint["end_date"], m["end_date"], m["end_date"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addSprints(1)

	req, _ := http.NewRequest("GET", "/sprint/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/sprint/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/sprint/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)

	deleteTableExists()
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func deleteTableExists() {
	if _, err := a.DB.Exec(tableDelectionQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM sprints")
	a.DB.Exec("ALTER SEQUENCE sprints_id_seq RESTART WITH 1")
}

func addSprints(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO sprints(name, start_date, end_date) VALUES($1, $2, $3)", "Sprint "+strconv.Itoa(i),
			time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC),
			time.Date(2018, 2, 1, 12, 0, 0, 0, time.UTC))
	}
}

const tableCreationQuery = `CREATE TABLE public.sprints (
	id serial NOT NULL,
	name varchar(100) NOT NULL,
	start_date date NULL,
	end_date date NULL,
	CONSTRAINT sprint_pkey PRIMARY KEY (id)
)
WITH (
	OIDS=FALSE
) ;`

const tableDelectionQuery = `DROP TABLE public.sprints;`
