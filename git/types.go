package git

// FileStatus represents the git status of a file
type FileStatus int

const (
	StatusStaged FileStatus = iota
	StatusUnstaged
	StatusUntracked
)

func (s FileStatus) String() string {
	switch s {
	case StatusStaged:
		return "staged"
	case StatusUnstaged:
		return "unstaged"
	case StatusUntracked:
		return "untracked"
	default:
		return "unknown"
	}
}

// GitStatus holds parsed git status information
type GitStatus struct {
	Staged      []string
	Unstaged    []string
	Untracked   []string
	Branch      string
	IsClean     bool
}

// CommitInfo holds HEAD commit information
type CommitInfo struct {
	Hash      string
	ShortHash string
	Message   string
	Author    string
	Date      string
	IsPushed  bool
}
