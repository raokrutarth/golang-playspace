# src control relate commands
alias c="clear"
alias gits="git status && git log -n5 --oneline"
alias gitncp='git commit -am "automation commit" && git status --untracked-files && git push origin HEAD && date +"%r"'
alias token='python -c "import secrets; import sys; print(secrets.token_urlsafe(int(sys.argv[1])))"'
alias psql="docker exec -it gops-dev-db psql -U app -d gops_db"
