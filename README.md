<p align="center">
   <img src="https://ik.imagekit.io/turnupdev/F2_1__3LRCtY9uz.png" width="350" height="350" alt="f2">
</p>

<p align="center">
   <a href="https://www.codacy.com/manual/ayoisaiah/f2?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=ayoisaiah/f2&amp;utm_campaign=Badge_Grade"><img src="https://api.codacy.com/project/badge/Grade/7136493cf477467387381890cb25dc9e" alt="Codacy Badge"></a>
   <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat" alt=""></a>
   <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Github Actions"></a>
   <a href="https://golang.org"><img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="made-with-Go"></a>
   <a href="https://goreportcard.com/report/github.com/ayoisaiah/f2"><img src="https://goreportcard.com/badge/github.com/ayoisaiah/f2" alt="GoReportCard"></a>
   <a href="https://github.com/ayoisaiah/f2"><img src="https://img.shields.io/github/go-mod/go-version/ayoisaiah/f2.svg" alt="Go.mod version"></a>
   <a href="https://github.com/ayoisaiah/f2/blob/master/LICENCE"><img src="https://img.shields.io/github/license/ayoisaiah/f2.svg" alt="LICENCE"></a>
   <a href="https://github.com/ayoisaiah/f2/releases/"><img src="https://img.shields.io/github/release/ayoisaiah/f2.svg" alt="Latest release"></a>
</p>

<h1 align="center">F2 - Command-line batch renaming tool</h1>

**F2** is a cross-platform command-line tool for batch renaming files and directories **quickly** and **safely**. Written in Go!

<img src="https://ik.imagekit.io/turnupdev/f2_EsdXrHHKt.png?tr:q-100" alt="Screenshot of F2 in action">

## Why should I use F2?

- F2 helps you organise your filesystem through batch renaming so that your files and directories can have a consistent naming scheme.
- It offers a comprehensive set of renaming options and scales well from trivial string replacements to more complex operations involving regular expressions.
- F2 prioritises correctness and safety by ensuring that a renaming operation does not result in conflicts or errors. It runs several [validations](https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection) before carrying out a renaming operation and provides an easy way to automatically [fix any detected conflicts](https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection#auto-fixing-conflicts).
- F2 supports all the standard [renaming recipes](https://github.com/ayoisaiah/f2/wiki/Renaming-Recipes) including (but not limited to) string replacement, insertion of text as a prefix, suffix or other position in the file name, stripping a set of characters, changing the case of a set of letters, using auto incrementing numbers, swapping parts of the file name, e.t.c.
- F2 provides several [built-in variables](https://github.com/ayoisaiah/f2/wiki/Built-in-variables) for added flexibility in the renaming process. These variables are based on file attributes such as Exif information for images and ID3 tags for audio files.
- F2 is very fast and won't waste your time. See [benchmarks](#benchmarks).
- F2 allows you to [revert any renaming operation](https://github.com/ayoisaiah/f2/wiki/Undoing-a-renaming-operation) performed with the program. This means you don't have to worry about making a mistake because you can always get back to the previous state without breaking a sweat.
- F2 has good test coverage with equal attention paid to all supported platforms (Linux, Windows and macOS).
- F2 is [well documented](https://github.com/ayoisaiah/f2/wiki) so that you won't have to scratch your head while figuring out what you can do with it. Lots of examples are provided to aid comprehension.

## Main features

- Safe and transparent. F2 uses a dry run mode by default so you can review the exact changes that will be made to your filesystem before making them.
- Cross-platform with full support for Linux, macOS, and Windows. It also runs on less commonly-used platforms, like Termux (Android).
- Extremely fast, even when working with a large amount of files.
- Automatically detects potential conflicts such as file collisions, or overrides and reports them to you.
- Provides several built-in variables for the easier renaming of certain file types. At the moment, Exif data for images and ID3 data for audio files are supported.
- Supports find and replace using regular expressions, including capture groups.
- Ignores hidden directories and files by default.
- Supports recursive renaming for both files and directories.
- Supports using an ascending integer for renaming (e.g 001, 002, 003, e.t.c.).
- Supports undoing an operation from a map file.
- Extensive documentation and examples for each and every option that is provided.
- Extensive unit testing with close to 100% coverage.

## Benchmarks

**Environment**
- **OS**: Ubuntu 20.04.2 LTS on Windows 10 x86_64
- **CPU**: Intel i7-7560U (4) @ 2.400GHz
- **Kernel**:  4.19.128-microsoft-standard

Renaming **10,000** MP3 files using their ID3 attributes (~1.6 seconds):

```bash
$ hyperfine --warmup 3 'f2 -f ".*" -r "{{id3.artist}}_{{id3.album}}_{{id3.track}}_{{r}
}.mp3" -x'
Benchmark #1: f2 -f ".*" -r "{{id3.artist}}_{{id3.album}}_{{id3.track}}_{{r}}.mp3" -x
  Time (mean ± σ):      1.691 s ±  0.031 s    [User: 1.326 s, System: 0.744 s]
  Range (min … max):    1.634 s …  1.736 s    10 runs
```

Renaming **100,000** files in a single operation using random names (~5 seconds):

```bash
$ hyperfine --warmup 3 'f2 -f ".*" -r "{{r}}_%03d" -x'
Benchmark #1: f2 -f ".*" -r "{{r}}_%03d" -x
  Time (mean ± σ):      4.938 s ±  0.328 s    [User: 2.792 s, System: 2.770 s]
  Range (min … max):    4.421 s …  5.474 s    10 runs
```

Renaming **100,000** JPEG files using their Exif attributes (~30 seconds):

```bash
$ hyperfine --warmup 3 'f2 -f ".*" -r "{{x.make}}_{{x.model}}_{{x.iso}}_{{x.wh}}_{{r}}_%03d.jpg" -x'
Benchmark #1: f2 -f ".*" -r "{{x.make}}_{{x.model}}_{{x.iso}}_{{x.wh}}_{{r}}_%03d.jpg" -x
  Time (mean ± σ):     31.143 s ±  1.691 s    [User: 34.792 s, System: 4.779 s]
  Range (min … max):   29.317 s … 33.355 s    10 runs
```

## Installation

F2 is written in Go, so you can install it through `go install` (requires Go 1.16 or
later):

```bash
$ go install github.com/ayoisaiah/f2/cmd/f2@latest
```

You can also install it via `npm` if you have it installed:

```bash
$ npm i @ayoisaiah/f2 -g
```

Otherwise, you can download precompiled binaries for Linux, Windows, and macOS on the [releases page](https://github.com/ayoisaiah/f2/releases).

## Command-line options

This is the output of `f2 --help`:

```plaintext
DESCRIPTION:
  F2 is a command-line tool for batch renaming multiple files and directories quickly and safely

USAGE:
   f2 FLAGS [OPTIONS] [PATHS...]

AUTHOR:
   Ayooluwa Isaiah <ayo@freshman.tech>

VERSION:
   v1.5.3

FLAGS:
   --find <pattern>, -f <pattern>       Search <pattern>. Treated as a regular expression by default. Use -s or --string-mode to opt out
   --replace <string>, -r <string>      Replacement <string>. If omitted, defaults to an empty string. Supports built-in and regex capture variables
   --exclude <pattern>, -E <pattern>    Exclude files/directories that match the given find pattern. Treated as a regular expression. Multiple exclude <pattern>s can be specified.
   --exec, -x                           Execute the batch renaming operation (default: false)
   --recursive, -R                      Rename files recursively (default: false)
   --max-depth <integer>, -m <integer>  positive <integer> indicating the maximum depth for a recursive search (set to 0 for no limit) (default: 0)
   --undo, -u                           Undo the last operation performed in the current working directory. (default: false)
   --sort <sort>                        Sort the matches according to the provided <sort> (possible values: default, size, mtime, btime, atime, ctime)
   --sortr <sort>                       Same as <sort> but presents the matches in the reverse order (possible values: default, size, mtime, btime, atime, ctime)
   --ignore-case, -i                    Ignore case (default: false)
   --quiet, -q                          Don't print out any information including errors (default: false)
   --ignore-ext, -e                     Ignore extension (default: false)
   --include-dir, -d                    Include directories (default: false)
   --only-dir, -D                       Rename only directories (implies include-dir) (default: false)
   --hidden, -H                         Include hidden files and directories (default: false)
   --fix-conflicts, -F                  Fix any detected conflicts with auto indexing (default: false)
   --string-mode, -s                    Opt into string literal mode by treating find expressions as non-regex strings (default: false)
   --help, -h                           show help (default: false)
   --version, -v                        print the version (default: false)

DOCUMENTATION:
  https://github.com/ayoisaiah/f2/wiki

WEBSITE:
  https://github.com/ayoisaiah/f2
```

## Documentation

Visit the [wiki page](https://github.com/ayoisaiah/f2/wiki) to learn all the
operations that can be achieved with F2.

## Credits

F2 relies on other open source software listed below:

- [urfave/cli](https://github.com/urfave/cli)
- [gookit/color](https://github.com/gookit/color)
- [olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)
- [rwcarlsen/goexif](https://github.com/rwcarlsen/goexif)
- [djherbis/times](https://github.com/djherbis/times)
- [dhowden/tag](https://github.com/dhowden/tag)

## Contribute

Bug reports and feature requests are much welcome! Please open an issue before creating a pull request.

## Licence

Created by Ayooluwa Isaiah and released under the terms of the [MIT Licence](http://opensource.org/licenses/MIT).
