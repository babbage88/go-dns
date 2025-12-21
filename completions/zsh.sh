# Zsh completion for dnsctl
#
# Installation (recommended):
#   mkdir -p ~/.zsh/completions
#   dnsctl completion zsh > ~/.zsh/completions/_dnsctl
#
# Ensure the following is in your ~/.zshrc:
#   fpath+=~/.zsh/completions
#   autoload -Uz compinit
#   compinit
#
# Reload shell:
#   exec zsh
#
# Temporary (current shell only):
#   source <(dnsctl completion zsh)

#compdef dnsctl

_dnsctl() {
  _arguments \
    '1:command:(clean-zones completion)' \
    '*::arg:->args'

  case $state in
    args)
      case $words[2] in
        clean-zones)
          _arguments \
            '--file[Path to YAML file]:file:_files' \
            '--timeout[Ping timeout]' \
            '--workers[Number of workers]' \
            '--dry-run[Do not write output]'
          ;;
        completion)
          _arguments \
            '1:shell:(bash zsh)'
          ;;
      esac
    ;;
  esac
}

compdef _dnsctl dnsctl
