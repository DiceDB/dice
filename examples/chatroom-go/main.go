// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package main

import (
	"chatroom-go/svc"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dicedb/dicedb-go/wire"
)

const gap = "\n\n"

var (
	username string
)

var M model

func init() {
	M = initialModel()
}

func main() {
	log.Println("Chatroom application started.")

	if len(os.Args) < 2 {
		fmt.Println("go run main.go <username>")
		os.Exit(1)
	}

	username = os.Args[1]
	if username == "" {
		fmt.Println("go run main.go <username>")
		os.Exit(1)
	}

	go svc.ListenForMessages(func(result *wire.Result) {
		M.AddMessage(result.GetGETRes().Value)
		M.Refresh()
	})

	p := tea.NewProgram(&M)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message ..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)
	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m *model) AddMessage(message string) {
	tokens := strings.Split(message, ":")
	username := tokens[0]
	msg := strings.Join(tokens[1:], ":")
	m.messages = append(m.messages, m.senderStyle.Render(username+": ")+msg)
}

func (m *model) Refresh() {
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
	m.viewport.GotoBottom()
}

func (m *model) Init() tea.Cmd {
	svc.Subscribe()
	return textarea.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
		if len(m.messages) > 0 {
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoTop()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			svc.SendMessage(username, m.textarea.Value())
			m.textarea.Reset()
			m.Refresh()
		}
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}
