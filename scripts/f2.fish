complete --command f2 --condition 'not __fish_should_complete_switches' --exclusive --long-option csv --description "Rename using a CSV file" --keep-order --arguments '(__fish_complete_suffix .csv)'

complete --command f2 --long-option find --short-option f --description "Search for specified pattern" --exclusive

complete --command f2 --long-option replace --short-option r --description "Replacement pattern for matches" --exclusive

complete --command f2 --long-option undo --short-option u --description "Undo the last renaming operation in current directory" --no-files

complete --command f2 --long-option allow-overwrites --description "Allow overwriting existing files" --no-files

complete --command f2 --long-option exclude --short-option E --description "Exclude files and directories matching pattern" --no-files

complete --command f2 --long-option exec --short-option x --description "Execute renaming operation" --no-files

complete --command f2 --long-option fix-conflicts --short-option F --description "Auto fix renaming conflicts" --no-files

complete --command f2 --long-option help --short-option h --description "Display help and exit" --no-files

complete --command f2 --long-option hidden --short-option H --description "Match hidden files" --no-files

complete --command f2 --long-option include-dir --short-option d --description "Match directories" --no-files

complete --command f2 --long-option ignore-case --short-option i --description "Make searches case insensitive" --no-files

complete --command f2 --long-option ignore-ext --short-option e --description "Ignore file extension" --no-files

complete --command f2 --long-option json --description "Enable json output" --no-files

complete --command f2 --long-option max-depth --short-option m --description "Specify max depth for recursive search" --no-files

complete --command f2 --long-option no-color --description "Disable coloured output" --no-files

complete --command f2 --long-option only-dir --short-option D --description "Rename only directories" --no-files

complete --command f2 --long-option quiet --short-option q --description "Disable all output except errors" --no-files

complete --command f2 --long-option recursive --short-option R --description "Search for matches in subdirectories" --no-files

complete --command f2 --long-option replace-limit --short-option l --description "Limit the matches to be replaced" --no-files

set -l sort_args "
  default\t'Alphabetical order'
  size\t'Sort by file size'
  mtime\t'Sort by file last modified time'
  btime\t'Sort by file creation time'
  atime\t'Sort by file last access time'
  ctime\t'Sort by file metadata last change time'
"

complete --command f2 --long-option sort --description "Sort matches in ascending order" --exclusive --keep-order --arguments $sort_args

complete --command f2 --long-option sortr --description "Sort matches in descending order" --exclusive --keep-order --arguments $sort_args

complete --command f2 --long-option string-mode --short-option s --description "Treat the search pattern as a non-regex string" --no-files

complete --command f2 --long-option verbose --short-option V --description "Enable verbose output" --no-files

complete --command f2 --long-option version --short-option v --description "Display version and exit" --no-files

