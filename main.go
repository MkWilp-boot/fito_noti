package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/toast.v1"
)

type aula struct {
	NOME_AULA   string
	HORA_INICIO string
	HORA_FIM    string
	NOME_DIA    string
	ID_AULA     int
}

const (
	SETE_45 string = "19:45"
	OITO_30 string = "20:30"
	NOVE_30 string = "21:30"
	DEZ_15  string = "22:15"
	ONZE_0  string = "23:00"
)

var (
	aulaAtual string = ""
	opt       string
	clear     map[string]func()
	idAula    int = 0
)

func takeErr(err error) {
	if err != nil {
		panic(err)
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

func openDB() *sql.DB {
	SQLite, err := sql.Open("sqlite3", "aula.db")
	takeErr(err)
	SQLite.SetMaxOpenConns(1)

	return SQLite
}

func consultaDiaSemana() int {
	return int(time.Now().Weekday())
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

func getAulaAtual(aulas []aula) aula {
	hora := time.Now().Hour()
	min := time.Now().Minute()

	aul := aula{
		NOME_AULA:   "NONE",
		HORA_INICIO: "00:00",
		HORA_FIM:    "00:00",
		NOME_DIA:    "NONE",
		ID_AULA:     -1,
	}
	if hora == 19 && min < 45 {
		aul = aulas[0]
	} else if hora == 19 && min >= 45 {
		aul = aulas[1]
	} else if hora == 20 && min < 30 {
		aul = aulas[1]
	} else if hora == 20 && min >= 45 {
		aul = aulas[2]
	} else if hora == 21 && min < 30 {
		aul = aulas[2]
	} else if hora == 21 && min >= 30 {
		aul = aulas[3]
	} else if hora == 22 && min < 15 {
		aul = aulas[3]
	} else if hora == 22 && min >= 15 {
		aul = aulas[4]
	}
	return aul
}

func listenAula(dia int, db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	sliceAula := make([]aula, 0)

	query := searchDB(db, dia)
	for query.Next() {
		var al aula
		query.Scan(&al.NOME_AULA,
			&al.HORA_INICIO,
			&al.HORA_FIM,
			&al.NOME_DIA,
			&al.ID_AULA)
		sliceAula = append(sliceAula, al)
	}

	fmt.Println("Modo de notificação ativado")
	fmt.Println("Para sair aperte CTRL + C")

	for {
		aula_atual := getAulaAtual(sliceAula)
		if aula_atual.ID_AULA == -1 {
			txt := fmt.Sprint("Suas aulas acabaram por hoje")
			go genNoti("App", "Notificação de aula", txt)
			break
		} else if aula_atual.ID_AULA != idAula {
			idAula = aula_atual.ID_AULA
			txt := fmt.Sprintf("Sua aula de %s iniciará em breve", aula_atual.NOME_AULA)
			go genNoti("App", "Notificação de aula", txt)
		}
		fmt.Println(aula_atual)

		time.Sleep(time.Second * 30)
	}
}

func genNoti(appID, title, message string) {
	notification := toast.Notification{
		AppID:   appID,
		Title:   title,
		Message: message,
	}
	err := notification.Push()
	takeErr(err)
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
