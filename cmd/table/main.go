package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type User struct {
	Name string `table:"name,color:green,text-align:center"`
}

func main() {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("#5A56E0"))
	s := base.Render

	center := lipgloss.NewStyle().Inherit(base).Align(lipgloss.Center).PaddingLeft(1).PaddingRight(1)

	t := table.New().Border(lipgloss.ASCIIBorder())
	t.Headers(center.Render("LEFT"), "RIGHT")
	t.Row("Bubble Tea", s("Milky"))
	t.Row("Milk Tea", s("Also milky"))
	t.Row("Actual milk", s("Milky as well"))
	// t.Offset(10)
	fmt.Println(t.Render())
}
