package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/rai/interactive-git/ui"
)

// View renders the application
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.err != "" {
		return m.renderError()
	}

	switch m.state {
	case StateFileList:
		return m.renderFileList()
	case StateCommitMessage, StateCommitDate:
		return m.renderCommitView()
	case StateModifyHead:
		return m.renderModifyHeadView()
	case StateHelp:
		return m.renderHelp()
	default:
		return m.renderFileList()
	}
}

// renderError renders the error view
func (m Model) renderError() string {
	return ui.ErrorStyle.Render("[ERROR] " + m.err)
}

// renderCommitView renders the commit workflow view
func (m Model) renderCommitView() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Title
	title := ui.TitleStyle.Render("Commit Staged Files")
	sections = append(sections, "", title, "")

	// Show files to be committed
	filesList := "Files to commit:\n" + m.getStagedFilesList()
	sections = append(sections, filesList, "")

	// Show input based on commit state
	if m.commitState == CommitStateMessage {
		// Show message input
		sections = append(sections, ui.TitleStyle.Render("Commit Message"))
		sections = append(sections, m.commitTextarea.View())
		sections = append(sections, "")
		sections = append(sections, ui.HelpStyle.Render("[Ctrl+D] Continue  [Esc] Cancel"))
	} else if m.commitState == CommitStateDate {
		// Show date input (optional)
		sections = append(sections, ui.TitleStyle.Render("Commit Date (Optional)"))
		sections = append(sections, "Leave empty for current time")
		sections = append(sections, "Format: YYYY-MM-DD or YYYY-MM-DD HH:MM:SS")
		sections = append(sections, "")
		sections = append(sections, m.commitInput.View())
		sections = append(sections, "")
		sections = append(sections, ui.HelpStyle.Render("[Enter] Commit  [Esc] Back"))
	}

	content := strings.Join(sections, "\n")
	return lipgloss.NewStyle().Padding(1).Render(content)
}

// renderFileList renders the main file list view
func (m Model) renderFileList() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Main content
	content := m.renderMainContent()
	sections = append(sections, content)

	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHeader renders the header
func (m Model) renderHeader() string {
	width := m.width
	if width < 30 {
		width = 30
	}

	title := "gitUI"
	divider := strings.Repeat("━", width)

	titleLine := lipgloss.Place(
		width, 1,
		lipgloss.Center, lipgloss.Center,
		title,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		divider,
		titleLine,
		divider,
	)
}

// renderMainContent renders the main content (file list and preview)
func (m Model) renderMainContent() string {
	// If preview is focused, show it full screen (works even on small terminals)
	if m.previewFocused && m.showPreview {
		// Subtract border (2 chars) and padding (2 chars) overhead
		previewWidth := m.width - 4
		if previewWidth < 20 {
			previewWidth = 20
		}
		return m.renderPreview(previewWidth, m.layout.ListHeight()+m.layout.PreviewHeight()-2)
	}

	// If preview is disabled or layout doesn't support split view, just show list
	if !m.showPreview || !m.layout.HasPreviewPane() {
		// Build status title for list
		statusTitle := fmt.Sprintf(
			"Files - Staged: %d | Unstaged: %d | Untracked: %d | Selected: %d",
			m.gitStatus.StagedCount(),
			m.gitStatus.UnstagedCount(),
			m.gitStatus.UntrackedCount(),
			len(m.selectedFiles),
		)
		m.list.Title = statusTitle

		// Subtract border (2 chars) and padding (2 chars) overhead
		listWidth := m.width - 4
		if listWidth < 20 {
			listWidth = 20
		}
		listView := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorBlue).
			Width(listWidth).
			Height(m.layout.ListHeight()).
			Padding(0, 1).
			Render(m.list.View())
		return listView
	}

	// Use consistent height for both panes
	paneHeight := m.layout.ListHeight()

	// Build status title for list
	statusTitle := fmt.Sprintf(
		"Files - Staged: %d | Unstaged: %d | Untracked: %d | Selected: %d",
		m.gitStatus.StagedCount(),
		m.gitStatus.UnstagedCount(),
		m.gitStatus.UntrackedCount(),
		len(m.selectedFiles),
	)
	m.list.Title = statusTitle

	// Render file list pane
	// Subtract border (2 chars) and padding (2 chars) overhead
	listWidth := m.layout.ListWidth - 4
	if listWidth < 20 {
		listWidth = 20
	}
	listView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorBlue).
		Width(listWidth).
		Height(paneHeight).
		Padding(0, 1).
		Render(m.list.View())

	// Render preview pane with same height
	// Subtract border (2 chars) and padding (2 chars) overhead
	previewWidth := m.layout.PreviewWidth - 4
	if previewWidth < 20 {
		previewWidth = 20
	}
	previewView := m.renderPreview(previewWidth, paneHeight)

	// Join horizontally
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listView,
		previewView,
	)

	return content
}

// renderPreview renders the preview pane
func (m Model) renderPreview(width, height int) string {
	if width < 10 || height < 3 {
		return ""
	}

	title := "Preview"
	if m.previewFocused {
		title = "Preview (FOCUSED)"
	}
	var content string

	if m.list.Index() >= 0 && m.list.Index() < len(m.files) {
		file := m.files[m.list.Index()]
		if m.previewFocused {
			title = fmt.Sprintf("Preview: %s (%s) [FOCUSED]", file.Path, file.Status.String())
		} else {
			title = fmt.Sprintf("Preview: %s (%s)", file.Path, file.Status.String())
		}

		// Show preview content
		if m.previewContent == "" {
			content = "[...] Loading preview..."
		} else {
			// Content is ready - show it
			content = m.viewport.View()
		}
	} else {
		content = "[No file selected]"
	}

	previewBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorBlue).
		Width(width).
		Height(height).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			ui.PreviewTitleStyle.Render(title),
			content,
		))

	return previewBox
}

// renderFooter renders the footer with keybinding hints
func (m Model) renderFooter() string {
	var sections []string

	// Status or error line
	if m.err != "" {
		sections = append(sections, ui.ErrorStyle.Render("[!] "+m.err))
	} else if m.status != "" {
		statusLine := m.status
		if m.processing {
			statusLine = statusLine + " [...]"
		}
		sections = append(sections, ui.InfoStyle.Render(statusLine))
	}

	// Show keybinding hints
	keybindingHint := ui.HelpStyle.Render("[Space] Toggle  [a] Select All  [d] Deselect All  [Enter] Apply  [c] Commit  [m] Modify HEAD  [p] Preview  [?] Help  [q] Quit")
	sections = append(sections, keybindingHint)

	footer := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		Padding(0, 1).
		Render(footer)
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Help content
	helpContent := m.renderHelpContent()
	sections = append(sections, helpContent)

	// Instructions
	instructions := ui.HelpStyle.Render("Press [?] or [q] to close help")
	sections = append(sections, instructions)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHelpContent renders the help content
func (m Model) renderHelpContent() string {
	var helpLines []string

	helpLines = append(helpLines, "")
	helpLines = append(helpLines, ui.TitleStyle.Render("Navigation"))
	helpLines = append(helpLines, "  ↑/k, ↓/j       Move up/down in list")
	helpLines = append(helpLines, "  Home/g, End/G   Jump to top/bottom")
	helpLines = append(helpLines, "")

	helpLines = append(helpLines, ui.TitleStyle.Render("Selection"))
	helpLines = append(helpLines, "  Space/Tab       Toggle file selection")
	helpLines = append(helpLines, "  a               Select all files")
	helpLines = append(helpLines, "  d               Deselect all files")
	helpLines = append(helpLines, "")

	helpLines = append(helpLines, ui.TitleStyle.Render("Actions"))
	helpLines = append(helpLines, "  Enter           Stage/unstage selected files")
	helpLines = append(helpLines, "  c               Commit staged files")
	helpLines = append(helpLines, "  m               Modify HEAD commit")
	helpLines = append(helpLines, "  p               Focus/unfocus preview pane")
	helpLines = append(helpLines, "  /               Search/filter files")
	helpLines = append(helpLines, "")

	helpLines = append(helpLines, ui.TitleStyle.Render("Git Status Symbols"))
	helpLines = append(helpLines, fmt.Sprintf("  %s   Staged file",
		ui.StagedStyle.Render("+")))
	helpLines = append(helpLines, fmt.Sprintf("  %s   Unstaged file",
		ui.UnstagedStyle.Render("-")))
	helpLines = append(helpLines, fmt.Sprintf("  %s   Untracked file",
		ui.UntrackedStyle.Render("?")))
	helpLines = append(helpLines, "")

	helpLines = append(helpLines, ui.TitleStyle.Render("Other"))
	helpLines = append(helpLines, "  ?               Toggle this help")
	helpLines = append(helpLines, "  q/Ctrl+C        Quit")

	content := strings.Join(helpLines, "\n")

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorBlue).
		Padding(1).
		Render(content)
}

// renderModifyHeadView renders the HEAD modification view
func (m Model) renderModifyHeadView() string {
	if m.processing {
		return lipgloss.NewStyle().Padding(1).Render("Processing...")
	}

	switch m.headModifyState {
	case HeadModifyStateMenu:
		return m.renderHeadModifyMenu()
	case HeadModifyStateAmendMessage:
		return m.renderHeadAmendMessageView()
	default:
		return m.renderHeadModifyMenu()
	}
}

// renderHeadModifyMenu renders the HEAD modify menu
func (m Model) renderHeadModifyMenu() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Title
	title := ui.TitleStyle.Render("Modify HEAD Commit")
	sections = append(sections, "", title, "")

	// HEAD info
	if m.headInfo != nil {
		headContent := fmt.Sprintf(
			"Current commit: %s\nMessage: %s\nAuthor: %s\nDate: %s",
			m.headInfo.ShortHash,
			m.headInfo.Message,
			m.headInfo.Author,
			m.headInfo.Date,
		)
		sections = append(sections, ui.PreviewStyle.Render(headContent), "")
	}

	// Menu options
	sections = append(sections, ui.TitleStyle.Render("Options:"))
	sections = append(sections, "  [m] Amend commit message")
	sections = append(sections, "  [f] Soft reset (modify files)")
	sections = append(sections, "")
	sections = append(sections, ui.HelpStyle.Render("[Esc] Cancel"))

	content := strings.Join(sections, "\n")
	return lipgloss.NewStyle().Padding(1).Render(content)
}

// renderHeadAmendMessageView renders the amend message input view
func (m Model) renderHeadAmendMessageView() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Title
	title := ui.TitleStyle.Render("Amend Commit Message")
	sections = append(sections, "", title, "")

	// Current message
	if m.headInfo != nil {
		sections = append(sections, "Current message:")
		sections = append(sections, ui.InfoStyle.Render(m.headInfo.Message))
		sections = append(sections, "")
	}

	// Message input
	sections = append(sections, ui.TitleStyle.Render("New Message:"))
	sections = append(sections, m.headMessageTextarea.View())
	sections = append(sections, "")
	sections = append(sections, ui.HelpStyle.Render("[Ctrl+D] Confirm  [Esc] Cancel"))

	content := strings.Join(sections, "\n")
	return lipgloss.NewStyle().Padding(1).Render(content)
}
