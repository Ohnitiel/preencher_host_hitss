package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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

func login() (string, error) {
	login_url := "https://host.globalhitss.com/"
	response, err := http.PostForm(
		login_url,
		url.Values{
			"UserName": {},
			"Password": {},
			"Language": {"pt"},
			"bandera":  {"1"},
		},
	)
	if err != nil {
		fmt.Println("Erro ao realizar o login.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cookies := response.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "__RequestVerificationToken" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("Não foi possível encontrar o token")
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
		list:   l,
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model

	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 255

		switch i {
		case 0:
			t.Placeholder = "Username"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = -1
		}

		m.inputs[i] = t
	}

	return m
}

func main() {
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Algo deu errado: %v\n", err)
		os.Exit(1)
	}
}

func main2() {
	cookie, err := login()
	if err != nil {
		fmt.Print(err)
		os.Exit(4)
	}

	fillHours(cookie, CalendarForYear(time.Now().Year()))
}
