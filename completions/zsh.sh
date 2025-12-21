#compdef dnsctl

_arguments \
  '--timeout[Ping timeout]:duration' \
  '--workers[Number of parallel workers]:int' \
  '--dry-run[Do not modify output]' \
  'shellcompletion[Generate shell completion]:shell:(bash zsh)' \
  '*:file:_files'
