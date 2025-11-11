package ui

// Layout manages the split-pane layout calculations
type Layout struct {
	TotalWidth    int
	TotalHeight   int
	ListWidth     int
	PreviewWidth  int
	ContentHeight int
}

// NewLayout creates a new layout based on terminal size
func NewLayout(width, height int) Layout {
	l := Layout{
		TotalWidth:   width,
		TotalHeight:  height,
		ContentHeight: height - 5, // Subtract header (3) and footer (2)
	}

	// Minimum heights
	if l.ContentHeight < 5 {
		l.ContentHeight = 5
	}

	// Calculate widths
	// If width is small, disable preview pane
	if width < 100 {
		l.ListWidth = width - 2
		l.PreviewWidth = 0
	} else if width < 140 {
		// 50/50 split
		l.ListWidth = width / 2
		l.PreviewWidth = width - l.ListWidth - 2
	} else {
		// 40/60 split (more space for preview)
		l.ListWidth = (width * 2) / 5
		l.PreviewWidth = width - l.ListWidth - 2
	}

	// Ensure minimum widths
	if l.ListWidth < 30 {
		l.ListWidth = 30
		l.PreviewWidth = 0
	}
	if l.PreviewWidth < 30 {
		l.PreviewWidth = 0
	}

	return l
}

// HasPreviewPane returns true if there's space for the preview pane
func (l Layout) HasPreviewPane() bool {
	return l.PreviewWidth > 0
}

// ListHeight returns the height available for the file list
func (l Layout) ListHeight() int {
	return l.ContentHeight - 2 // Subtract borders and padding
}

// PreviewHeight returns the height available for the preview pane
func (l Layout) PreviewHeight() int {
	return l.ContentHeight - 2 // Subtract borders and padding
}
