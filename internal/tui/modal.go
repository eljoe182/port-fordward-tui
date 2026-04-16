package tui

import (
	"fmt"

	"port-forward-tui/internal/app/catalog"
	"port-forward-tui/internal/domain"

	tea "github.com/charmbracelet/bubbletea"
)

type modalOption struct {
	Label string
	Value string
}

func (m Model) openContextModal() Model {
	if len(m.contexts) == 0 {
		return m
	}
	m.modalKind = ModalContext
	m.modalCursor = indexOf(m.contexts, m.contextName)
	m.modalInput = ""
	return m
}

func (m Model) openNamespaceModal() Model {
	if len(m.namespaces) == 0 {
		return m
	}
	m.modalKind = ModalNamespace
	m.modalCursor = indexOf(m.namespaces, m.namespace)
	m.modalInput = ""
	return m
}

func (m Model) openFilterModal() Model {
	m.modalKind = ModalFilter
	m.modalCursor = indexModalOption(m.filterOptions(), string(m.filterMode))
	m.modalInput = ""
	return m
}

func (m Model) openSortModal() Model {
	m.modalKind = ModalSort
	m.modalCursor = indexModalOption(m.sortOptions(), string(m.sortMode))
	m.modalInput = ""
	return m
}

func (m Model) openSearchModal() Model {
	m.modalKind = ModalSearch
	m.modalCursor = 0
	m.modalInput = m.query
	return m
}

func (m Model) closeModal() Model {
	m.modalKind = ModalNone
	m.modalCursor = 0
	m.modalInput = ""
	return m
}

func (m Model) modalOptions() []modalOption {
	switch m.modalKind {
	case ModalContext:
		options := make([]modalOption, 0, len(m.contexts))
		for _, item := range m.contexts {
			options = append(options, modalOption{Label: item, Value: item})
		}
		return options
	case ModalNamespace:
		options := make([]modalOption, 0, len(m.namespaces))
		for _, item := range m.namespaces {
			options = append(options, modalOption{Label: item, Value: item})
		}
		return options
	case ModalFilter:
		return m.filterOptions()
	case ModalSort:
		return m.sortOptions()
	default:
		return nil
	}
}

func (m Model) filterOptions() []modalOption {
	return []modalOption{
		{Label: "All targets", Value: string(catalog.FilterAll)},
		{Label: "Services only", Value: string(catalog.FilterServices)},
		{Label: "Pods only", Value: string(catalog.FilterPods)},
		{Label: "Favorites only", Value: string(catalog.FilterFavorites)},
	}
}

func (m Model) sortOptions() []modalOption {
	return []modalOption{
		{Label: "Smart", Value: string(catalog.SortSmart)},
		{Label: "Name", Value: string(catalog.SortName)},
		{Label: "Recent", Value: string(catalog.SortRecent)},
		{Label: "Favorites", Value: string(catalog.SortFavorites)},
		{Label: "Type", Value: string(catalog.SortType)},
	}
}

func (m Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.modalKind == ModalSearch {
		return m.handleSearchModalKey(msg)
	}

	options := m.modalOptions()
	if len(options) == 0 {
		return m.closeModal(), nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		return m.closeModal(), nil
	case tea.KeyEnter:
		return m.applyModalSelection(options[m.modalCursor].Value)
	case tea.KeyUp:
		m.modalCursor = maxInt(0, m.modalCursor-1)
		return m, nil
	case tea.KeyDown:
		if m.modalCursor < len(options)-1 {
			m.modalCursor++
		}
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	}

	switch string(msg.Runes) {
	case "j":
		if m.modalCursor < len(options)-1 {
			m.modalCursor++
		}
		return m, nil
	case "k":
		m.modalCursor = maxInt(0, m.modalCursor-1)
		return m, nil
	}

	return m, nil
}

func (m Model) handleSearchModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return m.closeModal(), nil
	case tea.KeyEnter:
		m.query = m.modalInput
		m.queryBuffer = m.modalInput
		m = m.closeModal()
		return m, loadCatalogCmd(m.ctx, m.deps, m.loadOptions())
	case tea.KeyBackspace:
		if len(m.modalInput) > 0 {
			m.modalInput = m.modalInput[:len(m.modalInput)-1]
		}
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	}
	for _, r := range msg.Runes {
		m.modalInput += string(r)
	}
	return m, nil
}

func (m Model) applyModalSelection(value string) (tea.Model, tea.Cmd) {
	switch m.modalKind {
	case ModalContext:
		m = m.closeModal()
		if m.deps.ConfigStore == nil || m.deps.Discovery == nil {
			m.contextName = value
			return m, nil
		}
		return m, saveConfigAndReloadCmd(m.ctx, m.deps, m.loadOptions(), func(cfg *domain.AppConfig) {
			cfg.CurrentContext = value
			cfg.CurrentNamespace = ""
		})
	case ModalNamespace:
		m = m.closeModal()
		if m.deps.ConfigStore == nil || m.deps.Discovery == nil {
			m.namespace = value
			return m, nil
		}
		return m, saveConfigAndReloadCmd(m.ctx, m.deps, m.loadOptions(), func(cfg *domain.AppConfig) {
			cfg.CurrentContext = m.contextName
			cfg.CurrentNamespace = value
		})
	case ModalFilter:
		m.filterMode = catalog.FilterMode(value)
		m = m.closeModal()
		if m.deps.ConfigStore == nil || m.deps.Discovery == nil {
			return m, nil
		}
		return m, loadCatalogCmd(m.ctx, m.deps, m.loadOptions())
	case ModalSort:
		m.sortMode = catalog.SortMode(value)
		m = m.closeModal()
		if m.deps.ConfigStore == nil || m.deps.Discovery == nil {
			return m, nil
		}
		return m, loadCatalogCmd(m.ctx, m.deps, m.loadOptions())
	default:
		return m.closeModal(), nil
	}
}

func indexModalOption(options []modalOption, value string) int {
	for i, option := range options {
		if option.Value == value {
			return i
		}
	}
	return 0
}

func (m Model) renderModal() string {
	switch m.modalKind {
	case ModalSearch:
		return renderSearchModal(m.modalInput)
	case ModalContext, ModalNamespace, ModalFilter, ModalSort:
		return renderSelectorModal(modalTitle(m.modalKind), m.modalOptions(), m.modalCursor)
	default:
		return ""
	}
}

func modalTitle(kind ModalKind) string {
	switch kind {
	case ModalContext:
		return "Select context"
	case ModalNamespace:
		return "Select namespace"
	case ModalFilter:
		return "Select filter"
	case ModalSort:
		return "Select sort"
	case ModalSearch:
		return "Search catalog"
	default:
		return "Selector"
	}
}

func renderSearchModal(input string) string {
	return fmt.Sprintf("Search catalog\n\nquery: [%s_]\n\nEnter apply • Esc cancel", input)
}
