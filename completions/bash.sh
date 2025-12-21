_dnsctl_complete() {
	local cur prev opts
	COMPREPLY=()
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[COMP_CWORD-1]}"

	opts="--timeout --workers --dry-run --help shellcompletion"

	if [[ ${cur} == -* ]]; then
		COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
		return 0
	fi
}
complete -F _dnsctl_complete dnsctl
