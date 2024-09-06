package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	base_url       = "https://host.globalhitss.com"
	login_url      = base_url + "/Security/Login"
	cookie_url     = base_url + "/Horas/CapturaHoras2"
	activities_url = base_url + "/Horas/ActualizaActividades2"

	projectId  = 73556
	ResourceId = "50900135"
)

var calendar = CalendarForYear(time.Now().Year())

// Http Request related
var (
	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	aspx string
)

type Activities struct {
	Id_CapturaActividad string `json:"Id_CapturaActividad"`
	Id_Actividad        string `json:"Id_Actividad"`
	HorasCapturadas     string `json:"HorasCapturadas"`
	Comentario          string `json:"Comentario"`
	HorasExtras         bool   `json:"HorasExtras"`
	HorasNocturnas      bool   `json:"HorasNocturnas"`
	Bloqueada           bool   `json:"Bloqueada"`
}

type FillData struct {
	Id_Proyecto              int          `json:"Id_Proyecto"`
	Id_Recurso               string       `json:"Id_Recurso"`
	FechaDia                 string       `json:"FechaDia"`
	Comentario               string       `json:"Comentario"`
	Actividades              []Activities `json:"Actividades"`
	PantallaCaptura          bool         `json:"pantallaCaptura"`
	Latitude                 float64      `json:"Latitude"`
	Longitude                float64      `json:"Longitude"`
	RequestVerificationToken string       `json:"__RequestVerificationToken"`
}

func login(username string, password string) (string, string, error) {
	re := regexp.MustCompile(`__RequestVerificationToken".*?value="(.*?)"`)

	body := []byte(fmt.Sprintf(
		"UserName=%s&Password=%s&Language=pt&bandera=1", username, password,
	))
	payload := bytes.NewBuffer(body)

	req, err := http.NewRequest("POST", base_url, payload)
	if err != nil {
		fmt.Printf("Erro ao criar request para url: %s.\nErro: %e",
			base_url, err)
		os.Exit(3)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Erro ao executar o login. %e", err)
		os.Exit(5)
	}
	if response.StatusCode != 302 {
		fmt.Printf("Falha na autenticação. Status: %d\n", response.StatusCode)
		os.Exit(401)
	}

	cookies := response.Cookies()
	for _, c := range cookies {
		if c.Name == ".ASPXAUTH" {
			aspx = c.Value
		}
	}

	req, err = http.NewRequest("GET", cookie_url, nil)
	if err != nil {
		fmt.Printf("Erro ao criar request para url: %s.\nErro: %e",
			cookie_url, err)
		os.Exit(3)
	}
	req.Header.Add("Cookie", fmt.Sprintf(".ASPXAUTH=%s", aspx))
	req.Header.Add("Cookie", "HOST=Cultura=pt&Auxiliar=")
	req.Header.Add("Referer", login_url)

	response, err = client.Do(req)
	if err != nil {
		fmt.Printf("Erro ao coletar cookie. %e", err)
		os.Exit(5)
	}

	text, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Erro ao ler a resposta. %e", err)
		os.Exit(7)
	}
	response.Body.Close()
	token := re.FindSubmatch(text)

	cookies = response.Cookies()
	for _, c := range cookies {
		if c.Name == "__RequestVerificationToken" {
			fmt.Println("Login efetuado com sucesso!", "Usuário: ", username)
			return c.Value, string(token[1]), nil
		}
	}

	return "", "", fmt.Errorf("Cookie não encontrado.")
}

func initialModel() model {
	items := []list.Item{
		item("Usuário e senha"),
		// item("Token"),
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
	var cookie, token string
	var err error

	m := initialModel()
	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Algo deu errado: %v\n", err)
		os.Exit(1)
	}

	if m.quitting {
		os.Exit(0)
	}

	if m.list.Index() == 0 {
		cookie, token, err = login(m.inputs[0].Value(), m.inputs[1].Value())
		if err != nil {
			fmt.Printf("Erro ao realizar o login: %v\n", err)
			os.Exit(1)
		}
	} else if m.list.Index() == 1 {
		cookie, token, aspx = m.inputs[0].Value(), m.inputs[1].Value(), m.inputs[2].Value()
	}

	current_time := time.Now()
	startDate := time.Date(current_time.Year(), current_time.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	pw := &progressWriter{
		total: endDate.Sub(startDate).Hours() / 24,
		onProgress: func(ratio float64) {
			p.Send(progressMsg(ratio))
		},
	}

	progress := progressModel{
		pw:       pw,
		progress: progress.New(progress.WithDefaultGradient()),
	}

	go pw.Start(cookie, token, calendar, startDate, endDate)
	os.Exit(0)
}
