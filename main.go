package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

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

func calculateEaster(year int) time.Time {
	Gauss := map[int][2]int{
		1599: {22, 2},
		1699: {22, 2},
		1799: {23, 3},
		1899: {24, 4},
		1999: {24, 5},
		2099: {24, 5},
		2199: {24, 6},
		2299: {25, 7},
	}
	century := (year/100)*100 + 99
	x := Gauss[century][0]
	y := Gauss[century][1]
	a := year % 19
	b := year % 4
	c := year % 7
	d := (19*a + x) % 30
	e := (2*b + 4*c + 6*d + y) % 7

	if (d + e) > 9 {
		day := d + e - 9
		if day == 26 {
			day = 19
		}
		if day == 25 && d == 28 && a > 10 {
			day = 18
		}
		return time.Date(year, time.April, day, 0, 0, 0, 0, time.UTC)
	} else {
		day := d + e + 22
		return time.Date(year, time.March, day, 0, 0, 0, 0, time.UTC)
	}
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

func fillHours(requestToken string, holidays map[time.Time]bool) {
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
	current_month := current_time.Format("2006-01")

	for day := 1; day <= 31; day++ {
		date_str := fmt.Sprintf("%s-%02d", current_month, day)
		date, err := time.Parse("2006-01-02", date_str)
		if err != nil {
			fmt.Printf("Ignorando data inválida: %s", date)
			continue
		}

		_, ok := holidays[date]
		if ok {
			continue
		}

		data = FillData{
			Id_Proyecto:                projectId,
			Id_Recurso:                 ResourceId,
			FechaDia:                   date_str,
			Comentario:                 "",
			Actividades:                activities,
			pantallaCaptura:            true,
			Latitude:                   0,
			Longitude:                  0,
			__RequestVerificationToken: requestToken,
		}

		payload, _ := json.Marshal(data)

		_, err = http.Post(url, "application/json",
			bytes.NewBuffer(payload))
		if err != nil {
			fmt.Printf("Erro ao enviar os dados, dia %s: %s\n",
				date, err.Error())
		}
	}
}

func initialModel() model {
	return model{
		choices:  []string{"Utilizar usuário e senha", "Utilizar token"},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
        s := "Como deseja utilizar o programa:\n\n"

	for i, choice := range m.choices {
                cursor := " "
                if m.cursor == i {
                        cursor = ">"
                }

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\n Para sair, pressione Ctrl+C ou Q"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Algo deu errado: %v\n", err)
		os.Exit(1)
	}
}

func main2() {
	current_year := time.Now().Year()
	pascoa := calculateEaster(current_year)
	carnival := pascoa.AddDate(0, 0, -47)
	corpus_christi := pascoa.AddDate(0, 0, 60)

	holidays := map[time.Time]bool{
		carnival:                 true,
		corpus_christi:           true,
		pascoa.AddDate(0, 0, -2): true, // Sexta-feira Santa
		time.Date(current_year, time.January, 1, 0, 0, 0, 0, time.UTC):   true, // Ano Novo
		time.Date(current_year, time.April, 21, 0, 0, 0, 0, time.UTC):    true, // Tiradentes
		time.Date(current_year, time.May, 1, 0, 0, 0, 0, time.UTC):       true, // Dia do Trabalho
		time.Date(current_year, time.September, 7, 0, 0, 0, 0, time.UTC): true, // Independência do Brasil
		time.Date(current_year, time.October, 12, 0, 0, 0, 0, time.UTC):  true, // Nossa Senhora Aparecida
		time.Date(current_year, time.November, 2, 0, 0, 0, 0, time.UTC):  true, // Finados
		time.Date(current_year, time.November, 15, 0, 0, 0, 0, time.UTC): true, // Proclamação da República
		time.Date(current_year, time.November, 21, 0, 0, 0, 0, time.UTC): true, // Consciência Negra
		time.Date(current_year, time.December, 25, 0, 0, 0, 0, time.UTC): true, // Natal
	}

	cookie, err := login()
	if err != nil {
		fmt.Print(err)
		os.Exit(4)
	}

	fillHours(cookie, holidays)
}
