package tui

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonsth131/ctfd-cli/api"
	"github.com/jonsth131/ctfd-cli/tui/constants"
)

func StartTea(url string, logging bool) {
	if logging {
		if f, err := tea.LogToFile("debug.log", "ctfd-cli"); err != nil {
			fmt.Println("Couldn't open a file for logging:", err)
			os.Exit(1)
		} else {
			defer func() {
				err = f.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()
		}
	} else {
		log.SetOutput(io.Discard)
	}

	client, err := api.NewApiClient(url)
	if err != nil {
		fmt.Println("Failed to create Api Client")
		log.Fatal(err)
	}

	constants.C = client

	m, _ := InitLogin()
	constants.P = tea.NewProgram(m, tea.WithAltScreen())
	if _, err := constants.P.Run(); err != nil {
		log.Fatal(err)
	}
}
