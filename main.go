package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	base_url   = "https://host.globalhitss.com"
	login_url  = base_url + "/Security/Login"
	cookie_url = base_url + "/Horas/CapturaHoras2"
)

var calendar = CalendarForYear(time.Now().Year())

type Activities struct {
	Id_CapturaActividad string
	// Id_Actividad        string
	HorasCapturadas string
	Comentario      string
	HorasExtras     bool
	HorasNocturnas  bool
	Bloqueada       bool
}

type FillData struct {
	Id_Proyecto                int
	Id_Recurso                 string
	FechaDia                   string
	Comentario                 string
	Actividades                []Activities
	pantallaCaptura            bool
	Latitude                   float64
	Longitude                  float64
	__RequestVerificationToken string
}

func login(username string, password string) (string, error) {
	var aspx string

	body := []byte(fmt.Sprintf(
		"UserName=%s&Password=%s&Language=pt&bandera=1", username, password,
	))
	payload := bytes.NewBuffer(body)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("POST", base_url, payload)
	if err != nil {
		fmt.Printf("Erro ao criar o POST. %e", err)
		os.Exit(100)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Erro ao executar o login. %e", err)
		os.Exit(200)
	}

	cookies := response.Cookies()
	for _, c := range cookies {
		if c.Name == ".ASPXAUTH" {
			aspx = c.Value
		}
	}

	req, err = http.NewRequest("GET", cookie_url, nil)
	if err != nil {
		fmt.Printf("Erro ao criar o POST. %e", err)
		os.Exit(300)
	}
	req.Header.Add("Cookie", fmt.Sprintf(".ASPXAUTH=%s", aspx))
	req.Header.Add("Cookie", "HOST=Cultura=pt&Auxiliar=")
	req.Header.Add("Referer", login_url)

	response, err = client.Do(req)
	if err != nil {
		fmt.Printf("Erro ao executar o login. %e", err)
		os.Exit(400)
	}

	cookies = response.Cookies()
	for _, c := range cookies {
		if c.Name == "__RequestVerificationToken" {
			return c.Value, nil
		}
	}

	return "", fmt.Errorf("Cookie não encontrado.")
}

func fillHours(requestToken string, calendar Calendar) {
	var data FillData
	activities := make([]Activities, 1)

	activities[0] = Activities{
		Id_CapturaActividad: "0",
		// Id_Actividad:        "973214",
		HorasCapturadas: "8.0",
		Comentario:      "t",
		HorasExtras:     false,
		HorasNocturnas:  false,
		Bloqueada:       false,
	}

	url := "https://host.globalhitss.com/Horas/ActualizaActividades2"
	projectId := 73556
	ResourceId := "50900135"

	current_time := time.Now()
	startDate := time.Date(current_time.Year(), current_time.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
		if !calendar[d] {
			continue
		}

		data = FillData{
			Id_Proyecto:                projectId,
			Id_Recurso:                 ResourceId,
			FechaDia:                   d.Format("2006-01-02"),
			Comentario:                 "",
			Actividades:                activities,
			pantallaCaptura:            true,
			Latitude:                   0,
			Longitude:                  0,
			__RequestVerificationToken: requestToken,
		}

		payload, _ := json.Marshal(data)

		_, err := http.Post(url, "application/json",
			bytes.NewBuffer(payload))
		if err != nil {
			fmt.Printf("Erro ao enviar os dados, dia %s: %s\n",
				d.Format("2006-01-02"), err.Error())
		}
	}
}

func initialModel() model {
	items := []list.Item{
		item("Usuário e senha"),
		item("Token"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Como deseja realizar o login?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.HelpStyle = helpStyle

	m := model{
		list:     l,
		chosen:   false,
		quitting: false,
	}

	return m
}

func main() {
	m := initialModel()
	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Algo deu errado: %v\n", err)
		os.Exit(1)
	}

	if m.list.Index() == 0 {
		if m.quitting {
			os.Exit(0)
		}
		cookie, err := login(m.inputs[0].Value(), m.inputs[1].Value())
		if err != nil {
			fmt.Printf("Erro ao realizar o login: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Token:", cookie)
	} else if m.list.Index() == 1 {
		fillHours(m.inputs[0].Value(), calendar)
	}
}
