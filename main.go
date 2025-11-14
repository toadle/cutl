package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"cutl/internal"
	"cutl/internal/tui"
	"cutl/internal/version"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:     "cutl",
	Version: version.GetVersion(),
	Short:   "A cozy tool to sift through and modify JSONL files.",
	Long:    `The main use case is to quickly view and edit large JSONL files in the terminal. The main use-case in mind was the need to manage datasets for machine learning tasks.`,

	Run: func(cmd *cobra.Command, args []string) {
		var debug, _ = cmd.Flags().GetBool("debug")
		var inputPath, _ = cmd.Flags().GetString("input")
		var loggerFile = initDebugLog(debug)
		if loggerFile != nil {
			defer loggerFile.Close()
		}

		if inputPath == "" {
			fmt.Println("Please provide a path to a JSONL file using --input.")
			os.Exit(1)
		}

		var ui *tui.Model = tui.New(inputPath)
		p := tea.NewProgram(ui, tea.WithAltScreen())
		internal.InitMessageRelay(p.Send)

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

func initDebugLog(debug bool) *os.File {
	var loggerFile *os.File

	if debug {
		var fileErr error
		newConfigFile, fileErr := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if fileErr == nil {
			log.SetOutput(newConfigFile)
			log.SetTimeFormat(time.Kitchen)
			log.SetReportCaller(true)
			log.SetLevel(log.DebugLevel)
			log.Info("Logging to debug.log")
		} else {
			loggerFile, _ = tea.LogToFile("debug.log", "debug")
			fmt.Println("Failed setting up logging", fileErr)
		}
	} else {
		// Disable logging entirely when not in debug mode
		log.SetLevel(log.FatalLevel) // Only show fatal errors
		log.SetOutput(os.Stderr)     // Ensure output goes to stderr, not stdout
	}
	return loggerFile
}

func main() {
	cmd.PersistentFlags().Bool("debug", false, "passing this flag will allow writing debug output to debug.log")
	cmd.PersistentFlags().String("input", "", "Pfad zu einer JSONL-Datei, die beim Start geladen wird")
	
	// Custom version template to show full version info
	cmd.SetVersionTemplate(fmt.Sprintf("%s\n", version.GetFullVersion()))
	
	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}
