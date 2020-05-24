# Goname - Command-line batch renaming tool
[![Go Report Card](https://goreportcard.com/badge/github.com/ayoisaiah/goname)](https://goreportcard.com/report/github.com/ayoisaiah/goname)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/7136493cf477467387381890cb25dc9e)](https://www.codacy.com/manual/ayoisaiah/goname?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=ayoisaiah/goname&amp;utm_campaign=Badge_Grade)
[![HitCount](http://hits.dwyl.com/ayoisaiah/goname.svg)](http://hits.dwyl.com/ayoisaiah/goname)
[![PR's Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com)

Goname is a cross-platform command-line tool for batch renaming files and directories safely. Written in Go!

## Features

  - Supports Linux, macOS, and Windows.
  - Supports filtering files using regular expression.
  - Safe by default. Goname will not modify any file names until you tell it to.
  - Supports piping files through other programs like `find` or `rg`.
  - Detects potential conflicts and errors and reports them to you.
  - Supports recursive renaming of both files and directories.
  - Supports renaming using a template
  - Supports using an ascending integer for renaming (e.g 001, 002, 003, e.t.c.).

## Installation

If you have Go installed, you can install `goname` using the command below:

```bash
$ go get github.com/ayoisaiah/goname/cmd/...
```

Otherwise, you can download precompiled binaries for Linux, Windows, and macOS [here](https://github.com/ayoisaiah/goname/releases) (only for amd64).

## Usage

**Note**: running these commands will only print out the changes to be made. If you want to proceed, use the `-x` flag.

All the examples below assume the following directory structure:

```bash
.
├── morebad
│   ├── pic3-bad.jpg
│   └── pic4-bad.webp
├── pic1-bad.jpg
├── pic2-bad.png
├── Screenshot from 2020-04-16 18-25-15.png
├── Screenshot from 2020-04-16 18-27-24.png
├── Screenshot from 2020-04-19 22-17-02.png
├── Screenshot from 2020-04-23 01-07-22.png
├── Screenshot from 2020-05-10 08-51-16.png
└── Screenshot from 2020-05-20 23-29-50.png
```

### Basic find and replace

Replace all instances of `Screenshot from ` in the current directory with `IMG`:

```bash
$ goname --find "Screenshot from " --replace "IMG-"
Screenshot from 2020-04-23 01-07-22.png ➟ IMG-2020-04-23 01-07-22.png ✅
Screenshot from 2020-04-19 22-17-02.png ➟ IMG-2020-04-19 22-17-02.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ IMG-2020-04-16 18-27-24.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ IMG-2020-05-10 08-51-16.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ IMG-2020-05-20 23-29-50.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ IMG-2020-04-16 18-25-15.png ✅
```

### Recursive find and replace

Replace all instances of `bad` to `good` in the current directory and sub
directories.

```bash
$ goname --find "bad" --replace "good" **
morebad/pic3-bad.jpg ➟ morebad/pic3-good.jpg ✅
morebad/pic4-bad.webp ➟ morebad/pic4-good.webp ✅
pic1-bad.jpg ➟ pic1-good.jpg ✅
pic2-bad.png ➟ pic2-good.png ✅
```

### Include directories

By default, directories are exempted from the renaming operation. Use the `-D`
flag to include them:

```bash
$ goname --find "bad" --replace "good" -D **
pic2-bad.png ➟ pic2-good.png ✅
pic1-bad.jpg ➟ pic1-good.jpg ✅
morebad/pic4-bad.webp ➟ morebad/pic4-good.webp ✅
morebad/pic3-bad.jpg ➟ morebad/pic3-good.jpg ✅
morebad ➟ moregood ✅
```

### Operate on directories only

Use the `**/` pattern to operate only on directories and subdirectories. The `-D` flag also needs to be present:

```bash
$ goname --find "bad" --replace "good" -D **/
morebad ➟ moregood ✅
```

### Strip out unwanted text

You can strip out text by leaving out the `--replace` flag. It defaults to an
empty string:

```bash
$ goname --find "Screenshot from "
Screenshot from 2020-04-19 22-17-02.png ➟ 2020-04-19 22-17-02.png ✅
Screenshot from 2020-04-23 01-07-22.png ➟ 2020-04-23 01-07-22.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ 2020-04-16 18-25-15.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ 2020-05-20 23-29-50.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ 2020-05-10 08-51-16.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ 2020-04-16 18-27-24.png ✅
```

### Rename using an auto incrementing number

You can specify an auto incrementing integer in the replacement string using the
format below:

  - `%d`: 1,2,3 e.t.c
  - `%02d`: 01, 02, 03, e.t.c.
  - `%03d`: 001, 002, 003, e.t.c.

If you want to padd the number with ten zeros, use `%010d`. And so on.

```bash
$ goname --find "Screenshot from " --replace "IMG-%03d_"
Screenshot from 2020-04-19 22-17-02.png ➟ IMG-001_2020-04-19 22-17-02.png ✅
Screenshot from 2020-04-23 01-07-22.png ➟ IMG-002_2020-04-23 01-07-22.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ IMG-003_2020-04-16 18-25-15.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ IMG-004_2020-05-20 23-29-50.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ IMG-005_2020-05-10 08-51-16.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ IMG-006_2020-04-16 18-27-24.png ✅
```

### Use a template

You can use the replacement string as a template for the new filenames instead of replacing the matched text in the original. Use `-T` or `--template-mode` to opt in.

The replacement string tokens may come in handy in template mode:

  - `{og}` is the original filename (excluding the extension)
  - `{ext}` is the file extension

For example:

```bash
$ goname --find "Screenshot from " --replace "Screenshot-%03d{ext}" -T
Screenshot from 2020-04-19 22-17-02.png ➟ Screenshot-001.png ✅
Screenshot from 2020-04-23 01-07-22.png ➟ Screenshot-002.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ Screenshot-003.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ Screenshot-004.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ Screenshot-005.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ Screenshot-006.png ✅
```

## Safe guards

  - Your filesystem remains the same until the `--exec` or `-x` flag is included. This allows you to verify the changes before proceeding.

  - If an operation will overwrite existing files, you will recieve a warning. The `-F` or `--force` flag can be used to proceed anyway.

```bash
$ goname --find "pic2" --replace "pic1-bad.jpg" -T
pic2-bad.png ➟ pic1-bad.jpg [File exists] ❌
Conflict detected: overwriting existing file(s)
Use the -F flag to ignore conflicts and rename anyway
```

  - If an operation results in two files having the same name, a warning will be printed. The `-F` or `--force` flag can be used to proceed anyway.

```bash
$ goname --find "2020-04-16" --replace "screenshot.png" -T
Screenshot from 2020-04-16 18-25-15.png ➟ screenshot.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ screenshot.png ❌
Conflict detected: overwriting newly renamed path
Use the -F flag to ignore conflicts and rename anyway
```

  - If an operation results in a file having an empty filename, an error will be displayed.

```bash
$ goname --find "pic1-bad.jpg" --replace ""
Error detected: Operation resulted in empty filename
pic1-bad.jpg ➟ [Empty filename] ❌
```

## TODO

- [ ] Write tests
- [ ] Add undo support
- [ ] Override starting integer for auto incrementing numbers in filenames


## Credit and sources

Goname relies heavily on other open source software listed below:

  - [urfave/cli](https://github.com/urfave/cli)
  - [gookit/color](https://github.com/gookit/color)

## Contribute

Bug reports, or pull requests are much welcome!

## Licence

Created by Ayooluwa Isaiah and released under the terms of the [MIT Licence](http://opensource.org/licenses/MIT).
