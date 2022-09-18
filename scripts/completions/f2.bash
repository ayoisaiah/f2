#!/usr/bin/env bash
f2_opts="
  --csv
  --find
  --replace
  --undo
  --allow-overwrites
  --exclude
  --exec
  --fix-conflicts
  --help
  --hidden
  --include-dir
  --ignore-case
  --ignore-ext
  --json
  --max-depth
  --no-color
  --only-dir
  --quiet
  --recursive
  --replace-limit
  --sort
  --sortr
  --string-mode
  --verbose
  --version
"
__f2_completions()
{
  cur="${COMP_WORDS[COMP_CWORD]}"
  COMPREPLY=($(compgen -W "${f2_opts}" -- "$cur"))
}

complete -F __f2_completions f2
