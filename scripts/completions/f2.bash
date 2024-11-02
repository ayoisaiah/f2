#!/usr/bin/env bash
f2_opts="
  --csv
  --find
  --replace
  --undo
  --allow-overwrites
  --clean
  --exclude
  --exclude-dir
  --exec
  --fix-conflicts
  --fix-conflicts-pattern
  --help
  --hidden
  --include-dir
  --ignore-case
  --ignore-ext
  --json
  --max-depth
  --no-color
  --only-dir
  --pair
  --pair-order
  --quiet
  --recursive
  --replace-limit
  --reset-index-per-dir
  --sort
  --sortr
  --sort-per-dir
  --sort-var
  --string-mode
  --target-dir
  --verbose
  --version
"
__f2_completions()
{
  cur="${COMP_WORDS[COMP_CWORD]}"
  COMPREPLY=($(compgen -W "${f2_opts}" -- "$cur"))
}

complete -F __f2_completions f2
