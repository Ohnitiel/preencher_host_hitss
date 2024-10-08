package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	p *tea.Program

	progressHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#353535"))
	progressPadding   = 2
	progressMaxWidth  = 100
)

type progressWriter struct {
	total      float64
	current    float64
	currentDay string
	onProgress func(float64)
}

func (pw *progressWriter) Start(
	headerRequestToken string, requestToken string, calendar Calendar,
	startDate time.Time, endDate time.Time,
) {
	var data FillData

	activities := make([]Activities, 1)

	activities[0] = Activities{
		Id_CapturaActividad: "0",
		Id_Actividad:        "973217",
		HorasCapturadas:     "8.0",
		Comentario:          "",
		HorasExtras:         false,
		HorasNocturnas:      false,
		Bloqueada:           false,
	}

	for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
		if !calendar[d] {
			continue
		}
		pw.currentDay = d.Format("02/01/2006")

		data = FillData{
			Id_Proyecto:              projectId,
			Id_Recurso:               ResourceId,
			FechaDia:                 d.Format("2006-01-02"),
			Comentario:               "",
			Actividades:              activities,
			PantallaCaptura:          true,
			Latitude:                 0,
			Longitude:                0,
			RequestVerificationToken: requestToken,
		}

		json_data, _ := json.Marshal(data)
		payload := bytes.NewBuffer([]byte(json_data))

		req, err := http.NewRequest("POST", activities_url, payload)
		if err != nil {
			fmt.Printf("Erro ao criar request para url: %s.\nErro: %e",
				activities_url, err)
			os.Exit(3)
		}

		req.Header.Add("__requestverificationtoken", requestToken)
		req.Header.Add("content-type", "application/json; charset=UTF-8")
		req.Header.Add("cookie", "HOST=Cultura=pt&Auxiliar=;")
		req.Header.Add("cookie", fmt.Sprintf("__RequestVerificationToken=%s;", headerRequestToken))
		req.Header.Add("cookie", fmt.Sprintf(".ASPXAUTH=%s", aspx))

		response, err := client.Do(req)
		if err != nil {
			fmt.Printf("Erro ao enviar os dados, dia %s: %s\n",
				d.Format("2006-01-02"), err.Error())
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			fmt.Printf("Falha ao enviar os dados, dia %s. Status: %d\n",
				d.Format("02/01/2006"), response.StatusCode)
		}

		pw.onProgress(float64(d.Sub(startDate).Hours() / 24 / pw.total))
	}

	pw.onProgress(float64(1))
}

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg {
		return nil
	})
}

type progressMsg float64

type progressModel struct {
	pw       *progressWriter
	progress progress.Model
}

func (m progressModel) Init() tea.Cmd {
	return nil
}

func (m progressModel) View() string {
	pad := strings.Repeat(" ", progressPadding)
	return "\n" +
		pad + "Preenchendo dia: " + m.pw.currentDay + "\n\n" +
		pad + m.progress.View() + "\n\n" +
		pad + progressHelpStyle.Render("Press ctrl+c to quit")
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - progressPadding*2 - 4
		if m.progress.Width > progressMaxWidth {
			m.progress.Width = progressMaxWidth
		}
		return m, nil

	case progressMsg:
		var cmds []tea.Cmd

		if msg >= 1.0 {
			cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
		}

		cmds = append(cmds, m.progress.SetPercent(float64(msg)))
		return m, tea.Batch(cmds...)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
	return m, nil
}
