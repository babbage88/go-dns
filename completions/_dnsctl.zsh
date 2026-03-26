# Zsh completion for dnsctl
#
# Installation:
#   Add this line to your .zshrc:
#   source <(dnsctl completion zsh)
#
# Or save the output to a file and source it:
#   dnsctl completion zsh > ~/.zsh/completions/_dnsctl
#   Then add to .zshrc:
#   fpath=(~/.zsh/completions $fpath)

_dnsctl() {
  local -a commands
  
  commands=(
    'clean-zones:Clean and validate DNS zones'
    'completion:Generate shell completion script'
  )
  
  if (( CURRENT == 2 )); then
    _describe -t commands 'dnsctl command' commands
    return
  fi
  
  case "${words[2]}" in
    clean-zones)
      _arguments \
        '(--file)--file[YAML file to process]:file:_files' \
        '(--timeout)--timeout[Ping timeout]:duration:(1s 2s 5s 10s)' \
        '(--workers)--workers[Number of parallel ping workers]:count:(1 2 4 8 16)' \
        '(--dry-run)--dry-run[Do not modify output]'
      ;;
    completion)
      _arguments '1: :(bash zsh)'
      ;;
  esac
}

compdef _dnsctl dnsctl
