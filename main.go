package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type aula struct {
	NOME_AULA   string
	HORA_INICIO string
	HORA_FIM    string
	NOME_DIA    string
	ID_AULA     int
}

var (
	aulaAtual string = ""
	opt       string
	clear     map[string]func()
	idAula    int = 0
)

func takeErr(err error) {
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}

func init() {
	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func ClearScreen() {
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	} else {
		panic("Plataforma inválida!")
	}
}

func consultaDiaSemana() int {
	return int(time.Now().Weekday())
}

func listAulaDiaFile(dia int, db *sql.DB) {
	file, err := os.Create("aulas.txt")
	takeErr(err)
	writer := bufio.NewWriter(file)

	query := searchDB(db, dia)

	for query.Next() {
		var al aula
		query.Scan(&al.NOME_AULA,
			&al.HORA_INICIO,
			&al.HORA_FIM,
			&al.NOME_DIA,
			&al.ID_AULA)

		texto := fmt.Sprintf("AULA: %s\nINÍCIO: %s\nFIM: %s\nDIA DA SEMANA: %s\n-------------\n", al.NOME_AULA, al.HORA_INICIO, al.HORA_FIM, al.NOME_DIA)
		_, err = writer.WriteString(texto)
		takeErr(err)
	}
	writer.Flush()
}

func listAulaDiaSTOUT(dia int, db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	query := searchDB(db, dia)

	for query.Next() {
		var al aula
		query.Scan(&al.NOME_AULA,
			&al.HORA_INICIO,
			&al.HORA_FIM,
			&al.NOME_DIA,
			&al.ID_AULA)
		texto := fmt.Sprintf("AULA: %s\nINÍCIO: %s\nFIM: %s\nDIA DA SEMANA: %s\n-------------\n", al.NOME_AULA, al.HORA_INICIO, al.HORA_FIM, al.NOME_DIA)
		fmt.Println(texto)
	}
}

func operator(opt *string, wg *sync.WaitGroup, sq *sql.DB) {
	defer wg.Done()
	wgFunc := new(sync.WaitGroup)
	wgFunc.Add(1)

	ClearScreen()
	switch *opt {
	case "1":
		listenAula(consultaDiaSemana(), sq, wgFunc)
	case "2":
		listAulaDiaFile(consultaDiaSemana(), sq)
	case "3":
		listAulaDiaSTOUT(consultaDiaSemana(), sq, wgFunc)
	default:
		fmt.Println("opt - inválida")
		time.Sleep(2)
	}
}

func main() {
	SQLite := openDB()
	defer SQLite.Close()

	for {
		wg := new(sync.WaitGroup)
		wg.Add(1)

		fmt.Println("+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+")
		fmt.Println("+ 1 - Notificar 10min antes da aula         +")
		fmt.Println("+ 2 - Escrever em arquivo (aulas do dia)    +")
		fmt.Println("+ 3 - Notificar aulas do dia                +")
		fmt.Println("+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+")
		fmt.Print("-> ")
		fmt.Scanln(&opt)
		operator(&opt, wg, SQLite)
	}
}
