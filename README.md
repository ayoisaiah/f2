# F2 - Command-line batch renaming tool

[![Go Report Card](https://goreportcard.com/badge/github.com/ayoisaiah/f2)](https://goreportcard.com/report/github.com/ayoisaiah/f2)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/7136493cf477467387381890cb25dc9e)](https://www.codacy.com/manual/ayoisaiah/f2?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=ayoisaiah/f2&amp;utm_campaign=Badge_Grade)
[![HitCount](http://hits.dwyl.com/ayoisaiah/f2.svg)](http://hits.dwyl.com/ayoisaiah/f2)
[![PR's Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com)

F2 is a cross-platform command-line tool for batch renaming files and directories **safely** and **speedily**. Written in Go!

## Features

- Supports Linux, macOS, and Windows.
- Supports filtering files using regular expression, including capture groups.
- Exclude or include dotfiles.
- Safe by default. F2 will not modify any file names until you tell it to.
- Supports piping files through other programs such as `find` or `rg`.
- Detects potential conflicts and errors and reports them to you.
- Supports recursive renaming of both files and directories.
- Supports renaming using a template.
- Supports using an ascending integer for renaming (e.g 001, 002, 003, e.t.c.).
- Supports undoing the last successful operation.
- Extensive unit testing.

## Installation

F2 is written in Go so you can build it from source with `go install`:

```bash
$ go get github.com/ayoisaiah/f2/cmd/...
```

Otherwise, you can download precompiled binaries for Linux, Windows, and macOS on the [releases page](https://github.com/ayoisaiah/f2/releases) (only for 64-bit machines).

## Examples

**Note**: F2 does not make any changes to your filesystem by default. It only prints out the results of the operation. To enforce the changes, use the `-x` or `--exec` flag.

### Basic find and replace

Replace all instances of `Screenshot from ` in the current directory with `IMG`:

```bash
$ f2 --find "Screenshot from " --replace "IMG-"
Screenshot from 2020-04-23 01-07-22.png ➟ IMG-2020-04-23 01-07-22.png ✅
Screenshot from 2020-04-19 22-17-02.png ➟ IMG-2020-04-19 22-17-02.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ IMG-2020-04-16 18-27-24.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ IMG-2020-05-10 08-51-16.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ IMG-2020-05-20 23-29-50.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ IMG-2020-04-16 18-25-15.png ✅
```

[You can also pass a list of files to be renamed as arguments]

### Recursive find and replace

Replace all instances of `bad` to `good` in the current directory and sub
directories.

```bash
$ f2 --find "bad" --replace "good" **
morebad/pic3-bad.jpg ➟ morebad/pic3-good.jpg ✅
morebad/pic4-bad.webp ➟ morebad/pic4-good.webp ✅
pic1-bad.jpg ➟ pic1-good.jpg ✅
pic2-bad.png ➟ pic2-good.png ✅
```

### Include directories

By default, directories are exempted from the renaming operation. Use the `-D`
flag to include them:

```bash
$ f2 --find "bad" --replace "good" -D **
pic2-bad.png ➟ pic2-good.png ✅
pic1-bad.jpg ➟ pic1-good.jpg ✅
morebad/pic4-bad.webp ➟ morebad/pic4-good.webp ✅
morebad/pic3-bad.jpg ➟ morebad/pic3-good.jpg ✅
morebad ➟ moregood ✅
```

### Operate on directories only

Use the `**/` pattern to operate only on directories and subdirectories. The `-D` flag also needs to be present:

```bash
$ f2 --find "bad" --replace "good" -D **/
morebad ➟ moregood ✅
```

### Strip out unwanted text

You can strip out text by leaving out the `--replace` flag. It defaults to an
empty string:

```bash
$ f2 --find "Screenshot from "
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

```bash
$ f2 --find "Screenshot from " --replace "IMG-%03d_"
Screenshot from 2020-04-19 22-17-02.png ➟ IMG-001_2020-04-19 22-17-02.png ✅
Screenshot from 2020-04-23 01-07-22.png ➟ IMG-002_2020-04-23 01-07-22.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ IMG-003_2020-04-16 18-25-15.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ IMG-004_2020-05-20 23-29-50.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ IMG-005_2020-05-10 08-51-16.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ IMG-006_2020-04-16 18-27-24.png ✅
```

You can also specify the number to start from using the `--start-num` flag:

```bash
$ f2 --find "Screenshot from " --replace "IMG-%03d_" --start-num 20
Screenshot from 2020-04-19 22-17-02.png ➟ IMG-020_2020-04-19 22-17-02.png ✅
Screenshot from 2020-04-23 01-07-22.png ➟ IMG-021_2020-04-23 01-07-22.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ IMG-022_2020-04-16 18-25-15.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ IMG-023_2020-05-20 23-29-50.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ IMG-024_2020-05-10 08-51-16.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ IMG-025_2020-04-16 18-27-24.png ✅
```

### Use a template

[Change to selecting the entire file]

You can use the replacement string as a template for the new filenames instead of replacing the matched text in the original. Use `-T` or `--template-mode` to opt in.

The replacement string tokens may come in handy in template mode:

  - `{og}` is the original filename (excluding the extension)
  - `{ext}` is the file extension

For example:

```bash
$ f2 --find "Screenshot from " --replace "Screenshot-%03d{ext}" -T
Screenshot from 2020-04-19 22-17-02.png ➟ Screenshot-001.png ✅
Screenshot from 2020-04-23 01-07-22.png ➟ Screenshot-002.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ Screenshot-003.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ Screenshot-004.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ Screenshot-005.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ Screenshot-006.png ✅
```

This is helpful if you want to add a prefix or a suffix to a set of files:

```bash
# prefix
$ f2 -f "pic" -r "lagos-{og}{ext}" -T
pic1-bad.jpg ➟ lagos-pic1-bad.jpg ✅
pic2-bad.png ➟ lagos-pic2-bad.png ✅

# suffix
$ f2 -f "pic" -r "{og}_ios{ext}" -T
pic1-bad.jpg ➟ pic1-bad_ios.jpg ✅
pic2-bad.png ➟ pic2-bad_ios.png ✅
```

## Regular expression examples

The `--find` flag can accept regular expressions in addition to plain text.
Here's a few examples:

### Strip out whitespace

Use `\s` to match whitespace in a string. Leaving out the `--replace` flag will
default to an empty string.

```bash
f2 --f "\s"
Screenshot from 2020-04-19 22-17-02.png ➟ Screenshotfrom2020-04-1922-17-02.png ✅
Screenshot from 2020-04-23 01-07-22.png ➟ Screenshotfrom2020-04-2301-07-22.png ✅
Screenshot from 2020-04-16 18-25-15.png ➟ Screenshotfrom2020-04-1618-25-15.png ✅
Screenshot from 2020-05-20 23-29-50.png ➟ Screenshotfrom2020-05-2023-29-50.png ✅
Screenshot from 2020-05-10 08-51-16.png ➟ Screenshotfrom2020-05-1008-51-16.png ✅
Screenshot from 2020-04-16 18-27-24.png ➟ Screenshotfrom2020-04-1618-27-24.png ✅
```

## Safe guards

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

  - If you change your mind regarding a renaming operation, you can undo your changes using the `--undo` or `-U` flag. This only works for the last successful operation.

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

Bug reports, feature requests, or pull requests are much welcome!

## Licence

Created by Ayooluwa Isaiah and released under the terms of the [MIT Licence](http://opensource.org/licenses/MIT).
