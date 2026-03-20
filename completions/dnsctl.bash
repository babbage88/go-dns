# Bash completion for dnsctl
#
# Installation (recommended):
#   dnsctl completion bash > /etc/bash_completion.d/dnsctl
#
# Or for current user only:
#   dnsctl completion bash > ~/.bash_completion
#
# Reload shell:
#   source ~/.bashrc
#
# Temporary (current shell only):
#   source <(dnsctl completion bash)

_dnsctl() {
  local cur prev
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"

  case "${COMP_WORDS[1]}" in
    clean-zones)
      COMPREPLY=( $(compgen -W "--file --timeout --workers --dry-run" -- "$cur") )
      ;;
    completion)
      COMPREPLY=( $(compgen -W "bash zsh" -- "$cur") )
      ;;
    *)
      COMPREPLY=( $(compgen -W "clean-zones completion" -- "$cur") )
      ;;
  esac
}

complete -F _dnsctl dnsctl
