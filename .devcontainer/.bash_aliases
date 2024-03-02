# src control relate commands
alias c="clear"
alias gits="git status && git log --oneline -n5"
alias gitca="git commit -a --amend --no-edit && git status --untracked-files"
alias gitpl="git pull origin main --rebase"
alias gitpf="git push origin HEAD --force && date +\"%r\""
alias gitncp='git commit -am "automation commit" && git status --untracked-files && git push origin HEAD && date +"%r"'