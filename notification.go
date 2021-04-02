package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gopkg.in/toast.v1"
)

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
