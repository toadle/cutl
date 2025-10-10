package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	normal          = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#dddddd"}
	normalDim       = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}
	gray            = lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"}
	midGray         = lipgloss.AdaptiveColor{Light: "#B2B2B2", Dark: "#4A4A4A"}
	dimGray         = lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}
	darkGray        = lipgloss.AdaptiveColor{Light: "#585858", Dark: "#585858"}
	brightGray      = lipgloss.AdaptiveColor{Light: "#847A85", Dark: "#979797"}
	dimBrightGray   = lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"}
	indigo          = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	dimIndigo       = lipgloss.AdaptiveColor{Light: "#9498FF", Dark: "#494690"}
	subtleIndigo    = lipgloss.AdaptiveColor{Light: "#7D79F6", Dark: "#514DC1"}
	dimSubtleIndigo = lipgloss.AdaptiveColor{Light: "#BBBDFF", Dark: "#383584"}
	cream           = lipgloss.AdaptiveColor{Light: "#FFFDF5", Dark: "#FFFDF5"}
	yellowGreen     = lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#ECFD65"}
	dullYellowGreen = lipgloss.AdaptiveColor{Light: "#6BCB94", Dark: "#9BA92F"}
	fuchsia         = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}
	dimFuchsia      = lipgloss.AdaptiveColor{Light: "#F1A8FF", Dark: "#99519E"}
	dullFuchsia     = lipgloss.AdaptiveColor{Dark: "#AD58B4", Light: "#F793FF"}
	dimDullFuchsia  = lipgloss.AdaptiveColor{Light: "#F6C9FF", Dark: "#6B3A6F"}
	red             = lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#ED567A"}
	faintRed        = lipgloss.AdaptiveColor{Light: "#FF6F91", Dark: "#C74665"}

	green        = lipgloss.AdaptiveColor{Light: "#2BCB79", Dark: "#2BCB79"}
	semiDimGreen = lipgloss.AdaptiveColor{Light: "#35D79C", Dark: "#036B46"}
	dimGreen     = lipgloss.AdaptiveColor{Light: "#72D2B0", Dark: "#0B5137"}
)

var (
	App  = lipgloss.NewStyle().Padding(2, 4)
	Logo = lipgloss.NewStyle()

	Label         = lipgloss.NewStyle().Foreground(indigo)
	Header        = lipgloss.NewStyle().MarginBottom(1)
	Results       = lipgloss.NewStyle().MarginBottom(1)
	CommandPanel  = lipgloss.NewStyle().MarginBottom(1)
	ProgressPanel = lipgloss.NewStyle().MarginBottom(1)
	DetailPanel   = lipgloss.NewStyle().MarginBottom(1).Border(lipgloss.NormalBorder(), true).BorderForeground(darkGray).BorderBottom(false).BorderLeft(false).BorderRight(false)

	CommandTitle                 = Label.Background(dullFuchsia).Foreground(cream).Padding(0, 1).MarginRight(2)
	CommandLabel                 = Label.Foreground(dullFuchsia).MarginRight(2)
	CommandLabelTrigger          = Label.Foreground(fuchsia).Bold(true)
	CommandSelectionLabel        = CommandLabel.Foreground(indigo)
	CommandSelectionLabelTrigger = CommandLabelTrigger.Foreground(indigo).Bold(true)
	CommandMeta                  = Label.Foreground(midGray).MarginLeft(2)
	CommandStatus                = Label.Foreground(semiDimGreen).MarginTop(1)

	Text      = lipgloss.NewStyle().Foreground(normal)
	InfoLabel = Label.Foreground(darkGray)
	OkLabel   = Label.Foreground(green)
	NoLabel   = Label.Foreground(faintRed)
)
