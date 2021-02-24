<p align="center">
   <img src="https://ik.imagekit.io/turnupdev/F2_1__3LRCtY9uz.png" width="350" height="350" alt="f2">
</p>

<p align="center">
   <a href="https://www.codacy.com/manual/ayoisaiah/f2?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=ayoisaiah/f2&amp;utm_campaign=Badge_Grade"><img src="https://api.codacy.com/project/badge/Grade/7136493cf477467387381890cb25dc9e" alt="Codacy Badge"></a>
   <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat" alt=""></a>
   <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Github Actions"></a>
</p>

<h1 align="center">F2 - Command-line batch renaming tool</h1>

**F2** is a cross-platform command-line tool for batch renaming files and directories **quickly** and **safely**. Written in Go!

<img src="https://ik.imagekit.io/turnupdev/f2_5sH344M5q.png?tr:q-100" alt="Screenshot of F2 in action">

## Table of Contents

- [Features](#features)
- [Benchmarks](#benchmarks)
- [Installation](#installation)
- [Command-line options](#command-line-options)
- [Examples](#examples)
   - [Basic find and replace](#basic-find-and-replace)
   - [Recursive find and replace](#recursive-find-and-replace)
   - [Include directories](#include-directories)
   - [Ignore extensions](#ignore-extensions)
   - [Strip out unwanted text](#strip-out-unwanted-text)
   - [Rename using an auto incrementing number](#rename-using-an-auto-incrementing-number)
   - [Replace spaces with underscores](#replace-spaces-with-underscores)
   - [Use regex capture variables](#use-regex-capture-variables)
   - [Use a variable](#use-a-variable)
   - [Directories are auto created if necessary](#directories-are-auto-created-if-necessary)
   - [Conflict detection](#conflict-detection)
   - [Undoing changes](#undoing-changes)
- [Credits](#credits)
- [Contribute](#contribute)
- [Licence](#licence)

## Features

- Extremely fast (see [benchmarks](#benchmarks)).
- Supports Linux, macOS, and Windows.
- Supports filtering files using regular expression, including capture groups.
- Ignores hidden directories and files by default.
- Safe. F2 will not modify any file names until you tell it to.
- Detects potential conflicts such as file collisions, or overwrites.
- Supports recursive renaming of both files and directories.
- Supports using an ascending integer for renaming (e.g 001, 002, 003, e.t.c.).
- Supports undoing an operation from a map file.
- Extensive unit testing.

## Benchmarks

Recursive batch renaming of 10,000 files from pic-{n}.png to {n}.png.

**Versions**:
- [f2](https://github.com/ayoisaiah/f2) (Go) — v1.0.0
- [rnm](https://github.com/neurobin/rnm) (C++) — v4.0.9
- [rnr](https://github.com/ChuckDaniels87/rnr) (Rust) — v0.3.0
- [brename](https://github.com/shenwei356/brename) (Go) — v2.11.0

**Environment**:
- **OS**: Ubuntu 20.04.2 LTS on Windows 10 x86_64
- **CPU**: Intel i7-7560U (4) @ 2.400GHz
- **Kernel**:  4.19.128-microsoft-standard

Preparation script: [prepare-script.sh](https://gist.github.com/ayoisaiah/868437602e73084ebc11efcec262e92c).

```bash
$ hyperfine --warmup 3 --prepare "./prepare-script.sh" 'rnr -sfr "pic-(\d+).*" "\$1.png" dir1/' 'f2 -f "pic-(\d+).*" -r \$1.png -x -R' 'brename -p "pic-(\d+).*" -r \$1.png -q -R' 'rnm -q -rs "/pic-(\d+).*$/\1.png/" dir1/ -dp -1'
Benchmark #1: rnr -sfr "pic-(\d+).*" "\$1.png" dir1/
  Time (mean ± σ):     944.1 ms ±  20.2 ms    [User: 238.5 ms, System: 666.5 ms]
  Range (min … max):   916.0 ms … 982.2 ms    10 runs

Benchmark #2: f2 -f "pic-(\d+).*" -r \$1.png -x -R
  Time (mean ± σ):     292.4 ms ±  11.8 ms    [User: 141.8 ms, System: 217.4 ms]
  Range (min … max):   276.4 ms … 311.0 ms    10 runs

Benchmark #3: brename -p "pic-(\d+).*" -r \$1.png -q -R
  Time (mean ± σ):     602.3 ms ±  10.8 ms    [User: 202.1 ms, System: 311.7 ms]
  Range (min … max):   587.2 ms … 626.9 ms    10 runs

Benchmark #4: rnm -q -rs "/pic-(\d+).*$/\1.png/" dir1/ -dp -1
  Time (mean ± σ):     821.0 ms ±  43.2 ms    [User: 564.6 ms, System: 254.7 ms]
  Range (min … max):   783.9 ms … 926.2 ms    10 runs

Summary
  'f2 -f "pic-(\d+).*" -r \$1.png -x -R' ran
    2.06 ± 0.09 times faster than 'brename -p "pic-(\d+).*" -r \$1.png -q -R'
    2.81 ± 0.19 times faster than 'rnm -q -rs "/pic-(\d+).*$/\1.png/" dir1/ -dp -1'
    3.23 ± 0.15 times faster than 'rnr -sfr "pic-(\d+).*" "\$1.png" dir1/'
```

## Installation

F2 is written in Go, so you can install it through `go get` (requires Go 1.16 or
later):

```bash
$ go get -u github.com/ayoisaiah/f2/cmd/...
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
   v1.1.0

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
   --include-dir, -d              Include directories (default: false)
   --only-dir, -D                 Rename only directories (implies include-dir) (default: false)
   --hidden, -H                   Include hidden files and directories (default: false)
   --fix-conflicts, -F            Fix any detected conflicts with auto indexing (default: false)
   --help, -h                     show help (default: false)
   --version, -v                  print the version (default: false)

WEBSITE:
  https://github.com/ayoisaiah/f2
```

## Examples

**Notes**:
- F2 does not make any changes to your filesystem by default (performs a dry run).
- To enforce the changes, include the `--exec` or `-x` flag.
- The `-f` or `--find` flag supports regular expressions. If omitted, it matches the entire filename of each file.
- The `-r` or `--replace` flag supports variables

### Basic find and replace

Replace all instances of `Screenshot` in the current directory with `Image`:

```bash
$ f2 -f 'Screenshot' -r 'Image'
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
$ f2 -f 'js' -r 'ts' -R
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

Directories are exempted from the renaming operation by default. Use the `-d` or
`--include-dir` flag to include them.

*Original tree*:

```plaintext
.
├── pic-1.avif
└── pics
    ├── pic-02.avif
    └── pic-03.avif
```

```bash
$ f2 -f 'pic' -r 'image' -d -x
```

*Renamed tree*:

```plaintext
.
├── image-1.avif
└── images
    ├── pic-02.avif
    └── pic-03.avif
```

You can also rename only directories by using the `-D` or `--only-dir` flag:

*Original tree*:

```plaintext
.
├── pic-1.avif
└── pics
    ├── pic-02.avif
    └── pic-03.avif
```

```bash
$ f2 -f 'pic' -r 'image' -D -x
```

*Renamed tree*:

```plaintext
.
├── pic-1.avif
└── images
    ├── pic-02.avif
    └── pic-03.avif
```

### Ignore extensions

The file extension is matched by default. If this behaviour is not desired, use
the `--ignore-ext` or `-e` flag:

```bash
$ ls
a-jpeg-file.jpeg file.jpeg
```

```bash
$ f2 -f "jpeg" -r "jpg" -e
+------------------+-----------------+--------+
|      INPUT       |     OUTPUT      | STATUS |
+------------------+-----------------+--------+
| a-jpeg-file.jpeg | a-jpg-file.jpeg | ok     |
+------------------+-----------------+--------+
```

### Strip out unwanted text

You can strip out text by leaving out the `-r` flag. It defaults to an empty string:

```bash
$ f2 -f 'pic-'
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
$ f2 -f '.*\.' -r '%03d.'
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
$ f2 -f '.*\.' -r '%03d.' -n 20
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
$ f2 -f '\s' -r '_'
+--------------------+--------------------+--------+
|       INPUT        |       OUTPUT       | STATUS |
+--------------------+--------------------+--------+
| Screenshot (1).png | Screenshot_(1).png | ok     |
| Screenshot (2).png | Screenshot_(2).png | ok     |
| Screenshot (3).png | Screenshot_(3).png | ok     |
+--------------------+--------------------+--------+
```

### Ignore cases

Use the `-i` or `--ignore-case` flag:

```bash
$ f2 -f 'jpeg' -r 'jpg' -i
+--------+--------+--------+
| INPUT  | OUTPUT | STATUS |
+--------+--------+--------+
| a.JPEG | a.jpg  | ok     |
| b.jpeg | b.jpg  | ok     |
| c.jPEg | c.jpg  | ok     |
+--------+--------+--------+
```

### Use regex capture variables

Regex capture variables are supported:

```bash
$ f2 -f '.* S(\d+).E(\d+).*.(mp4)' -r 'S$1 E$2.$3'
+--------------------------------------+-------------+--------+
|                INPUT                 |   OUTPUT    | STATUS |
+--------------------------------------+-------------+--------+
| No Pressure (2021) S01.E01.2160p.mp4 | S01 E01.mp4 | ok     |
| No Pressure (2021) S01.E02.2160p.mp4 | S01 E02.mp4 | ok     |
| No Pressure (2021) S01.E03.2160p.mp4 | S01 E03.mp4 | ok     |
+--------------------------------------+-------------+--------+
```

```bash
$ f2 -f '(\w+) \((\d+)\).(\w+)' -r '$2-$1.$3'
+--------------------+------------------+--------+
|       INPUT        |      OUTPUT      | STATUS |
+--------------------+------------------+--------+
| Screenshot (1).png | 1-Screenshot.png | ok     |
| Screenshot (2).png | 2-Screenshot.png | ok     |
| Screenshot (3).png | 3-Screenshot.png | ok     |
+--------------------+------------------+--------+
```

### Directories are auto created if necessary

Assuming the following directory:

```bash
$ ls
x-y-z.pdf
```

```bash
$ f2 -f '-' -r '/' -x
```

*Result*

```bash
.
└── x
    └── y
        └── z.pdf
```

### Use a variable

The replacement string can contain the following variables that will be replaced with their corresponding value.

  - `{{f}}` is the original filename (excluding the extension)
  - `{{ext}}` is the file extension

This is helpful if you want to add a prefix or a suffix to a set of files:

```bash
$ f2 -r '{{f}}_journal{{ext}}' # suffix
+-------------------+---------------------------+--------+
|       INPUT       |          OUTPUT           | STATUS |
+-------------------+---------------------------+--------+
| 2021-02-20.md     | 2021-02-20_journal.md     | ok     |
| 2021-02-21.md     | 2021-02-21_journal.md     | ok     |
| 2021-02-22.md     | 2021-02-22_journal.md     | ok     |
+-------------------+---------------------------+--------+
```

```bash
$ f2 -r 'journal_{{f}}{{ext}}' # prefix
+-------------------+---------------------------+--------+
|       INPUT       |          OUTPUT           | STATUS |
+-------------------+---------------------------+--------+
| 2021-02-20.md     | journal_2021-02-20.md     | ok     |
| 2021-02-21.md     | journal_2021-02-21.md     | ok     |
| 2021-02-22.md     | journal_2021-02-22.md     | ok     |
+-------------------+---------------------------+--------+
```

*More variables coming soon*

### Conflict detection

F2 detects any conflicts that may arise during a renaming operation. If you append the `-F` or `--fix-conflicts` flag, it can auto fix the conflicts for you. Here are three examples:

#### 1. File already exists

```bash
$ ls
a.txt b.txt
```

```bash
$ f2 -f 'a' -r 'b'
+-------+--------+--------------------------+
| INPUT | OUTPUT |          STATUS          |
+-------+--------+--------------------------+
| a.txt | b.txt  | ❌ [Path already exists] |
+-------+--------+--------------------------+
```

You can append the `-F` flag to fix the conflict. This will add a number to the target file to differentiate it from the existing one.

```bash
$ f2 -f "a" -r "b" -F
+-------+-----------+--------+
| INPUT |  OUTPUT   | STATUS |
+-------+-----------+--------+
| a.txt | b (2).txt | ok     |
+-------+-----------+--------+
```

#### 2. Overwriting newly renamed path

```bash
$ ls
a.txt b.txt
```

```bash
$ f2 -f 'a|b' -r 'c'
+-------+--------+-------------------------------------+
| INPUT | OUTPUT |               STATUS                |
+-------+--------+-------------------------------------+
| a.txt | c.txt  | ❌ [Overwriting newly renamed path] |
| b.txt | c.txt  | ❌ [Overwriting newly renamed path] |
+-------+--------+-------------------------------------+
```

You can append the `-F` flag to fix the conflict. This will add a number to the target path to differentiate it from the

```bash
$ f2 -f 'a|b' -r 'c' -F
+-------+-----------+--------+
| INPUT |  OUTPUT   | STATUS |
+-------+-----------+--------+
| a.txt | c.txt     | ok     |
| b.txt | c (2).txt | ok     |
+-------+-----------+--------+
```

#### 3. Empty filename

```bash
$ ls
a.txt b.txt
```

```bash
$ f2 -f 'a.txt'
+-------+--------+---------------------+
| INPUT | OUTPUT |       STATUS        |
+-------+--------+---------------------+
| a.txt |        | ❌ [Empty filename] |
+-------+--------+---------------------+
```

You can append the `-F` flag to fix the conflict. The filename will not be changed in this context.

```bash
$ f2 -f 'a.txt' -F
+-------+--------+--------+
| INPUT | OUTPUT | STATUS |
+-------+--------+--------+
| a.txt | a.txt  | ok     |
+-------+--------+--------+
```

### Undoing changes

Before you can undo a change, you must create a map file before making the
change:

```bash
$ ls
a.txt b.txt
```

```bash
$ f2 -f 'txt' -r 'md' -o map.json -x
+-------+--------+--------+
| INPUT | OUTPUT | STATUS |
+-------+--------+--------+
| a.txt | a.md   | ok     |
| b.txt | b.md   | ok     |
+-------+--------+--------+
```

```bash
$ ls
a.md b.md map.json
```

```bash
$ f2 -u map.json
+-------+--------+--------+
| INPUT | OUTPUT | STATUS |
+-------+--------+--------+
| a.md  | a.txt  | ok     |
| b.md  | b.txt  | ok     |
+-------+--------+--------+
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
