#compdef _f2 f2

function _f2 {
  local line

  _arguments -C \
    "--csv[Rename using a CSV file]" \
    "--find[Search for specified pattern]" \
    "-f[Search for specified pattern]" \
    "--replace[Replacement pattern for matches]" \
    "-r[Replacement pattern for matches]" \
    "--undo[Undo the last renaming operation in current directory]" \
    "-u[Undo the last renaming operation in current directory]" \
    "--allow-overwrites[Allow overwriting existing files]" \
    "--exclude[Exclude files and directories matching pattern]" \
    "-E[Exclude files and directories matching pattern]" \
    "--exec[Execute renaming operation]" \
    "-x[Execute renaming operation]" \
    "--fix-conflicts[Auto fix renaming conflicts]" \
    "-F[Auto fix renaming conflicts]" \
    "--help[Display help and exit]" \
    "-h[Display help and exit]" \
    "--hidden[Match hidden files]" \
    "-H[Match hidden files]" \
    "--include-dir[Match directories]" \
    "-d[Match directories]" \
    "--ignore-case[Make searches case insensitive]" \
    "-i[Make searches case insensitive]" \
    "--ignore-ext[Ignore file extension]" \
    "-e[Ignore file extension]" \
    "--json[Enable json output]" \
    "--max-depth[Specify max depth for recursive search]" \
    "-m[Specify max depth for recursive search]" \
    "--no-color[Disable coloured output]" \
    "--only-dir[Rename only directories]" \
    "-D[Rename only directories]" \
    "--quiet[Disable all output except errors]" \
    "-q[Disable all output except errors]" \
    "--recursive[Search for matches in subdirectories]" \
    "-R[Search for matches in subdirectories]" \
    "--replace-limit[Limit the matches to be replaced]" \
    "-R[Limit the matches to be replaced]" \
    "--sort[Sort matches in ascending order]" \
    "--sortr[Sort matches in descending order]" \
    "--string-mode[Treat the search pattern as a non-regex string]" \
    "-s[Treat the search pattern as a non-regex string]" \
    "--verbose[Enable verbose output]" \
    "-V[Enable verbose output]" \
    "--version[Display version and exit]" \
    "-v[Display version and exit]" \
}
