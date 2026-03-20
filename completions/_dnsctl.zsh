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
  local -a subcommands
  
  subcommands=(
    'clean-zones:Clean and validate DNS zones'
    'completion:Generate shell completion script'
  )
  
  _arguments \
    '1: :->command' \
    '*::arg:->args'
  
  case $state in
    command)
      _describe 'command' subcommands
      ;;
    args)
      case ${words[2]} in
        clean-zones)
          _arguments \
            '--file[YAML file to process]:file:_files' \
            '--timeout[Ping timeout (default: 2s)]:duration:(1s 2s 5s 10s)' \
            '--workers[Number of parallel ping workers (default: 8)]:count:(1 2 4 8 16)' \
            '--dry-run[Do not modify output]'
          ;;
        completion)
          _arguments '1: :(bash zsh)'
          ;;
      esac
      ;;
  esac
}

compdef _dnsctl dnsctl
