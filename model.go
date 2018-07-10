package main

import "database/sql"

type sprint struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

func (s *sprint) getSprint(db *sql.DB) error {
	return db.QueryRow("SELECT name, start_date, end_date FROM sprints WHERE id=$1",
		s.ID).Scan(&s.Name, &s.StartDate, &s.EndDate)
}

func (s *sprint) updateSprint(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE sprints SET name=$1, start_date=$2, end_date=$3 WHERE id=$4",
			s.Name, s.StartDate, s.EndDate, s.ID)

	return err
}

func (s *sprint) deleteSprint(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sprints WHERE id=$1", s.ID)

	return err
}

func (s *sprint) createSprint(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO sprints(name, start_date, end_date) VALUES($1, $2, $3) RETURNING id",
		s.Name, s.StartDate, s.EndDate).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}

func getSprints(db *sql.DB, start, count int) ([]sprint, error) {
	rows, err := db.Query(
		"SELECT * FROM sprints LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sprints := []sprint{}

	for rows.Next() {
		var s sprint
		if err := rows.Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate); err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}

	return sprints, nil
}
