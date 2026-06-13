package tui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---

var (
	styleNormal    = lipgloss.NewStyle()
	styleSelected  = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Bold(true)
	styleDir       = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Bold(true)
	styleExec      = lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
	styleSymlink   = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	styleHidden    = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	styleHeader    = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	styleStatusBar = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252")).Padding(0, 1)
	styleBreadcrumb = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	styleError     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	styleDim = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
)

// --- Keybindings ---

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Home     key.Binding
	End      key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	ToggleHidden key.Binding
	Delete   key.Binding
	Quit     key.Binding
	Help     key.Binding
}

var keys = keyMap{
	Up:           key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:         key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:        key.NewBinding(key.WithKeys("enter", "l"), key.WithHelp("enter/l", "open")),
	Back:         key.NewBinding(key.WithKeys("backspace", "h"), key.WithHelp("bspc/h", "back")),
	Home:         key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("home/g", "top")),
	End:          key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("end/G", "bottom")),
	PageUp:       key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("pgup/^u", "page up")),
	PageDown:     key.NewBinding(key.WithKeys("pgdown", "ctrl+d"), key.WithHelp("pgdn/^d", "page down")),
	ToggleHidden: key.NewBinding(key.WithKeys("."), key.WithHelp(".", "toggle hidden")),
	Delete:       key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:         key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}

// --- Messages ---

type dirLoadedMsg struct {
	path    string
	entries []fs.DirEntry
	err     error
}

type errMsg struct{ err error }

// --- Model ---

type Model struct {
	path        string
	entries     []fs.DirEntry
	cursor      int
	offset      int // scroll offset
	width       int
	height      int
	showHidden  bool
	showHelp    bool
	statusMsg   string
	err         error
}

func New(startPath string) Model {
	if startPath == "" {
		startPath, _ = os.Getwd()
	}
	return Model{
		path:       startPath,
		showHidden: false,
	}
}

func (m Model) Init() tea.Cmd {
	return loadDir(m.path)
}

// --- Update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case dirLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.path = msg.path
		m.entries = filterEntries(msg.entries, m.showHidden)
		m.cursor = 0
		m.offset = 0
		m.err = nil
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		m.err = nil
		m.statusMsg = ""

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, keys.ToggleHidden):
			m.showHidden = !m.showHidden
			return m, loadDir(m.path)

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset--
				}
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.entries)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.listHeight() {
					m.offset++
				}
			}

		case key.Matches(msg, keys.Home):
			m.cursor = 0
			m.offset = 0

		case key.Matches(msg, keys.End):
			m.cursor = len(m.entries) - 1
			m.offset = max(0, len(m.entries)-m.listHeight())

		case key.Matches(msg, keys.PageUp):
			m.cursor = max(0, m.cursor-m.listHeight())
			m.offset = max(0, m.offset-m.listHeight())

		case key.Matches(msg, keys.PageDown):
			m.cursor = min(len(m.entries)-1, m.cursor+m.listHeight())
			m.offset = min(max(0, len(m.entries)-m.listHeight()), m.offset+m.listHeight())

		case key.Matches(msg, keys.Enter):
			return m, m.openSelected()

		case key.Matches(msg, keys.Back):
			parent := filepath.Dir(m.path)
			if parent != m.path {
				return m, loadDir(parent)
			}

		case key.Matches(msg, keys.Delete):
			return m, m.deleteSelected()
		}
	}

	return m, nil
}

// --- View ---

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Breadcrumb
	b.WriteString(styleBreadcrumb.Render(" 󰉋  " + m.path))
	b.WriteString("\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	if m.showHelp {
		b.WriteString(m.helpView())
	} else {
		b.WriteString(m.listView())
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.statusBar())

	return b.String()
}

func (m Model) listView() string {
	if len(m.entries) == 0 {
		return styleDim.Render("  (empty directory)\n")
	}

	var b strings.Builder
	visible := m.listHeight()
	end := min(m.offset+visible, len(m.entries))

	for i := m.offset; i < end; i++ {
		entry := m.entries[i]
		line := m.renderEntry(entry, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) renderEntry(entry fs.DirEntry, selected bool) string {
	info, _ := entry.Info()
	name := entry.Name()
	hidden := strings.HasPrefix(name, ".")

	var icon, styledName string

	switch {
	case entry.IsDir():
		icon = " "
		styledName = styleDir.Render(name)
	case entry.Type()&fs.ModeSymlink != 0:
		icon = " "
		styledName = styleSymlink.Render(name)
	case info != nil && info.Mode()&0111 != 0:
		icon = " "
		styledName = styleExec.Render(name)
	default:
		icon = " "
		styledName = styleNormal.Render(name)
	}

	if hidden {
		styledName = styleHidden.Render(name)
	}

	size := ""
	if info != nil && !entry.IsDir() {
		size = styleDim.Render(humanizeSize(info.Size()))
	}

	prefix := "  "
	if selected {
		prefix = styleSelected.Render(" ❯ ")
		line := fmt.Sprintf("%s%s%s", prefix, icon, styledName)
		if size != "" {
			pad := m.width - lipgloss.Width(line) - lipgloss.Width(size) - 2
			if pad > 0 {
				line += strings.Repeat(" ", pad) + size
			}
		}
		return styleSelected.Render(" ❯ ") + icon + styledName +
			strings.Repeat(" ", max(0, m.width-lipgloss.Width(prefix+icon+styledName+size)-2)) + size
	}

	line := fmt.Sprintf("%s%s%s", prefix, icon, styledName)
	if size != "" {
		pad := m.width - lipgloss.Width(line) - lipgloss.Width(size) - 2
		if pad > 0 {
			line += strings.Repeat(" ", pad) + size
		}
	}
	return line
}

func (m Model) statusBar() string {
	if m.err != nil {
		return styleError.Render(" ✗ " + m.err.Error())
	}

	total := len(m.entries)
	pos := 0
	if total > 0 {
		pos = m.cursor + 1
	}

	left := fmt.Sprintf(" %d/%d", pos, total)
	hidden := ""
	if m.showHidden {
		hidden = "  show hidden"
	}
	right := fmt.Sprintf("%s  ? help  q quit ", hidden)

	if m.statusMsg != "" {
		left = " " + m.statusMsg
	}

	pad := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return styleStatusBar.Render(left + strings.Repeat(" ", pad) + right)
}

func (m Model) helpView() string {
	bindings := []key.Binding{
		keys.Up, keys.Down, keys.Enter, keys.Back,
		keys.Home, keys.End, keys.PageUp, keys.PageDown,
		keys.ToggleHidden, keys.Delete, keys.Help, keys.Quit,
	}
	var b strings.Builder
	b.WriteString(styleHeader.Render("\n  Keybindings\n\n"))
	for _, kb := range bindings {
		fmt.Fprintf(&b, "  %-18s %s\n",
			styleSelected.Render(kb.Help().Key),
			kb.Help().Desc,
		)
	}
	return b.String()
}

// --- Commands ---

func loadDir(path string) tea.Cmd {
	return func() tea.Msg {
		abs, err := filepath.Abs(path)
		if err != nil {
			return dirLoadedMsg{err: err}
		}
		entries, err := os.ReadDir(abs)
		return dirLoadedMsg{path: abs, entries: entries, err: err}
	}
}

func (m Model) openSelected() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.cursor]
	if entry.IsDir() {
		return loadDir(filepath.Join(m.path, entry.Name()))
	}
	return nil
}

func (m Model) deleteSelected() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.cursor]
	target := filepath.Join(m.path, entry.Name())
	return func() tea.Msg {
		if err := os.RemoveAll(target); err != nil {
			return errMsg{err: fmt.Errorf("delete failed: %w", err)}
		}
		return loadDir(m.path)()
	}
}

// --- Helpers ---

func filterEntries(entries []fs.DirEntry, showHidden bool) []fs.DirEntry {
	if showHidden {
		return entries
	}
	filtered := entries[:0:len(entries)]
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func (m Model) listHeight() int {
	// breadcrumb(1) + divider(1) + statusbar(1) + newline(1) = 4
	h := m.height - 4
	if h < 1 {
		return 1
	}
	return h
}

func humanizeSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

