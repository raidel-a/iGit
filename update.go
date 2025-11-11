package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"

	"github.com/rai/interactive-git/ui"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.err != "" {
			// If there's an error, only allow quitting or dismissing
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			}
			return m, nil
		}

		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Recalculate layout
		m.layout = ui.NewLayout(m.width, m.height)

		// Calculate shared pane height for split mode
		paneHeight := m.layout.ListHeight()
		// Viewport height: paneHeight - border (2) - title line (1)
		viewportHeight := paneHeight - 3
		if viewportHeight < 1 {
			viewportHeight = 1
		}

		// Adjust list size based on layout
		// Subtract 4 for border (2) + padding (2)
		if m.layout.HasPreviewPane() && m.showPreview {
			m.list.SetWidth(m.layout.ListWidth - 4)
			m.viewport.Width = m.layout.PreviewWidth - 4
		} else {
			m.list.SetWidth(m.width - 4)
			m.viewport.Width = m.width - 4
		}
		m.list.SetHeight(paneHeight)
		m.viewport.Height = viewportHeight

		// Fetch initial diff for current file
		if m.showPreview && len(m.files) > 0 {
			currentFile := m.getCurrentFile()
			if currentFile != nil {
				return m, m.fetchDiffCmd(*currentFile)
			}
		}

		return m, nil

	case gitStatusMsg:
		m.gitStatus = msg.status
		m.files = msg.status.AllFiles()

		// Properly set items in the list
		// Create a slice of list.Item interface
		var listFileItems []list.Item
		for _, f := range m.files {
			listFileItems = append(listFileItems, f)
		}
		m.list.SetItems(listFileItems)

		// Ensure list has a selection (defaults to -1, needs to be 0)
		if m.list.Index() < 0 && len(m.files) > 0 {
			m.list.Select(0)
		}

		// Fetch initial diff for first file
		if m.showPreview && len(m.files) > 0 && m.ready && m.list.Index() >= 0 {
			m.lastFileIndex = m.list.Index()
			currentFile := m.getCurrentFile()
			if currentFile != nil {
				return m, m.fetchDiffCmd(*currentFile)
			}
		}

		m.lastFileIndex = -1
		return m, nil

	case errorMsg:
		m.err = msg.err
		if msg.err == "" {
			return m, nil
		}
		return m, m.clearError()

	case statusMsg:
		m.status = msg.msg
		if msg.msg == "" {
			return m, nil
		}
		return m, m.clearStatus()

	case gitStageMsg:
		m.processing = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, m.clearError()
		}
		m.status = fmt.Sprintf("Staged %d file(s)", len(msg.files))
		// Clear selection after staging
		m.deselectAll()
		return m, tea.Batch(m.refreshStatus(), m.clearStatus())

	case gitUnstageMsg:
		m.processing = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, m.clearError()
		}
		m.status = fmt.Sprintf("Unstaged %d file(s)", len(msg.files))
		// Clear selection after unstaging
		m.deselectAll()
		return m, tea.Batch(m.refreshStatus(), m.clearStatus())

	case gitRefreshMsg:
		return m, m.refreshStatus()

	case gitDiffMsg:
		if msg.err != nil {
			m.previewContent = fmt.Sprintf("Error loading diff: %v", msg.err)
		} else {
			m.previewContent = msg.content
		}
		m.viewport.SetContent(m.previewContent)
		return m, nil

	case gitCommitMsg:
		if msg.err != nil {
			m.err = fmt.Sprintf("Commit failed: %v", msg.err)
			return m, m.clearError()
		}
		m.status = msg.message
		m.state = StateFileList
		m.commitMessage = ""
		m.commitDate = ""
		return m, tea.Batch(m.refreshStatus(), m.clearStatus())

	case gitHeadInfoMsg:
		m.headInfo = msg.info
		return m, nil

	case gitAmendMsg:
		if msg.err != nil {
			m.err = fmt.Sprintf("Amendment failed: %v", msg.err)
			return m, m.clearError()
		}
		m.status = msg.message
		m.state = StateFileList
		m.headInfo = nil
		return m, tea.Batch(m.refreshStatus(), m.clearStatus())
	}

	// Handle list updates
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// If list index changed and preview is shown, fetch new diff
	if m.showPreview && m.ready && m.state == StateFileList {
		currentIndex := m.list.Index()
		if currentIndex >= 0 && currentIndex != m.lastFileIndex {
			m.lastFileIndex = currentIndex
			currentFile := m.getCurrentFile()
			if currentFile != nil {
				// Clear old content and fetch new diff
				m.previewContent = ""
				return m, m.fetchDiffCmd(*currentFile)
			}
		}
	}

	// Handle viewport updates
	_, vpCmd := m.viewport.Update(msg)

	return m, tea.Batch(cmd, vpCmd)
}

// handleKeyMsg handles key messages
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.state {
	case StateFileList:
		return m.handleFileListKeys(msg)
	case StateCommitMessage, StateCommitDate:
		return m.handleCommitKeys(msg)
	case StateModifyHead:
		return m.handleModifyHeadKeys(msg)
	case StateHelp:
		return m.handleHelpKeys(msg)
	default:
		return m.handleFileListKeys(msg)
	}
}

// handleFileListKeys handles keys in the file list view
func (m Model) handleFileListKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Select):
		// Toggle selection of current item
		m.toggleSelection(m.list.Index())
		return m, nil

	case key.Matches(msg, m.keys.SelectAll):
		m.selectAll()
		return m, nil

	case key.Matches(msg, m.keys.Deselect):
		m.deselectAll()
		return m, nil

	case key.Matches(msg, m.keys.TogglePreview):
		// Toggle focus between list and preview
		if m.showPreview && m.layout.HasPreviewPane() {
			m.previewFocused = !m.previewFocused
		}
		return m, nil

	case key.Matches(msg, m.keys.ToggleHelp):
		if m.state == StateFileList {
			m.state = StateHelp
		} else {
			m.state = StateFileList
		}
		return m, nil

	case key.Matches(msg, m.keys.Up):
		// If preview is focused, scroll up; otherwise navigate list
		if m.previewFocused && m.viewport.Height < len(strings.Split(m.previewContent, "\n")) {
			m.viewport.LineUp(3)
			return m, nil
		}
		// Let list handle navigation and fetch new diff if selection changed
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		currentIndex := m.list.Index()
		if m.showPreview && currentIndex >= 0 && currentIndex != m.lastFileIndex {
			m.lastFileIndex = currentIndex
			if currentFile := m.getCurrentFile(); currentFile != nil {
				m.previewContent = ""
				return m, tea.Batch(cmd, m.fetchDiffCmd(*currentFile))
			}
		}
		return m, cmd

	case key.Matches(msg, m.keys.Down):
		// If preview is focused, scroll down; otherwise navigate list
		if m.previewFocused && m.viewport.Height < len(strings.Split(m.previewContent, "\n")) {
			m.viewport.LineDown(3)
			return m, nil
		}
		// Let list handle navigation and fetch new diff if selection changed
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		currentIndex := m.list.Index()
		if m.showPreview && currentIndex >= 0 && currentIndex != m.lastFileIndex {
			m.lastFileIndex = currentIndex
			if currentFile := m.getCurrentFile(); currentFile != nil {
				m.previewContent = ""
				return m, tea.Batch(cmd, m.fetchDiffCmd(*currentFile))
			}
		}
		return m, cmd

	case key.Matches(msg, m.keys.Home):
		// Let list handle Home key and fetch diff if selection changed
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		currentIndex := m.list.Index()
		if m.showPreview && currentIndex >= 0 && currentIndex != m.lastFileIndex {
			m.lastFileIndex = currentIndex
			if currentFile := m.getCurrentFile(); currentFile != nil {
				m.previewContent = ""
				return m, tea.Batch(cmd, m.fetchDiffCmd(*currentFile))
			}
		}
		return m, cmd

	case key.Matches(msg, m.keys.End):
		// Let list handle End key and fetch diff if selection changed
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		currentIndex := m.list.Index()
		if m.showPreview && currentIndex >= 0 && currentIndex != m.lastFileIndex {
			m.lastFileIndex = currentIndex
			if currentFile := m.getCurrentFile(); currentFile != nil {
				m.previewContent = ""
				return m, tea.Batch(cmd, m.fetchDiffCmd(*currentFile))
			}
		}
		return m, cmd

	case key.Matches(msg, m.keys.Apply):
		selected := m.getSelectedFiles()
		if len(selected) == 0 {
			m.status = "No files selected"
			return m, m.clearStatus()
		}
		m.status = fmt.Sprintf("Processing %d file(s)...", len(selected))
		return m, m.applySelection()

	case key.Matches(msg, m.keys.Commit):
		if m.gitStatus.StagedCount() == 0 {
			m.status = "No files staged"
			return m, m.clearStatus()
		}
		m.enterCommitMode()
		return m, nil

	case key.Matches(msg, m.keys.ModifyHead):
		m.enterModifyHeadMode()
		m.processing = true
		return m, m.fetchHeadInfo()

	default:
		return m, nil
	}
}

// handleCommitKeys handles keys during commit input
func (m Model) handleCommitKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.commitState {
	case CommitStateMessage:
		return m.handleCommitMessageKeys(msg)
	case CommitStateDate:
		return m.handleCommitDateKeys(msg)
	default:
		return m, nil
	}
}

// handleCommitMessageKeys handles keys for commit message input
func (m Model) handleCommitMessageKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+d":
		// Proceed to date input
		m.commitMessage = m.commitTextarea.Value()
		if m.commitMessage == "" {
			m.err = "Commit message cannot be empty"
			return m, m.clearError()
		}
		m.proceedToDateInput()
		return m, nil

	case "esc":
		// Cancel commit
		m.cancelCommit()
		return m, nil

	default:
		// Handle textarea input
		var cmd tea.Cmd
		m.commitTextarea, cmd = m.commitTextarea.Update(msg)
		return m, cmd
	}
}

// handleCommitDateKeys handles keys for commit date input
func (m Model) handleCommitDateKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Proceed to commit
		m.commitDate = m.commitInput.Value()
		m.commitInput.Blur()
		m.commitTextarea.Blur()
		return m, m.commitCmd(m.commitMessage, m.commitDate)

	case "esc":
		// Go back to message input
		m.commitState = CommitStateMessage
		m.commitTextarea.Focus()
		m.commitInput.Reset()
		return m, nil

	default:
		// Handle text input
		var cmd tea.Cmd
		m.commitInput, cmd = m.commitInput.Update(msg)
		return m, cmd
	}
}

// handleHelpKeys handles keys in the help view
func (m Model) handleHelpKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.ToggleHelp), key.Matches(msg, m.keys.Quit):
		m.state = StateFileList
		return m, nil
	default:
		return m, nil
	}
}

// handleModifyHeadKeys handles keys during HEAD modification
func (m Model) handleModifyHeadKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.headModifyState {
	case HeadModifyStateMenu:
		return m.handleHeadMenuKeys(msg)
	case HeadModifyStateAmendMessage:
		return m.handleHeadAmendMessageKeys(msg)
	case HeadModifyStateAmendFiles:
		return m.handleHeadAmendFilesKeys(msg)
	default:
		return m, nil
	}
}

// handleHeadMenuKeys handles keys in the HEAD modify menu
func (m Model) handleHeadMenuKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "m":
		// Amend commit message
		m.enterAmendMessageMode()
		return m, nil

	case "f":
		// Soft reset (amend files)
		m.processing = true
		return m, m.softResetHeadCmd()

	case "esc", "q":
		// Cancel and return to file list
		m.cancelModifyHead()
		return m, nil

	default:
		return m, nil
	}
}

// handleHeadAmendMessageKeys handles keys for commit message amendment
func (m Model) handleHeadAmendMessageKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+d":
		// Confirm amendment
		newMessage := m.headMessageTextarea.Value()
		if newMessage == "" {
			m.err = "Commit message cannot be empty"
			return m, m.clearError()
		}
		m.processing = true
		m.headMessageTextarea.Blur()
		return m, m.amendMessageCmd(newMessage)

	case "esc":
		// Cancel and return to menu
		m.headModifyState = HeadModifyStateMenu
		m.headMessageTextarea.Blur()
		m.headMessageTextarea.Reset()
		return m, nil

	default:
		// Handle textarea input
		var cmd tea.Cmd
		m.headMessageTextarea, cmd = m.headMessageTextarea.Update(msg)
		return m, cmd
	}
}

// handleHeadAmendFilesKeys handles keys for soft reset
func (m Model) handleHeadAmendFilesKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Soft reset is automatic, just return to file list or show menu again
	m.cancelModifyHead()
	return m, nil
}
