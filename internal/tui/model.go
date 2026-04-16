package tui

type Tab string

const TabSelected Tab = "selected"

type Dependencies struct{}

type Model struct {
	activeTab Tab
	catalog   []string
}

func NewModel(_ Dependencies) Model {
	return Model{activeTab: TabSelected, catalog: []string{}}
}
