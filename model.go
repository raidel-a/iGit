package main

import (
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/rai/interactive-git/git"
	"github.com/rai/interactive-git/ui"
)

// AppState represents the current state of the application
type AppState int

const (
	StateFileList AppState = iota
	StateCommitMessage
	StateCommitDate
	StateModifyHead
	StateHelp
)

// CommitState represents the current commit input state
type CommitState int

const (
	CommitStateMessage CommitState = iota
	CommitStateDate
	CommitStateConfirm
)

// HeadModifyState represents the current HEAD modification state
type HeadModifyState int

const (
	HeadModifyStateMenu HeadModifyState = iota
	HeadModifyStateAmendMessage
	HeadModifyStateAmendFiles
)

// Model holds the application state
type Model struct {
	// State
	state      AppState
	width      int
	height     int
	ready      bool
	err        string
	status     string
	processing bool

	// Git data
	gitClient *git.Client
	files     []git.FileItem
	gitStatus git.GitStatus

	// UI Components
	list       list.Model
	viewport   viewport.Model
	keys       ui.KeyMap
	delegate   *FileDelegate

	// UI State
	selectedFiles   map[int]bool
	showPreview     bool
	previewFocused  bool // Track if preview pane has focus
	lastStatusMsg   time.Time
	lastFileIndex   int // Track last fetched file to avoid redundant diffs

	// Preview/Layout
	previewContent string
	diffCache      map[string]string // Cache file diffs
	layout         ui.Layout

	// Commit UI
	commitTextarea textarea.Model
	commitInput    textinput.Model
	commitMessage  string
	commitDate     string
	commitState    CommitState

	// HEAD Modification
	headInfo           *git.CommitInfo
	headModifyState    HeadModifyState
	headMessageTextarea textarea.Model
}

// FileDelegate is a custom delegate for rendering file items
type FileDelegate struct {
	styles FileStyles
}

type FileStyles struct {
	Normal   lipgloss.Style
	Selected lipgloss.Style
	Staged   lipgloss.Style
	Unstaged lipgloss.Style
	Untracked lipgloss.Style
}

// Height returns the height of a list item
func (d *FileDelegate) Height() int { return 1 }

// Spacing returns the spacing between items
func (d *FileDelegate) Spacing() int { return 0 }

// Update handles messages for the delegate (unused in this context)
func (d *FileDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// Render renders a file item
func (d *FileDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	fileItem, ok := item.(git.FileItem)
	if !ok {
		return
	}

	// Determine style
	var style lipgloss.Style
	if index == m.Index() {
		style = d.styles.Selected
	} else {
		switch fileItem.Status {
		case git.StatusStaged:
			style = d.styles.Staged
		case git.StatusUnstaged:
			style = d.styles.Unstaged
		case git.StatusUntracked:
			style = d.styles.Untracked
		default:
			style = d.styles.Normal
		}
	}

	// Build display string
	checkbox := " "
	if fileItem.Selected {
		checkbox = "X"
	}

	statusColor := ui.FileStatusColor(fileItem.StatusSymbol)
	statusStr := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(fileItem.StatusSymbol)

	line := fmt.Sprintf("[%s] %s %s", checkbox, statusStr, fileItem.Path)
	fmt.Fprint(w, style.Render(line))
}

// NewModel creates a new model
func NewModel() Model {
	// Initialize git client
	gitClient, err := git.NewClient(".")
	if err != nil {
		return Model{
			err: fmt.Sprintf("Error: %v", err),
		}
	}

	// Create list
	delegate := &FileDelegate{
		styles: FileStyles{
			Normal:    ui.ListItemNormalStyle,
			Selected:  ui.ListItemSelectedStyle,
			Staged:    ui.StagedStyle,
			Unstaged:  ui.UnstagedStyle,
			Untracked: ui.UntrackedStyle,
		},
	}

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.Styles.Title = ui.TitleStyle

	// Create viewport for preview
	vp := viewport.New(0, 0)
	vp.Style = ui.PreviewStyle

	// Create commit textarea
	ta := textarea.New()
	ta.Placeholder = "Enter commit message..."
	ta.SetWidth(60)
	ta.SetHeight(5)
	ta.ShowLineNumbers = false

	// Create commit date input
	ti := textinput.New()
	ti.Placeholder = "YYYY-MM-DD HH:MM:SS or leave empty"
	ti.CharLimit = 50
	ti.Width = 50

	// Create HEAD message textarea for amending
	headTA := textarea.New()
	headTA.Placeholder = "Enter new commit message..."
	headTA.SetWidth(60)
	headTA.SetHeight(5)
	headTA.ShowLineNumbers = false

	m := Model{
		state:               StateFileList,
		gitClient:           gitClient,
		list:                l,
		viewport:            vp,
		keys:                ui.DefaultKeyMap(),
		delegate:            delegate,
		selectedFiles:       make(map[int]bool),
		showPreview:         true,
		previewFocused:      false,
		ready:               false,
		lastFileIndex:       -1,
		diffCache:           make(map[string]string),
		layout:              ui.NewLayout(80, 24), // Default size, will be updated on first render
		commitTextarea:      ta,
		commitInput:         ti,
		commitState:         CommitStateMessage,
		headInfo:            nil,
		headModifyState:     HeadModifyStateMenu,
		headMessageTextarea: headTA,
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.fetchGitStatus()
}

// fetchGitStatus fetches the current git status
func (m Model) fetchGitStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.gitClient.Status()
		if err != nil {
			return errorMsg{err: fmt.Sprintf("Failed to get git status: %v", err)}
		}
		return gitStatusMsg{status: status}
	}
}

// Custom message types
type gitStatusMsg struct {
	status git.GitStatus
}

type gitHeadInfoMsg struct {
	info *git.CommitInfo
}

type errorMsg struct {
	err string
}

type statusMsg struct {
	msg string
}

// toggleSelection toggles the selection of a file at the given index
func (m *Model) toggleSelection(index int) {
	if index < 0 || index >= len(m.files) {
		return
	}
	m.selectedFiles[index] = !m.selectedFiles[index]
	m.files[index].Selected = m.selectedFiles[index]

	// Update the list item
	items := make([]list.Item, len(m.files))
	for i, f := range m.files {
		items[i] = f
	}
	m.list.SetItems(items)
}

// selectAll selects all files
func (m *Model) selectAll() {
	for i := range m.files {
		m.selectedFiles[i] = true
		m.files[i].Selected = true
	}

	items := make([]list.Item, len(m.files))
	for i, f := range m.files {
		items[i] = f
	}
	m.list.SetItems(items)
}

// deselectAll deselects all files
func (m *Model) deselectAll() {
	m.selectedFiles = make(map[int]bool)
	for i := range m.files {
		m.files[i].Selected = false
	}

	items := make([]list.Item, len(m.files))
	for i, f := range m.files {
		items[i] = f
	}
	m.list.SetItems(items)
}

// getSelectedFiles returns the selected files
func (m *Model) getSelectedFiles() []git.FileItem {
	var selected []git.FileItem
	for i, f := range m.files {
		if m.selectedFiles[i] {
			selected = append(selected, f)
		}
	}
	return selected
}

// clearStatus clears the status message after a delay
func (m *Model) clearStatus() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return statusMsg{msg: ""}
	})
}

// clearError clears the error message after a delay
func (m *Model) clearError() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return errorMsg{err: ""}
	})
}

// applySelection stages or unstages selected files
func (m *Model) applySelection() tea.Cmd {
	selected := m.getSelectedFiles()
	if len(selected) == 0 {
		return func() tea.Msg {
			return statusMsg{msg: "No files selected"}
		}
	}

	m.processing = true
	cmd := m.toggleSelectionCmd(selected)

	// After the git operation, refresh status and clear selection
	return tea.Batch(
		cmd,
		func() tea.Msg {
			return gitRefreshMsg{}
		},
	)
}

// refreshStatus fetches the latest git status
func (m *Model) refreshStatus() tea.Cmd {
	return m.refreshStatusCmd()
}

// getCurrentFile returns the currently selected file
func (m *Model) getCurrentFile() *git.FileItem {
	if m.list.Index() < 0 || m.list.Index() >= len(m.files) {
		return nil
	}
	return &m.files[m.list.Index()]
}

// togglePreview toggles the preview pane visibility
func (m *Model) togglePreview() {
	m.showPreview = !m.showPreview
	// Recalculate layout if preview toggle changes the effective width
	if m.showPreview && !m.layout.HasPreviewPane() {
		// Try to recalculate with current dimensions
		m.layout = ui.NewLayout(m.width, m.height)
	}
}

// enterCommitMode enters the commit message input state
func (m *Model) enterCommitMode() {
	m.state = StateCommitMessage
	m.commitState = CommitStateMessage
	m.commitMessage = ""
	m.commitDate = ""
	m.commitTextarea.Reset()
	m.commitTextarea.Focus()
}

// getStagedFilesList returns a formatted list of staged files
func (m *Model) getStagedFilesList() string {
	if len(m.gitStatus.Staged) == 0 {
		return "No files staged"
	}
	var result string
	for _, f := range m.gitStatus.Staged {
		result += fmt.Sprintf("  + %s\n", f)
	}
	return result
}

// proceedToDateInput moves to the date input state
func (m *Model) proceedToDateInput() {
	m.commitState = CommitStateDate
	m.commitInput.Reset()
	m.commitInput.Focus()
}

// cancelCommit cancels the commit and returns to file list
func (m *Model) cancelCommit() {
	m.state = StateFileList
	m.commitMessage = ""
	m.commitDate = ""
	m.commitTextarea.Blur()
	m.commitInput.Blur()
}

// fetchHeadInfo fetches the current HEAD commit information
func (m *Model) fetchHeadInfo() tea.Cmd {
	return func() tea.Msg {
		info, err := m.gitClient.GetHeadCommitInfo()
		if err != nil {
			return errorMsg{err: fmt.Sprintf("Failed to get HEAD info: %v", err)}
		}
		return gitHeadInfoMsg{info: info}
	}
}

// enterModifyHeadMode enters the HEAD modification menu
func (m *Model) enterModifyHeadMode() {
	m.state = StateModifyHead
	m.headModifyState = HeadModifyStateMenu
}

// enterAmendMessageMode enters the amend message input state
func (m *Model) enterAmendMessageMode() {
	m.headModifyState = HeadModifyStateAmendMessage
	if m.headInfo != nil {
		m.headMessageTextarea.SetValue(m.headInfo.Message)
	}
	m.headMessageTextarea.Focus()
}

// enterAmendFilesMode enters the amend files (soft reset) mode
func (m *Model) enterAmendFilesMode() {
	m.headModifyState = HeadModifyStateAmendFiles
	m.status = "Performing soft reset..."
}

// cancelModifyHead cancels HEAD modification and returns to file list
func (m *Model) cancelModifyHead() {
	m.state = StateFileList
	m.headModifyState = HeadModifyStateMenu
	m.headMessageTextarea.Blur()
	m.headInfo = nil
}
