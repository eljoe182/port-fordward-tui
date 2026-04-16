package tui

import tea "github.com/charmbracelet/bubbletea"

type KeyMap struct{}

func DefaultKeyMap() KeyMap { return KeyMap{} }

func isTabKey(msg tea.KeyMsg) bool { return msg.Type == tea.KeyTab }
