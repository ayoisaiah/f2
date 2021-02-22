<p align="center">
   <img src="https://ik.imagekit.io/turnupdev/F2_1__3LRCtY9uz.png" width="350" height="350" alt="f2">
</p>

<p align="center">
   <a href="https://goreportcard.com/report/github.com/ayoisaiah/f2"><img src="https://goreportcard.com/badge/github.com/ayoisaiah/f2" alt="Go Report Card"></a>
   <a href="https://www.codacy.com/manual/ayoisaiah/f2?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=ayoisaiah/f2&amp;utm_campaign=Badge_Grade"><img src="https://api.codacy.com/project/badge/Grade/7136493cf477467387381890cb25dc9e" alt="Codacy Badge"></a>
   <a href="http://hits.dwyl.com/ayoisaiah/f2"><img src="http://hits.dwyl.com/ayoisaiah/f2.svg" alt="HitCount"></a>
   <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat" alt=""></a>
   <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Github Actions"></a>
</p>

<h1 align="center">F2 - Command-line batch renaming tool</h1>

**F2** is a cross-platform command-line tool for batch renaming files and directories **quickly** and **safely**. Written in Go!

<img src="https://ik.imagekit.io/turnupdev/f2_5sH344M5q.png?tr:q-100" alt="Screenshot of F2 in action">


## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Command-line options](#command-line-options)
- [Examples](#examples)
   - [Basic find and replace](#basic-find-and-replace)
   - [Recursive find and replace](#recursive-find-and-replace)
   - [Include directories](#include-directories)
   - [Strip out unwanted text](#strip-out-unwanted-text)
   - [Rename using an auto incrementing number](#rename-using-an-auto-incrementing-number)
   - [Replace spaces with underscores](#replace-spaces-with-underscores)
   - [Use a variable](#use-a-variable)
   - [Conflict detection](#conflict-detection)
- [Credits](#credits)
- [Contribute](#contribute)
- [Licence](#licence)

## Features

- Extremely fast. Can work on 10,000 files in less than half a second.
- Supports Linux, macOS, and Windows.
- Supports filtering files using regular expression, including capture groups.
- Ignores hidden directories and files by default.
- Safe. F2 will not modify any file names until you tell it to.
- Detects potential conflicts such as file collisions, or overwrites.
- Supports recursive renaming of both files and directories.
- Supports renaming using a template.
- Supports using an ascending integer for renaming (e.g 001, 002, 003, e.t.c.).
- Supports undoing an operation from a map file.
- Extensive unit testing.

## Installation

F2 is written in Go, so you can install it through `go get` (Requires Go 1.16 or
later):

```bash
$ go get github.com/ayoisaiah/f2/cmd/...
```

Otherwise, you can download precompiled binaries for Linux, Windows, and macOS on the [releases page](https://github.com/ayoisaiah/f2/releases) (only for 64-bit machines).

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
   v1.0.0

FLAGS:
   --find string, -f string       Search string or regular expression.
   --replace string, -r string    Replacement string. If omitted, defaults to an empty string.
   --start-num value, -n value    Starting number when using numbering scheme in replacement string such as %03d (default: 1)
   --output-file value, -o value  Output a map file for the current operation
   --exec, -x                     Execute the batch renaming operation (default: false)
   --recursive, -R                Rename files recursively (default: false)
   --undo value, -u value         Undo a successful operation using a previously created map file
   --ignore-case, -i              Ignore case (default: false)
   --ignore-ext, -e               Ignore extension (default: false)
   --include-dir, -D              Rename directories (default: false)
   --hidden, -H                   Include hidden files and directories (default: false)
   --force, -F                    Force the renaming operation even when there are conflicts (may cause data loss). (default: false)
   --help, -h                     show help (default: false)
   --version, -v                  print the version (default: false)

WEBSITE:
  https://github.com/ayoisaiah/f2
```

## Examples

**Notes**:
- F2 does not make any changes to your filesystem by default (performs a dry run).
- To enforce the changes, include the `--exec` or `-x` flag.
- The `-f` or `--find` flag supports regular expressions.

### Basic find and replace

Replace all instances of `Screenshot` in the current directory with `Image`:

```bash
$ f2 -f "Screenshot" -r "Image"
+--------------------+---------------+--------+
|       INPUT        |    OUTPUT     | STATUS |
+--------------------+---------------+--------+
| Screenshot (1).png | Image (1).png | ok     |
| Screenshot (2).png | Image (2).png | ok     |
| Screenshot (3).png | Image (3).png | ok     |
+--------------------+---------------+--------+
```

### Recursive find and replace

Replace all instances of `js` to `ts` in the current directory and all sub directories (no depth limit).

```bash
$ f2 -f "js" -r "ts" -R
+---------------------+---------------------+--------+
|        INPUT        |       OUTPUT        | STATUS |
+---------------------+---------------------+--------+
| index-01.js         | index-01.ts         | ok     |
| index-02.js         | index-02.ts         | ok     |
| one/index-03.js     | one/index-03.ts     | ok     |
| one/index-04.js     | one/index-04.ts     | ok     |
| one/two/index-05.js | one/two/index-05.ts | ok     |
| one/two/index-06.js | one/two/index-06.ts | ok     |
+---------------------+---------------------+--------+
```

### Include directories

By default, directories are exempted from the renaming operation. Use the `-D`
flag to include them.

*Original tree*:

```plaintext
.
├── pic-1.avif
└── pics
    ├── pic-02.avif
    └── pic-03.avif
```

```bash
$ f2 -f "pic" -r "image" -D -x
```

*Renamed tree*:

```plaintext
.
├── image-1.avif
└── images
    ├── pic-02.avif
    └── pic-03.avif
```

### Strip out unwanted text

You can strip out text by leaving out the `-r` flag. It defaults to an empty string:

```bash
$ f2 -f "pic-"
+-------------+---------+--------+
|    INPUT    | OUTPUT  | STATUS |
+-------------+---------+--------+
| pic-02.avif | 02.avif | ok     |
| pic-03.avif | 03.avif | ok     |
+-------------+---------+--------+
```

### Rename using an auto incrementing number

You can specify an auto incrementing integer in the replacement string using the
format below:

  - `%d`: 1,2,3 e.t.c
  - `%02d`: 01, 02, 03, e.t.c.
  - `%03d`: 001, 002, 003, e.t.c.

```bash
$ f2 -f ".*\." -r "%03d."
+-----------------------------------------+---------+--------+
|                  INPUT                  | OUTPUT  | STATUS |
+-----------------------------------------+---------+--------+
| Screenshot from 2020-04-19 22-17-02.png | 001.png | ok     |
| Screenshot from 2020-04-19 23-17-02.png | 002.png | ok     |
| Screenshot from 2020-04-19 24-17-02.png | 003.png | ok     |
+-----------------------------------------+---------+--------+
```

You can also specify the number to start from using the `-n` flag:

```bash
$ f2 -f ".*\." -r "%03d." -n 20
+-----------------------------------------+---------+--------+
|                  INPUT                  | OUTPUT  | STATUS |
+-----------------------------------------+---------+--------+
| Screenshot from 2020-04-19 22-17-02.png | 020.png | ok     |
| Screenshot from 2020-04-19 23-17-02.png | 021.png | ok     |
| Screenshot from 2020-04-19 24-17-02.png | 022.png | ok     |
+-----------------------------------------+---------+--------+
```

### Replace spaces with underscores

```bash
$ f2 -f "\s" -r "_"
+--------------------+--------------------+--------+
|       INPUT        |       OUTPUT       | STATUS |
+--------------------+--------------------+--------+
| Screenshot (1).png | Screenshot_(1).png | ok     |
| Screenshot (2).png | Screenshot_(2).png | ok     |
| Screenshot (3).png | Screenshot_(3).png | ok     |
+--------------------+--------------------+--------+
Append the -x flag to apply the above changes
```

### Use a variable

The new file name can contain any number of variables that will be replaced with their corresponding value.

[Change to selecting the entire file]

The replacement string tokens may come in handy in template mode:

  - `{{f}}` is the original filename (excluding the extension)
  - `{{ext}}` is the file extension

For example:

This is helpful if you want to add a prefix or a suffix to a set of files:

## Conflict detection

  - F2 operates in print mode by default. Your filesystem remains the same until the `--exec` or `-x` flag is included. This allows you to verify any changes before proceeding.

  - If an operation will overwrite existing files, you will receive a warning. The `-F` or `--force` flag can be used to proceed anyway.

```bash
$ f2 --find "pic2" --replace "pic1-bad.jpg" -T -x
pic2-bad.png ➟ pic1-bad.jpg [File exists] ❌
Conflict detected: overwriting existing file(s)
Use the -F flag to ignore conflicts and rename anyway
```

  - If an operation results in two files having the same name, a warning will be printed. The `-F` or `--force` flag can be used to proceed anyway.

```bash
$ f2 --find "2020-04-16" --replace "screenshot.png" -T
Screenshot from 2020-04-16 18-25-15.png ➟ screenshot.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ screenshot.png ❌
Conflict detected: overwriting newly renamed path
Use the -F flag to ignore conflicts and rename anyway
```

  - If an operation results in a file having an empty filename, an error will be displayed.

```bash
$ f2 --find "pic1-bad.jpg" --replace ""
Error detected: Operation resulted in empty filename
pic1-bad.jpg ➟ [Empty filename] ❌
```

### Undo your changes

If you change your mind regarding a renaming operation, you can undo your changes using the `--undo` or `-U` flag. This only works for the last successful operation.

```bash
$ f2 -U
pic2-bad.png ➟ pic2-good.png ✅
pic1-bad.jpg ➟ pic1-good.jpg ✅
morebad/pic4-bad.webp ➟ morebad/pic4-good.webp ✅
morebad/pic3-bad.jpg ➟ morebad/pic3-good.jpg ✅
morebad ➟ moregood ✅
```

## Credits

F2 relies on other open source software listed below:

- [urfave/cli](https://github.com/urfave/cli)
- [gookit/color](https://github.com/gookit/color)
- [olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)

## Contribute

Bug reports and feature requests are much welcome! Please open an issue before creating a pull request.

## Licence

Created by Ayooluwa Isaiah and released under the terms of the [MIT Licence](http://opensource.org/licenses/MIT).
