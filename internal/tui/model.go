package tui

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	chromaLexers "github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	srch "github.com/The-True-Hooha/Bolt/internal/utils/search"
)

var (
	clrAccent   = lipgloss.AdaptiveColor{Light: "63", Dark: "69"}
	clrYellow   = lipgloss.AdaptiveColor{Light: "136", Dark: "214"}
	clrGreen    = lipgloss.AdaptiveColor{Light: "34", Dark: "76"}
	clrCyan     = lipgloss.AdaptiveColor{Light: "32", Dark: "51"}
	clrMagenta  = lipgloss.AdaptiveColor{Light: "125", Dark: "170"}
	clrRed      = lipgloss.AdaptiveColor{Light: "160", Dark: "196"}
	clrOrange   = lipgloss.AdaptiveColor{Light: "130", Dark: "208"}
	clrMuted    = lipgloss.AdaptiveColor{Light: "245", Dark: "242"}
	clrFg       = lipgloss.AdaptiveColor{Light: "235", Dark: "252"}
	clrSelBg    = lipgloss.AdaptiveColor{Light: "63", Dark: "62"}
	clrSelFg    = lipgloss.AdaptiveColor{Light: "255", Dark: "230"}
	clrMultiBg  = lipgloss.AdaptiveColor{Light: "54", Dark: "54"}
	clrBorder   = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}
	clrHdrBg    = lipgloss.AdaptiveColor{Light: "254", Dark: "236"}
	clrTitleBg  = lipgloss.AdaptiveColor{Light: "63", Dark: "62"}
	clrClipCopy = lipgloss.AdaptiveColor{Light: "34", Dark: "76"}
	clrClipCut  = lipgloss.AdaptiveColor{Light: "160", Dark: "196"}
)

var (
	sTitle = lipgloss.NewStyle().
		Background(clrTitleBg).Foreground(lipgloss.Color("16")).
		Bold(true).Padding(0, 2)

	sPath = lipgloss.NewStyle().
		Foreground(clrYellow).Bold(true).Padding(0, 1)

	sPane = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(clrBorder)

	sPaneActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrAccent)

	sHeader = lipgloss.NewStyle().
		Background(clrHdrBg).Foreground(clrMuted).Bold(true).Padding(0, 1)

	sSelected = lipgloss.NewStyle().
			Background(clrSelBg).Foreground(clrSelFg).Bold(true)

	sMultiSelected = lipgloss.NewStyle().
			Background(clrMultiBg).Foreground(lipgloss.Color("219"))

	// file type styles
	sDir     = lipgloss.NewStyle().Foreground(clrAccent).Bold(true)
	sExec    = lipgloss.NewStyle().Foreground(clrGreen).Bold(true)
	sSymlink = lipgloss.NewStyle().Foreground(clrCyan)
	sHidden  = lipgloss.NewStyle().Foreground(clrMuted)
	sNormal  = lipgloss.NewStyle().Foreground(clrFg)

	// extension-specific styles
	sGo       = lipgloss.NewStyle().Foreground(clrCyan)
	sJS       = lipgloss.NewStyle().Foreground(clrYellow)
	sPython   = lipgloss.NewStyle().Foreground(clrGreen)
	sRust     = lipgloss.NewStyle().Foreground(clrOrange)
	sMarkdown = lipgloss.NewStyle().Foreground(lipgloss.Color("189"))
	sImage    = lipgloss.NewStyle().Foreground(clrMagenta)
	sArchive  = lipgloss.NewStyle().Foreground(clrRed)
	sConfig   = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
	sMedia    = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))

	sMuted = lipgloss.NewStyle().Foreground(clrMuted)
	sSize  = lipgloss.NewStyle().Foreground(clrCyan)
	sDate  = lipgloss.NewStyle().Foreground(clrMuted)
	sPerms = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))

	sStatusBar = lipgloss.NewStyle().
			Background(clrHdrBg).Foreground(clrFg).Padding(0, 1)

	sKey   = lipgloss.NewStyle().Background(clrTitleBg).Foreground(lipgloss.Color("16")).Bold(true).Padding(0, 1)
	sError = lipgloss.NewStyle().Foreground(clrRed).Bold(true)
	sOk    = lipgloss.NewStyle().Foreground(clrGreen).Bold(true)
	sInfo  = lipgloss.NewStyle().Foreground(clrCyan)

	sFilterBar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(clrAccent).
			Padding(0, 1)

	sSearchBar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(clrGreen).
			Padding(0, 1)

	sSearchResult = lipgloss.NewStyle().Foreground(clrFg)
	sSearchCursor = lipgloss.NewStyle().Background(clrSelBg).Foreground(clrSelFg).Bold(true)

	sDiskBar = lipgloss.NewStyle().Foreground(clrMuted)
)

type sortMode int

const (
	sortName sortMode = iota
	sortSize
	sortDate
	sortType
)

func (s sortMode) String() string {
	return [...]string{"name", "size", "date", "type"}[s]
}

type clipMode int

const (
	clipNone clipMode = iota
	clipCopy
	clipCut
)

type clipboard struct {
	mode  clipMode
	paths []string
}

type inputMode int

const (
	modeNormal inputMode = iota
	modeFilter
	modeRename
	modeNewDir
	modeNewFile
	modeSearch
	modeCommand
)

type keyMap struct {
	Up           key.Binding
	Down         key.Binding
	Enter        key.Binding
	Back         key.Binding
	Home         key.Binding
	End          key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	GotoHome     key.Binding
	ToggleHidden key.Binding
	Filter       key.Binding
	Select       key.Binding
	SelectAll    key.Binding
	SortCycle    key.Binding
	Copy         key.Binding
	Cut          key.Binding
	Paste        key.Binding
	Delete       key.Binding
	Rename       key.Binding
	NewDir       key.Binding
	NewFile      key.Binding
	CopyPath     key.Binding
	NavBack      key.Binding
	NavForward   key.Binding
	Quit         key.Binding
	Help         key.Binding
	Search       key.Binding
	Command      key.Binding
}

var keys = keyMap{
	Up:           key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:         key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:        key.NewBinding(key.WithKeys("enter", "l"), key.WithHelp("↵/l", "open")),
	Back:         key.NewBinding(key.WithKeys("backspace", "h"), key.WithHelp("⌫/h", "back")),
	Home:         key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("home/g", "top")),
	End:          key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("end/G", "bottom")),
	PageUp:       key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("^u", "pg up")),
	PageDown:     key.NewBinding(key.WithKeys("pgdown", "ctrl+d"), key.WithHelp("^d", "pg dn")),
	GotoHome:     key.NewBinding(key.WithKeys("~"), key.WithHelp("~", "home dir")),
	ToggleHidden: key.NewBinding(key.WithKeys("."), key.WithHelp(".", "hidden")),
	Filter:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Select:       key.NewBinding(key.WithKeys(" "), key.WithHelp("spc", "select")),
	SelectAll:    key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("^a", "select all")),
	SortCycle:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort")),
	Copy:         key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
	Cut:          key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "cut")),
	Paste:        key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "paste")),
	Delete:       key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "delete")),
	Rename:       key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
	NewDir:       key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mkdir")),
	NewFile:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new file")),
	CopyPath:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yank path")),
	NavBack:      key.NewBinding(key.WithKeys("["), key.WithHelp("[", "nav back")),
	NavForward:   key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "nav fwd")),
	Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:         key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Search:       key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("^f", "search")),
	Command:      key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "command")),
}

type dirLoadedMsg struct {
	path    string
	entries []fs.DirEntry
	err     error
}
type previewLoadedMsg struct {
	content string
	isDir   bool
	isImage bool
	entries []fs.DirEntry
}
type diskUsageMsg struct{ free, total uint64 }
type errMsg struct{ err error }
type okMsg struct{ msg string }
type cmdOutputMsg struct{ out string }

type searchResultMsg struct {
	results []searchResult
	done    bool
}

type searchResult struct {
	path         string
	name         string
	isDir        bool
	contentMatch bool
}

type Model struct {
	path         string
	allEntries   []fs.DirEntry // unfiltered
	entries      []fs.DirEntry // after filter + sort
	cursor       int
	offset       int
	width        int
	height       int
	showHidden   bool
	showHelp     bool
	sort         sortMode
	selected     map[int]bool
	clip         clipboard
	preview      string
	previewDir   []fs.DirEntry
	previewIsDir bool
	previewIsImg bool
	diskFree     uint64
	diskTotal    uint64
	navHistory   []string
	navIdx       int
	filterQuery  string
	inputMode    inputMode
	inputValue   string
	inputPrompt  string
	statusMsg    string
	statusOk     bool
	err          error
	cmdOutput    string
	cmdInput     string

	searchQuery   string
	searchResults []searchResult
	searchCursor  int
	searchDone    bool
}

func New(startPath string) Model {
	if startPath == "" {
		startPath, _ = os.Getwd()
	}
	abs, _ := filepath.Abs(startPath)
	return Model{
		path:       abs,
		selected:   make(map[int]bool),
		navHistory: []string{abs},
		navIdx:     0,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadDir(m.path), loadDiskUsage(m.path))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case dirLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.path = msg.path
		m.allEntries = msg.entries
		m.entries = m.applyFilterSort()
		m.cursor, m.offset = 0, 0
		m.selected = make(map[int]bool)
		m.err = nil
		return m, tea.Batch(m.loadPreviewCmd(), loadDiskUsage(m.path))

	case previewLoadedMsg:
		m.preview = msg.content
		m.previewIsDir = msg.isDir
		m.previewIsImg = msg.isImage
		m.previewDir = msg.entries
		return m, nil

	case diskUsageMsg:
		m.diskFree, m.diskTotal = msg.free, msg.total
		return m, nil

	case okMsg:
		m.statusMsg, m.statusOk = msg.msg, true
		return m, nil

	case cmdOutputMsg:
		m.cmdOutput = msg.out
		m.inputMode = modeNormal
		if msg.out == "" {
			m.cmdInput = ""
		}
		return m, loadDir(m.path)

	case errMsg:
		m.err = msg.err
		return m, nil

	case searchResultMsg:
		m.searchResults = append(m.searchResults, msg.results...)
		m.searchDone = msg.done
		return m, nil

	case tea.KeyMsg:
		if m.inputMode == modeFilter {
			return m.handleFilterInput(msg)
		}
		if m.inputMode == modeSearch {
			return m.handleSearchInput(msg)
		}
		if m.inputMode == modeCommand {
			return m.handleCommandInput(msg)
		}
		if m.inputMode != modeNormal {
			return m.handleInput(msg)
		}
		m.err, m.statusMsg = nil, ""
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Help):
		m.showHelp = !m.showHelp

	case key.Matches(msg, keys.ToggleHidden):
		m.showHidden = !m.showHidden
		m.entries = m.applyFilterSort()
		m.clampCursor()

	case key.Matches(msg, keys.Filter):
		m.inputMode = modeFilter
		m.inputPrompt = "filter:"
		return m, nil

	case key.Matches(msg, keys.SortCycle):
		m.sort = (m.sort + 1) % 4
		m.entries = m.applyFilterSort()
		m.clampCursor()
		m.statusMsg = "sort: " + m.sort.String()
		m.statusOk = true

	case key.Matches(msg, keys.GotoHome):
		home, err := os.UserHomeDir()
		if err == nil {
			return m, m.navigate(home)
		}

	case key.Matches(msg, keys.NavBack):
		if m.navIdx > 0 {
			m.navIdx--
			return m, loadDir(m.navHistory[m.navIdx])
		}

	case key.Matches(msg, keys.NavForward):
		if m.navIdx < len(m.navHistory)-1 {
			m.navIdx++
			return m, loadDir(m.navHistory[m.navIdx])
		}

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.offset {
				m.offset--
			}
			return m, m.loadPreviewCmd()
		}

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.entries)-1 {
			m.cursor++
			if m.cursor >= m.offset+m.listHeight() {
				m.offset++
			}
			return m, m.loadPreviewCmd()
		}

	case key.Matches(msg, keys.Home):
		m.cursor, m.offset = 0, 0
		return m, m.loadPreviewCmd()

	case key.Matches(msg, keys.End):
		m.cursor = len(m.entries) - 1
		m.offset = max(0, len(m.entries)-m.listHeight())
		return m, m.loadPreviewCmd()

	case key.Matches(msg, keys.PageUp):
		m.cursor = max(0, m.cursor-m.listHeight())
		m.offset = max(0, m.offset-m.listHeight())
		return m, m.loadPreviewCmd()

	case key.Matches(msg, keys.PageDown):
		m.cursor = min(len(m.entries)-1, m.cursor+m.listHeight())
		m.offset = min(max(0, len(m.entries)-m.listHeight()), m.offset+m.listHeight())
		return m, m.loadPreviewCmd()

	case key.Matches(msg, keys.Enter):
		return m, m.openSelected()

	case key.Matches(msg, keys.Back):
		parent := filepath.Dir(m.path)
		if parent != m.path {
			return m, m.navigate(parent)
		}

	case key.Matches(msg, keys.Select):
		if len(m.entries) > 0 {
			m.selected[m.cursor] = !m.selected[m.cursor]
			if !m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			}
		}

	case key.Matches(msg, keys.SelectAll):
		if len(m.selected) == len(m.entries) {
			m.selected = make(map[int]bool)
		} else {
			for i := range m.entries {
				m.selected[i] = true
			}
		}

	case key.Matches(msg, keys.Copy):
		m.clip = m.buildClipboard(clipCopy)
		m.statusMsg = fmt.Sprintf("copied %d item(s)", len(m.clip.paths))
		m.statusOk = true

	case key.Matches(msg, keys.Cut):
		m.clip = m.buildClipboard(clipCut)
		m.statusMsg = fmt.Sprintf("cut %d item(s)", len(m.clip.paths))
		m.statusOk = true

	case key.Matches(msg, keys.Paste):
		return m, m.pasteCmd()

	case key.Matches(msg, keys.Delete):
		return m, m.deleteCmd()

	case key.Matches(msg, keys.Rename):
		if len(m.entries) > 0 {
			m.inputMode = modeRename
			m.inputValue = m.entries[m.cursor].Name()
			m.inputPrompt = "rename:"
		}

	case key.Matches(msg, keys.NewDir):
		m.inputMode = modeNewDir
		m.inputValue = ""
		m.inputPrompt = "new dir:"

	case key.Matches(msg, keys.NewFile):
		m.inputMode = modeNewFile
		m.inputValue = ""
		m.inputPrompt = "new file:"

	case key.Matches(msg, keys.CopyPath):
		if len(m.entries) > 0 {
			full := filepath.Join(m.path, m.entries[m.cursor].Name())
			m.statusMsg = full
			m.statusOk = true
		}

	case key.Matches(msg, keys.Search):
		m.inputMode = modeSearch
		m.searchQuery = ""
		m.searchResults = nil
		m.searchCursor = 0
		m.searchDone = false
		return m, nil

	case key.Matches(msg, keys.Command):
		m.inputMode = modeCommand
		m.cmdInput = ""
		m.cmdOutput = ""
		return m, nil
	}

	return m, nil
}

func (m Model) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.inputMode = modeNormal
		m.filterQuery = ""
		m.entries = m.applyFilterSort()
		m.clampCursor()
	case tea.KeyEnter:
		m.inputMode = modeNormal
	case tea.KeyBackspace:
		if len(m.filterQuery) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.filterQuery)
			m.filterQuery = m.filterQuery[:len(m.filterQuery)-size]
			m.entries = m.applyFilterSort()
			m.clampCursor()
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.filterQuery += string(msg.Runes)
			m.entries = m.applyFilterSort()
			m.clampCursor()
		}
	}
	return m, m.loadPreviewCmd()
}

func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.inputMode = modeNormal
		m.searchQuery = ""
		m.searchResults = nil
		m.searchDone = false
		return m, nil

	case tea.KeyEnter:
		// navigate to selected result's parent dir
		if m.searchCursor < len(m.searchResults) {
			r := m.searchResults[m.searchCursor]
			dir := r.path
			if !r.isDir {
				dir = filepath.Dir(r.path)
			}
			m.inputMode = modeNormal
			m.searchQuery = ""
			m.searchResults = nil
			m.searchDone = false
			return m, m.navigate(dir)
		}

	case tea.KeyUp:
		if m.searchCursor > 0 {
			m.searchCursor--
		}
		return m, nil

	case tea.KeyDown:
		if m.searchCursor < len(m.searchResults)-1 {
			m.searchCursor++
		}
		return m, nil

	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.searchQuery)
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-size]
			m.searchResults = nil
			m.searchCursor = 0
			m.searchDone = false
			return m, m.runSearchCmd()
		}

	default:
		if msg.Type == tea.KeyRunes {
			m.searchQuery += string(msg.Runes)
			m.searchResults = nil
			m.searchCursor = 0
			m.searchDone = false
			return m, m.runSearchCmd()
		}
	}
	return m, nil
}

func (m Model) handleCommandInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.inputMode = modeNormal
		m.cmdInput = ""
		m.cmdOutput = ""
	case tea.KeyEnter:
		if m.cmdInput != "" {
			return m, m.runCommandCmd()
		}
		m.inputMode = modeNormal
	case tea.KeyBackspace:
		if len(m.cmdInput) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.cmdInput)
			m.cmdInput = m.cmdInput[:len(m.cmdInput)-size]
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.cmdInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) runCommandCmd() tea.Cmd {
	input := strings.TrimSpace(m.cmdInput)
	cwd := m.path
	return func() tea.Msg {
		parts := strings.Fields(input)
		if len(parts) == 0 {
			return cmdOutputMsg{}
		}
		// cd is handled as TUI navigation, not a subprocess
		if parts[0] == "cd" {
			if len(parts) < 2 {
				home, _ := os.UserHomeDir()
				return dirLoadedMsg{path: home}
			}
			target := parts[1]
			if !filepath.IsAbs(target) {
				target = filepath.Join(cwd, target)
			}
			entries, err := os.ReadDir(target)
			if err != nil {
				return errMsg{fmt.Errorf("cd: %w", err)}
			}
			return dirLoadedMsg{path: target, entries: entries}
		}
		// use absolute path of the running bolt binary to avoid Windows relative-path restriction
		exe, err := os.Executable()
		if err != nil {
			return errMsg{fmt.Errorf("cannot find bolt executable: %w", err)}
		}
		cmd := exec.Command(exe, parts...)
		cmd.Dir = cwd
		out, err := cmd.CombinedOutput()
		if err != nil {
			return errMsg{fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))}
		}
		return cmdOutputMsg{out: strings.TrimSpace(string(out))}
	}
}

func (m Model) runSearchCmd() tea.Cmd {
	if m.searchQuery == "" {
		return nil
	}
	root := m.path
	query := m.searchQuery
	return func() tea.Msg {
		out := make(chan srch.Result, 128)
		var collected []searchResult
		go func() {
			defer close(out)
			_ = srch.Search(srch.Options{
				Root:          root,
				Pattern:       query,
				Mode:          srch.ModeFuzzy,
				Recursive:     true,
				Hidden:        false,
				Limit:         100,
				SearchContent: true,
			}, out)
		}()
		for r := range out {
			collected = append(collected, searchResult{
				path:         r.Path,
				name:         r.Name,
				isDir:        r.IsDir,
				contentMatch: r.ContentMatch,
			})
		}
		return searchResultMsg{results: collected, done: true}
	}
}

func (m Model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.inputMode = modeNormal
		m.inputValue = ""
	case tea.KeyEnter:
		switch m.inputMode {
		case modeRename:
			return m, m.commitRename()
		case modeNewDir:
			return m, m.commitNewDir()
		case modeNewFile:
			return m, m.commitNewFile()
		}
	case tea.KeyBackspace:
		if len(m.inputValue) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.inputValue)
			m.inputValue = m.inputValue[:len(m.inputValue)-size]
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.inputValue += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}
	if m.showHelp {
		return m.helpView()
	}

	header := m.headerView()
	status := m.statusBarView()

	reserved := lipgloss.Height(header) + lipgloss.Height(status)
	if m.inputMode != modeNormal {
		reserved++
	}
	innerH := m.height - reserved

	leftW := (m.width * 45 / 100) - 2
	rightW := m.width - leftW - 4

	left := sPaneActive.Width(leftW).Height(innerH - 2).Render(m.fileListView(leftW, innerH-2))
	var right string
	if m.previewIsImg && m.preview != "" {
		// image contains raw ANSI true-color sequences that lipgloss can't measure;
		// render it directly at exact pixel width to prevent bleed into adjacent pane
		right = m.imagePane(rightW, innerH-2)
	} else {
		right = sPane.Width(rightW).Height(innerH - 2).Render(m.previewView(rightW, innerH-2))
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	parts := []string{header, body}
	switch m.inputMode {
	case modeSearch:
		parts = append(parts, m.searchPanelView())
	case modeCommand:
		parts = append(parts, m.commandBarView())
	case modeNormal:
		// nothing
	default:
		parts = append(parts, m.inputBarView())
	}
	parts = append(parts, status)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) headerView() string {
	title := sTitle.Render(" ⚡ BOLT ")
	path := sPath.Render(shortenPath(m.path, m.width/3))

	diskInfo := ""
	if m.diskTotal > 0 {
		pct := int(100 * (m.diskTotal - m.diskFree) / m.diskTotal)
		diskInfo = sDiskBar.Render(fmt.Sprintf("  %s free  %d%% used",
			humanizeSize(int64(m.diskFree)), pct))
	}

	sortBadge := sKey.Render(fmt.Sprintf(" ↕ %s ", m.sort.String()))

	selBadge := ""
	if len(m.selected) > 0 {
		selBadge = sKey.Render(fmt.Sprintf(" ✓ %d ", len(m.selected)))
	}

	clipBadge := ""
	switch m.clip.mode {
	case clipCopy:
		clipBadge = lipgloss.NewStyle().Background(clrClipCopy).Foreground(lipgloss.Color("16")).Bold(true).Padding(0, 1).Render(fmt.Sprintf(" 󰆏 %d ", len(m.clip.paths)))
	case clipCut:
		clipBadge = lipgloss.NewStyle().Background(clrClipCut).Foreground(lipgloss.Color("16")).Bold(true).Padding(0, 1).Render(fmt.Sprintf(" ✂ %d ", len(m.clip.paths)))
	}

	right := diskInfo + "  " + selBadge + clipBadge + sortBadge + " "
	gap := max(0, m.width-lipgloss.Width(title)-lipgloss.Width(path)-lipgloss.Width(right))
	return title + path + strings.Repeat(" ", gap) + right
}

func (m Model) fileListView(w, h int) string {
	colName := max(12, w-30)

	hdr := sHeader.Width(w).Render(
		padR("  Name", colName) + padL("Size", 10) + padL("Modified", 14),
	)

	filterLine := ""
	if m.filterQuery != "" || m.inputMode == modeFilter {
		q := m.filterQuery
		if m.inputMode == modeFilter {
			q += "█"
		}
		filterLine = sFilterBar.Render(sInfo.Render("/") + " " + q +
			sMuted.Render(fmt.Sprintf("  %d matches", len(m.entries))))
	}

	if len(m.entries) == 0 {
		msg := "  (empty)"
		if m.filterQuery != "" {
			msg = "  no matches for: " + m.filterQuery
		}
		return hdr + "\n" + filterLine + "\n" + sMuted.Render(msg)
	}

	var rows []string
	rows = append(rows, hdr)
	if filterLine != "" {
		rows = append(rows, filterLine)
	}

	availH := h - len(rows)
	end := min(m.offset+availH, len(m.entries))
	for i := m.offset; i < end; i++ {
		rows = append(rows, m.renderRow(m.entries[i], i, w, colName))
	}

	if len(m.entries) > availH {
		pct := 100 * m.cursor / max(1, len(m.entries)-1)
		bar := renderScrollBar(pct, min(availH, 5))
		rows = append(rows, sMuted.Render(bar+fmt.Sprintf(" %d/%d", m.cursor+1, len(m.entries))))
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderRow(entry fs.DirEntry, idx int, _ int, colName int) string {
	info, _ := entry.Info()
	name := entry.Name()
	isHidden := strings.HasPrefix(name, ".")
	isSelected := m.selected[idx]

	icon := fileIcon(entry)
	nameStyle := fileNameStyle(entry, isHidden)
	styledName := nameStyle.Render(name)

	sizeStr := "—"
	dateStr := ""
	if info != nil {
		if !entry.IsDir() {
			sizeStr = humanizeSize(info.Size())
		}
		dateStr = info.ModTime().Format("Jan 02 15:04")
	}

	// highlight filter match
	if m.filterQuery != "" {
		styledName = highlightMatch(name, m.filterQuery, nameStyle)
	}

	displayName := icon + " " + styledName
	nameW := lipgloss.Width(displayName)
	if nameW < colName {
		displayName += strings.Repeat(" ", colName-nameW)
	}

	line := " " + displayName + padL(sizeStr, 10) + padL(dateStr, 14)

	switch {
	case idx == m.cursor && isSelected:
		plain := " " + icon + " " + name + padL(stripANSI(sizeStr), 10) + padL(stripANSI(dateStr), 14)
		return sSelected.Render("❯") + sMultiSelected.Render(plain)
	case idx == m.cursor:
		plain := " " + icon + " " + name + padL(stripANSI(sizeStr), 10) + padL(stripANSI(dateStr), 14)
		return sSelected.Render("❯" + plain)
	case isSelected:
		plain := " " + icon + " " + name + padL(stripANSI(sizeStr), 10) + padL(stripANSI(dateStr), 14)
		return sMultiSelected.Render(" " + plain)
	default:
		return " " + line
	}
}

func (m Model) previewView(w, h int) string {
	hdr := sHeader.Width(w).Render("  Preview")

	if len(m.entries) == 0 {
		return hdr + "\n" + sMuted.Render("  nothing selected")
	}

	entry := m.entries[m.cursor]
	info, _ := entry.Info()

	// sub-header with file metadata
	meta := ""
	if info != nil {
		perms := sPerms.Render(info.Mode().String())
		mod := sDate.Render(info.ModTime().Format("2006-01-02 15:04:05"))
		sz := ""
		if !entry.IsDir() {
			sz = sSize.Render("  " + humanizeSize(info.Size()))
		}
		meta = "\n " + fileIcon(entry) + " " + sNormal.Render(entry.Name()) + sz + "\n " + perms + "  " + mod + "\n"
	}

	if m.previewIsDir {
		return hdr + meta + m.dirPreview(w, h-4)
	}
	if m.previewIsImg {
		if m.preview == "" {
			return hdr + meta + "\n" + sMuted.Render("  (could not decode image)")
		}
		return hdr + meta + "\n" + m.preview
	}
	if m.preview == "" {
		return hdr + meta + "\n" + sMuted.Render("  (binary or empty file)")
	}

	lines := strings.Split(m.preview, "\n")
	maxLines := max(1, h-lipgloss.Height(hdr)-lipgloss.Height(meta)-1)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	var b strings.Builder
	b.WriteString(hdr)
	b.WriteString(meta)
	for i, line := range lines {
		if lipgloss.Width(line) > w-4 {
			line = truncate(line, w-4)
		}
		fmt.Fprintf(&b, " %s %s\n", sMuted.Render(fmt.Sprintf("%3d", i+1)), line)
	}
	return b.String()
}

func (m Model) dirPreview(_ int, h int) string {
	if len(m.previewDir) == 0 {
		return sMuted.Render("  (empty directory)")
	}
	var b strings.Builder
	dirs, files := 0, 0
	for _, e := range m.previewDir {
		if e.IsDir() {
			dirs++
		} else {
			files++
		}
	}
	fmt.Fprintf(&b, " %s  %s\n\n",
		sMuted.Render(fmt.Sprintf("%d dirs", dirs)),
		sMuted.Render(fmt.Sprintf("%d files", files)),
	)
	for i, e := range m.previewDir {
		if i >= h-3 {
			fmt.Fprintf(&b, " %s\n", sMuted.Render(fmt.Sprintf("  … %d more", len(m.previewDir)-i)))
			break
		}
		info, _ := e.Info()
		sz := ""
		if info != nil && !e.IsDir() {
			sz = sMuted.Render("  " + humanizeSize(info.Size()))
		}
		fmt.Fprintf(&b, " %s %s%s\n", fileIcon(e), fileNameStyle(e, strings.HasPrefix(e.Name(), ".")).Render(e.Name()), sz)
	}
	return b.String()
}

func (m Model) inputBarView() string {
	prompt := sKey.Render(" " + m.inputPrompt + " ")
	val := m.inputValue
	if m.inputMode == modeFilter {
		val = m.filterQuery
	}
	input := sNormal.Render(" " + val + "█")
	hint := sMuted.Render("  esc cancel  ↵ confirm")
	gap := max(0, m.width-lipgloss.Width(prompt)-lipgloss.Width(input)-lipgloss.Width(hint)-2)
	return prompt + input + strings.Repeat(" ", gap) + hint
}

func (m Model) imagePane(w, h int) string {
	border := lipgloss.RoundedBorder()
	borderStyle := lipgloss.NewStyle().Foreground(clrBorder)

	topBar := borderStyle.Render(border.TopLeft) +
		borderStyle.Render(strings.Repeat(border.Top, w)) +
		borderStyle.Render(border.TopRight)

	botBar := borderStyle.Render(border.BottomLeft) +
		borderStyle.Render(strings.Repeat(border.Bottom, w)) +
		borderStyle.Render(border.BottomRight)

	side := borderStyle.Render(border.Left)
	sideR := borderStyle.Render(border.Right)

	// Split rendered image into lines; strip trailing cursor/movement sequences
	raw := strings.TrimRight(m.preview, "\n\r")
	imgLines := strings.Split(raw, "\n")
	var rows []string
	rows = append(rows, topBar)
	for i := range h {
		var content string
		if i < len(imgLines) {
			line := imgLines[i]
			// strip any trailing non-color escape sequences (cursor moves, etc.)
			// by re-appending a hard reset and trusting lipgloss.Width for padding
			content = line + "\x1b[0m"
		}
		// pad to exactly w visible chars so right border aligns
		visW := lipgloss.Width(content)
		pad := max(0, w-visW)
		rows = append(rows, side+content+strings.Repeat(" ", pad)+sideR)
	}
	rows = append(rows, botBar)
	return strings.Join(rows, "\n")
}

func (m Model) commandBarView() string {
	prompt := sKey.Render(" : ")
	input := sNormal.Render(m.cmdInput + "█")
	hint := sMuted.Render("  ↵ run  esc cancel")
	gap := max(0, m.width-lipgloss.Width(prompt)-lipgloss.Width(input)-lipgloss.Width(hint)-2)
	bar := prompt + input + strings.Repeat(" ", gap) + hint
	if m.cmdOutput != "" {
		out := m.cmdOutput
		if lipgloss.Width(out) > m.width-2 {
			out = truncate(out, m.width-2)
		}
		return bar + "\n" + sInfo.Render(" "+out)
	}
	return bar
}

func (m Model) searchPanelView() string {
	spinner := ""
	if !m.searchDone && m.searchQuery != "" {
		spinner = sMuted.Render(" …")
	}
	prompt := sKey.Render(" search ")
	input := sNormal.Render(" " + m.searchQuery + "█")
	count := ""
	if len(m.searchResults) > 0 {
		count = sMuted.Render(fmt.Sprintf("  %d results", len(m.searchResults)))
	}
	hint := sMuted.Render("  ↑↓ nav  ↵ goto  esc close")
	gap := max(0, m.width-lipgloss.Width(prompt)-lipgloss.Width(input)-lipgloss.Width(count)-lipgloss.Width(hint)-lipgloss.Width(spinner)-2)
	header := prompt + input + count + spinner + strings.Repeat(" ", gap) + hint

	if len(m.searchResults) == 0 {
		if m.searchQuery == "" {
			return header
		}
		if m.searchDone {
			return header + "\n" + sMuted.Render("  no results")
		}
		return header
	}

	maxShow := min(8, len(m.searchResults))
	start := max(0, m.searchCursor-maxShow+1)
	if start+maxShow > len(m.searchResults) {
		start = max(0, len(m.searchResults)-maxShow)
	}

	var rows []string
	rows = append(rows, header)
	for i := start; i < start+maxShow && i < len(m.searchResults); i++ {
		r := m.searchResults[i]
		icon := " "
		if r.isDir {
			icon = sDir.Render(" ")
		}
		label := ""
		if r.contentMatch {
			label = sMuted.Render(" ~")
		}
		line := icon + " " + r.path + label
		if lipgloss.Width(line) > m.width-4 {
			line = icon + " " + truncate(r.path, m.width-6)
		}
		if i == m.searchCursor {
			rows = append(rows, sSearchCursor.Render("❯"+line))
		} else {
			rows = append(rows, sSearchResult.Render(" "+line))
		}
	}
	return sSearchBar.Width(m.width - 2).Render(strings.Join(rows, "\n"))
}

func (m Model) statusBarView() string {
	if m.err != nil {
		return sError.Render(" ✗ " + m.err.Error())
	}

	var left string
	if m.statusMsg != "" {
		if m.statusOk {
			left = sOk.Render(" ✓ ") + " " + m.statusMsg
		} else {
			left = " " + m.statusMsg
		}
	} else if len(m.entries) > 0 {
		entry := m.entries[m.cursor]
		info, _ := entry.Info()
		left = fmt.Sprintf(" %d/%d", m.cursor+1, len(m.entries))
		if info != nil {
			left += "  " + sPerms.Render(info.Mode().String()) +
				"  " + sDate.Render(info.ModTime().Format(time.RFC822))
		}
	} else {
		left = " 0 items"
	}

	shorts := []struct{ k, v string }{
		{"↑↓", "nav"}, {"↵", "open"}, {"spc", "sel"}, {"^a", "all"},
		{"c", "copy"}, {"x", "cut"}, {"p", "paste"}, {"D", "del"},
		{"r", "ren"}, {"m", "mkdir"}, {"n", "new"}, {"/", "filter"},
		{"^f", "search"}, {":", "cmd"}, {"s", "sort"}, {"~", "home"}, {"[/]", "history"}, {"?", "help"}, {"q", "quit"},
	}
	var parts []string
	for _, s := range shorts {
		parts = append(parts, sKey.Render(s.k)+" "+sMuted.Render(s.v))
	}
	right := " " + strings.Join(parts, "  ") + " "
	gap := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return sStatusBar.Width(m.width).Render(left + strings.Repeat(" ", gap) + right)
}

func (m Model) helpView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n %s\n\n", sTitle.Render(" ⚡ BOLT — Keybindings "))

	sections := []struct {
		heading string
		items   []key.Binding
	}{
		{"Navigation", []key.Binding{keys.Up, keys.Down, keys.Home, keys.End, keys.PageUp, keys.PageDown, keys.Enter, keys.Back, keys.GotoHome, keys.NavBack, keys.NavForward}},
		{"Selection & Clipboard", []key.Binding{keys.Select, keys.SelectAll, keys.Copy, keys.Cut, keys.Paste}},
		{"File Operations", []key.Binding{keys.Delete, keys.Rename, keys.NewDir, keys.NewFile, keys.CopyPath}},
		{"View", []key.Binding{keys.Filter, keys.Search, keys.SortCycle, keys.ToggleHidden, keys.Help}},
		{"App", []key.Binding{keys.Quit}},
	}

	for _, sec := range sections {
		fmt.Fprintf(&b, " %s\n\n", sHeader.Render(fmt.Sprintf("  %s  ", sec.heading)))
		for _, kb := range sec.items {
			fmt.Fprintf(&b, "   %-22s  %s\n",
				sKey.Render(" "+kb.Help().Key+" "),
				kb.Help().Desc,
			)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, " %s\n", sMuted.Render("press ? to close"))
	return b.String()
}

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

func (m Model) navigate(path string) tea.Cmd {
	// truncate forward history on new navigation
	return func() tea.Msg {
		abs, err := filepath.Abs(path)
		if err != nil {
			return errMsg{err}
		}
		entries, err := os.ReadDir(abs)
		return dirLoadedMsg{path: abs, entries: entries, err: err}
	}
}

func (m Model) loadPreviewCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.cursor]
	target := filepath.Join(m.path, entry.Name())
	w, h := m.previewDims()
	return func() tea.Msg {
		if entry.IsDir() {
			sub, _ := os.ReadDir(target)
			return previewLoadedMsg{isDir: true, entries: sub}
		}
		if isImageFile(entry.Name()) {
			rendered := renderImagePreview(target, w, h)
			return previewLoadedMsg{content: rendered, isImage: true}
		}
		info, err := os.Stat(target)
		if err != nil || info.Size() > 256*1024 {
			return previewLoadedMsg{}
		}
		data, err := os.ReadFile(target)
		if err != nil || isBinary(data) {
			return previewLoadedMsg{}
		}
		return previewLoadedMsg{content: syntaxHighlight(target, string(data))}
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

func (m Model) buildClipboard(mode clipMode) clipboard {
	var paths []string
	if len(m.selected) > 0 {
		for idx := range m.selected {
			if idx < len(m.entries) {
				paths = append(paths, filepath.Join(m.path, m.entries[idx].Name()))
			}
		}
	} else if len(m.entries) > 0 {
		paths = []string{filepath.Join(m.path, m.entries[m.cursor].Name())}
	}
	return clipboard{mode: mode, paths: paths}
}

func (m Model) pasteCmd() tea.Cmd {
	if m.clip.mode == clipNone || len(m.clip.paths) == 0 {
		return func() tea.Msg { return errMsg{fmt.Errorf("clipboard is empty")} }
	}
	dest := m.path
	paths := m.clip.paths
	mode := m.clip.mode
	return func() tea.Msg {
		for _, src := range paths {
			target := filepath.Join(dest, filepath.Base(src))
			if mode == clipCopy {
				if err := copyPath(src, target); err != nil {
					return errMsg{err}
				}
			} else {
				if err := os.Rename(src, target); err != nil {
					return errMsg{err}
				}
			}
		}
		return loadDir(dest)()
	}
}

func (m Model) deleteCmd() tea.Cmd {
	var targets []string
	if len(m.selected) > 0 {
		for idx := range m.selected {
			if idx < len(m.entries) {
				targets = append(targets, filepath.Join(m.path, m.entries[idx].Name()))
			}
		}
	} else if len(m.entries) > 0 {
		targets = []string{filepath.Join(m.path, m.entries[m.cursor].Name())}
	}
	path := m.path
	return func() tea.Msg {
		for _, t := range targets {
			if err := os.RemoveAll(t); err != nil {
				return errMsg{fmt.Errorf("delete failed: %w", err)}
			}
		}
		return loadDir(path)()
	}
}

func (m Model) commitRename() tea.Cmd {
	if m.inputValue == "" || len(m.entries) == 0 {
		return nil
	}
	src := filepath.Join(m.path, m.entries[m.cursor].Name())
	dst := filepath.Join(m.path, m.inputValue)
	path := m.path
	return func() tea.Msg {
		if err := os.Rename(src, dst); err != nil {
			return errMsg{fmt.Errorf("rename failed: %w", err)}
		}
		return loadDir(path)()
	}
}

func (m Model) commitNewDir() tea.Cmd {
	if m.inputValue == "" {
		return nil
	}
	target := filepath.Join(m.path, m.inputValue)
	path := m.path
	return func() tea.Msg {
		if err := os.MkdirAll(target, 0755); err != nil {
			return errMsg{fmt.Errorf("mkdir failed: %w", err)}
		}
		return loadDir(path)()
	}
}

func (m Model) commitNewFile() tea.Cmd {
	if m.inputValue == "" {
		return nil
	}
	target := filepath.Join(m.path, m.inputValue)
	path := m.path
	return func() tea.Msg {
		f, err := os.Create(target)
		if err != nil {
			return errMsg{fmt.Errorf("create failed: %w", err)}
		}
		f.Close()
		return loadDir(path)()
	}
}

func (m Model) applyFilterSort() []fs.DirEntry {
	entries := make([]fs.DirEntry, 0, len(m.allEntries))
	for _, e := range m.allEntries {
		if !m.showHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if m.filterQuery != "" && !fuzzyMatch(e.Name(), m.filterQuery) {
			continue
		}
		entries = append(entries, e)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		// dirs always first
		if a.IsDir() != b.IsDir() {
			return a.IsDir()
		}
		switch m.sort {
		case sortSize:
			ai, _ := a.Info()
			bi, _ := b.Info()
			if ai != nil && bi != nil {
				return ai.Size() < bi.Size()
			}
		case sortDate:
			ai, _ := a.Info()
			bi, _ := b.Info()
			if ai != nil && bi != nil {
				return ai.ModTime().After(bi.ModTime())
			}
		case sortType:
			return filepath.Ext(a.Name()) < filepath.Ext(b.Name())
		}
		return strings.ToLower(a.Name()) < strings.ToLower(b.Name())
	})
	return entries
}

func fuzzyMatch(name, query string) bool {
	name = strings.ToLower(name)
	query = strings.ToLower(query)
	qi := 0
	for _, ch := range name {
		if qi < len(query) && rune(query[qi]) == ch {
			qi++
		}
	}
	return qi == len(query)
}

func (m *Model) clampCursor() {
	if m.cursor >= len(m.entries) {
		m.cursor = max(0, len(m.entries)-1)
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
}

func (m Model) listHeight() int {
	return max(1, m.height-8)
}

func (m Model) previewDims() (w, h int) {
	leftW := (m.width * 45 / 100) - 2
	rightW := m.width - leftW - 4
	reserved := 6 // header + statusbar + borders
	return rightW - 2, max(1, m.height-reserved)
}

func fileIcon(e fs.DirEntry) string {
	if e.IsDir() {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(e.Name()))
	base := strings.ToLower(e.Name())

	switch base {
	case "dockerfile", ".dockerignore":
		return ""
	case "makefile", "gnumakefile":
		return ""
	case ".gitignore", ".gitmodules", ".gitattributes":
		return ""
	case "license", "licence":
		return ""
	case "readme.md", "readme.txt", "readme":
		return ""
	}

	switch ext {
	case ".go":
		return ""
	case ".js", ".mjs", ".cjs":
		return ""
	case ".ts", ".tsx":
		return ""
	case ".jsx":
		return ""
	case ".py":
		return ""
	case ".rs":
		return ""
	case ".rb":
		return ""
	case ".java":
		return ""
	case ".c", ".h":
		return ""
	case ".cpp", ".cc", ".cxx", ".hpp":
		return ""
	case ".cs":
		return "󰌛"
	case ".php":
		return ""
	case ".swift":
		return ""
	case ".kt", ".kts":
		return ""
	case ".html", ".htm":
		return ""
	case ".css":
		return ""
	case ".scss", ".sass":
		return ""
	case ".json":
		return ""
	case ".yaml", ".yml":
		return ""
	case ".toml":
		return ""
	case ".xml":
		return "󰗀"
	case ".md", ".mdx":
		return ""
	case ".txt":
		return ""
	case ".sh", ".bash", ".zsh", ".fish":
		return ""
	case ".ps1":
		return ""
	case ".env":
		return ""
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".bmp":
		return ""
	case ".mp4", ".mkv", ".avi", ".mov", ".webm":
		return ""
	case ".mp3", ".flac", ".wav", ".ogg", ".aac":
		return ""
	case ".pdf":
		return ""
	case ".zip", ".tar", ".gz", ".bz2", ".xz", ".7z", ".rar":
		return ""
	case ".exe", ".msi":
		return ""
	case ".deb", ".rpm", ".appimage":
		return ""
	case ".db", ".sqlite", ".sqlite3":
		return ""
	case ".lock":
		return ""
	case ".log":
		return ""
	case ".pem", ".key", ".cert", ".crt":
		return ""
	}

	info, _ := e.Info()
	if info != nil && info.Mode()&0111 != 0 {
		return ""
	}
	return ""
}

func fileNameStyle(e fs.DirEntry, isHidden bool) lipgloss.Style {
	if isHidden {
		return sHidden
	}
	if e.IsDir() {
		return sDir
	}
	if e.Type()&fs.ModeSymlink != 0 {
		return sSymlink
	}

	ext := strings.ToLower(filepath.Ext(e.Name()))
	switch ext {
	case ".go":
		return sGo
	case ".js", ".mjs", ".cjs", ".jsx", ".ts", ".tsx":
		return sJS
	case ".py":
		return sPython
	case ".rs":
		return sRust
	case ".md", ".mdx", ".txt":
		return sMarkdown
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".bmp":
		return sImage
	case ".mp4", ".mkv", ".avi", ".mov", ".mp3", ".flac", ".wav":
		return sMedia
	case ".zip", ".tar", ".gz", ".bz2", ".xz", ".7z", ".rar":
		return sArchive
	case ".json", ".yaml", ".yml", ".toml", ".env":
		return sConfig
	}

	info, _ := e.Info()
	if info != nil && info.Mode()&0111 != 0 {
		return sExec
	}
	return sNormal
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst, info.Mode())
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	buf := make([]byte, 32*1024)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if err != nil {
			break
		}
	}
	return out.Sync()
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

func padR(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func padL(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return strings.Repeat(" ", n-w) + s
}

func truncate(s string, n int) string {
	if lipgloss.Width(s) <= n {
		return s
	}
	return s[:max(0, n-1)] + "…"
}

func shortenPath(p string, maxW int) string {
	if len(p) <= maxW {
		return p
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(p, home) {
		p = "~" + p[len(home):]
	}
	if len(p) <= maxW {
		return p
	}
	parts := strings.Split(p, string(os.PathSeparator))
	if len(parts) > 3 {
		return "…" + string(os.PathSeparator) + strings.Join(parts[len(parts)-2:], string(os.PathSeparator))
	}
	return p
}

func renderScrollBar(pct, h int) string {
	filled := pct * h / 100
	var b strings.Builder
	for i := range h {
		if i <= filled {
			b.WriteString(sInfo.Render("█"))
		} else {
			b.WriteString(sMuted.Render("░"))
		}
	}
	return b.String()
}

func highlightMatch(name, query string, base lipgloss.Style) string {
	hl := lipgloss.NewStyle().Foreground(clrYellow).Bold(true)
	qi := 0
	query = strings.ToLower(query)
	var b strings.Builder
	for _, ch := range name {
		if qi < len(query) && strings.ToLower(string(ch)) == string(query[qi]) {
			b.WriteString(hl.Render(string(ch)))
			qi++
		} else {
			b.WriteString(base.Render(string(ch)))
		}
	}
	return b.String()
}

func isBinary(data []byte) bool {
	return slices.Contains(data[:min(len(data), 512)], byte(0))
}

func syntaxHighlight(path, content string) string {
	lexer := chromaLexers.Match(filepath.Base(path))
	if lexer == nil {
		lexer = chromaLexers.Analyse(content)
	}
	if lexer == nil {
		lexer = chromaLexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iter, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}
	var b strings.Builder
	if err := formatter.Format(&b, style, iter); err != nil {
		return content
	}
	return b.String()
}

func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case !inEsc:
			b.WriteRune(r)
		}
	}
	return b.String()
}
