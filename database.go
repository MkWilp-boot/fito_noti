package main

import (
	"database/sql"
)

func openDB() *sql.DB {
	SQLite, err := sql.Open("sqlite3", "aula.db")
	takeErr(err)
	SQLite.SetMaxOpenConns(1)

	return SQLite
}

func searchDB(db *sql.DB, dia int) *sql.Rows {
	query, err := db.Query(`SELECT 
								NOME_AULA,
								HORA_INICIO,
								HORA_FIM,
								SEMANA.NOME_DIA,
								ID_AULA
							FROM AULAS
							LEFT JOIN SEMANA 
							ON SEMANA.ID_DIA = AULAS.ID_DIA
							WHERE AULAS.ID_DIA = ?`, dia)
	takeErr(err)
	return query
}
