#!/bin/bash
# Interactive git staging/unstaging UI with fzf

# Check if in git repo
if ! git rev-parse --git-dir > /dev/null 2>&1; then
  echo "Error: Not in a git repository"
  exit 1
fi

git_interactive_add() {
  # ANSI color codes
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[0;33m'
  RESET='\033[0m'

  while true; do
    # Get status in parseable format
    staged=$(git diff --cached --name-only 2>/dev/null || echo "")
    unstaged=$(git diff --name-only 2>/dev/null || echo "")
    untracked=$(git ls-files --others --exclude-standard 2>/dev/null || echo "")

    # Build menu with indicators
    menu=""

    if [[ -n "$unstaged" ]]; then
      while IFS= read -r file; do
        menu+="${RED}- ${RESET}$file"$'\n'
      done <<< "$unstaged"
    fi

    if [[ -n "$staged" ]]; then
      while IFS= read -r file; do
        menu+="${GREEN}+ ${RESET}$file"$'\n'
      done <<< "$staged"
    fi

    if [[ -n "$untracked" ]]; then
      while IFS= read -r file; do
        menu+="${YELLOW}? ${RESET}$file"$'\n'
      done <<< "$untracked"
    fi
    
    # Remove trailing newline for cleaner display
    menu="${menu%$'\n'}"

    # Show status header
    clear
    width=$(tput cols)
    divider=$(printf "━%.0s" $(seq 1 $width))
    echo "$divider"
    printf "%*s\n" $(((5 + $width) / 2)) "gitUI"
    echo "$divider"
    # echo ""
    # echo -e " ${GREEN}[+] Staged${RESET}    ${RED}[-] Unstaged${RESET}    ${YELLOW}[?] Untracked${RESET}"

    # If no files to show, skip fzf and go straight to menu
    if [[ -z "$menu" ]]; then
      echo "✓ No changes to stage"
      echo ""
      selection=""
    else
      # Select files with fzf
      selection=$(echo -e "$menu" | \
        fzf \
          --ansi \
          --multi \
          --disabled \
          --prompt="> " \
          --preview='line={}; st=${line:0:1}; file=${line:2}; case "$st" in "+") /usr/bin/git diff --cached --color=always "$file";; "-") /usr/bin/git diff --color=always "$file";; "?") /usr/bin/head -100 "$file";; esac' \
          --preview-window=right:50% \
          --layout=reverse \
          --height=70% \
          --border \
          --info=default \
          --border-label-pos=bottom:center \
          --border-label="[Tab] select  [s] search  [Esc] exit  [Enter] apply  [q] menu" \
          --bind='tab:toggle,q:abort,s:enable-search+change-prompt([SEARCH] > ),esc:disable-search+change-prompt(> )' 2>/dev/null || true)
    fi
    
    # Process selected files (if any)
    if [[ -n "$selection" ]]; then
      echo ""
      echo "Processing..."

      while IFS= read -r line; do
        if [[ -z "$line" ]]; then
          continue
        fi

        status="${line:0:1}"
        filepath="${line:2}"

        case "$status" in
          "+")
            # Staged file -> unstage
            git reset HEAD -- "$filepath" > /dev/null 2>&1
            echo "  ✓ Unstaged: $filepath"
            ;;
          "-"|"?")
            # Unstaged/untracked file -> stage
            git add -- "$filepath" > /dev/null 2>&1
            echo "  ✓ Staged: $filepath"
            ;;
        esac
      done <<< "$selection"

      echo ""
    fi

    # Check if there are staged files and/or HEAD commit
    staged_count=$(git diff --cached --name-only 2>/dev/null | wc -l | tr -d ' ')
    has_commits=$(git rev-parse --verify HEAD > /dev/null 2>&1 && echo "yes" || echo "no")

    # Show menu if there are staged files or commits
    if [[ "$staged_count" -gt 0 ]] || [[ "$has_commits" == "yes" ]]; then
      if [[ "$staged_count" -gt 0 ]]; then
        echo -e "${GREEN}$staged_count file(s) staged${RESET}"
        echo ""
      fi
      echo "Options:"
      echo "  [s] Continue staging"
      if [[ "$staged_count" -gt 0 ]]; then
        echo "  [c] Commit staged files"
      fi
      if [[ "$has_commits" == "yes" ]]; then
        echo "  [m] Modify HEAD commit"
      fi
      echo "  [q] Quit"
      echo ""
      read -p "Choose option: " -n 1 choice
      echo ""

      case "$choice" in
        c|C)
          if [[ "$staged_count" -gt 0 ]]; then
            git_commit
          else
            echo "No files staged"
            sleep 1
          fi
          ;;
        m|M)
          if [[ "$has_commits" == "yes" ]]; then
            git_modify_head
          else
            echo "No commits to modify"
            sleep 1
          fi
          ;;
        q|Q)
          echo "Exiting..."
          return 0
          ;;
        s|S|*)
          # Continue loop (default)
          continue
          ;;
      esac
    else
      # No staged files and no commits - nothing to do
      echo "Nothing to commit. Exiting..."
      return 0
    fi
  done
}

git_commit() {
  clear
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Commit Staged Files"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  # Show staged files
  echo "Files to be committed:"
  git diff --cached --name-status | while read -r status file; do
    echo "  $status  $file"
  done
  echo ""

  # Get commit message
  echo "Enter commit message (Ctrl+C to cancel):"
  read -p "> " commit_msg

  if [[ -z "$commit_msg" ]]; then
    echo "Commit cancelled (empty message)"
    sleep 1
    return 1
  fi

  # Ask for custom date
  echo ""
  echo "Set custom commit date? (y/n):"
  read -p "> " -n 1 set_date
  echo ""

  commit_date=""
  if [[ "$set_date" =~ ^[Yy]$ ]]; then
    echo "Enter date (format: YYYY-MM-DD or YYYY-MM-DD HH:MM:SS, or 'now' for current):"
    read -p "> " commit_date

    if [[ -z "$commit_date" ]]; then
      echo "Invalid date, using current time"
      commit_date=""
    fi
  fi

  # Perform commit
  local commit_cmd="git commit -m \"$commit_msg\""
  if [[ -n "$commit_date" ]]; then
    commit_cmd="$commit_cmd --date=\"$commit_date\""
  fi

  if eval "$commit_cmd"; then
    echo ""
    echo -e "${GREEN}✓ Commit successful!${RESET}"
    sleep 2
  else
    echo ""
    echo -e "${RED}✗ Commit failed${RESET}"
    sleep 2
    return 1
  fi
}

git_modify_head() {
  # Check if HEAD has been pushed
  local current_branch=$(git branch --show-current)
  local is_pushed="no"

  if git branch -r --contains HEAD 2>/dev/null | grep -q "origin/$current_branch"; then
    is_pushed="yes"
  fi

  if [[ "$is_pushed" == "yes" ]]; then
    clear
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${RED}⚠ Warning: HEAD Already Pushed${RESET}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "The HEAD commit has already been pushed to the remote."
    echo "Modifying it will rewrite history and cause issues for others."
    echo ""
    echo -e "${RED}Operation blocked for safety.${RESET}"
    echo ""
    read -p "Press any key to continue..." -n 1
    return 1
  fi

  while true; do
    clear
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Modify HEAD Commit"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Show current commit info
    echo "Current commit:"
    git log -1 --pretty=format:"%C(yellow)%h%C(reset) - %s %C(green)(%cr)%C(reset) %C(blue)<%an>%C(reset)" HEAD
    echo ""
    echo ""

    echo "Options:"
    echo "  [d] View full diff"
    echo "  [m] Edit commit message"
    echo "  [f] Modify files in commit"
    echo "  [b] Back"
    echo ""
    read -p "Choose option: " -n 1 choice
    echo ""

    case "$choice" in
      d|D)
        git_view_head_diff
        ;;
      m|M)
        git_edit_commit_message
        ;;
      f|F)
        git_modify_commit_files
        ;;
      b|B)
        return 0
        ;;
    esac
  done
}

git_view_head_diff() {
  clear
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "HEAD Commit Diff"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  # Show commit info
  git log -1 --pretty=format:"%C(yellow)%h%C(reset) - %s %C(green)(%cr)%C(reset) %C(blue)<%an>%C(reset)" HEAD
  echo ""
  echo ""

  # Show full diff with color
  git show HEAD --color=always | less -R

  echo ""
  read -p "Press any key to continue..." -n 1
}

git_edit_commit_message() {
  clear
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Edit Commit Message"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  # Show current message
  echo "Current message:"
  git log -1 --pretty=format:"%s" HEAD
  echo ""
  echo ""

  # Get new message
  echo "Enter new commit message (leave empty to cancel):"
  read -p "> " new_msg

  if [[ -z "$new_msg" ]]; then
    echo "Edit cancelled"
    sleep 1
    return 1
  fi

  # Amend commit message
  if git commit --amend -m "$new_msg"; then
    echo ""
    echo -e "${GREEN}✓ Commit message updated!${RESET}"
    sleep 2
  else
    echo ""
    echo -e "${RED}✗ Failed to update message${RESET}"
    sleep 2
    return 1
  fi
}

git_modify_commit_files() {
  clear
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Modify Files in Commit"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  echo "This will:"
  echo "  1. Undo the HEAD commit (keeping changes staged)"
  echo "  2. Let you modify staged files"
  echo "  3. Re-commit with the original message"
  echo ""
  read -p "Continue? (y/n): " -n 1 confirm
  echo ""

  if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    sleep 1
    return 1
  fi

  # Store original commit message
  original_msg=$(git log -1 --pretty=format:"%s" HEAD)

  # Reset HEAD but keep changes staged
  if ! git reset --soft HEAD~1; then
    echo -e "${RED}✗ Failed to reset commit${RESET}"
    sleep 2
    return 1
  fi

  echo ""
  echo -e "${GREEN}✓ Commit undone, files are now staged${RESET}"
  echo ""
  echo "You can now modify the staged files using the main menu."
  echo "When ready, commit again to complete the amendment."
  echo ""
  echo "Original message: $original_msg"
  echo ""
  read -p "Press any key to return to main menu..." -n 1

  # Return to main loop - user will see staged files and can modify them
  return 0
}

# Run the function
git_interactive_add
