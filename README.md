<p align="center">
   <img src="https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png" width="250" height="250" alt="f2">
</p>

<p align="center">
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

<img src="https://ik.imagekit.io/turnupdev/f2-demo_tnCZlpBrkhX.gif?tr:q-100" alt="F2 in action on Ubuntu Linux">

## ⚡ Installation

F2 is written in Go, so you can install it through `go install` (requires Go 1.16 or later):

```bash
$ go install github.com/ayoisaiah/f2/cmd/f2@latest
```

### 📦 NPM package

You can also install F2 via its [npm package](https://www.npmjs.com/package/@ayoisaiah/f2):

With `npm`:

```bash
$ npm i @ayoisaiah/f2 -g
```

With `yarn`:

```bash
$ yarn global add @ayoisaiah/f2
```

Other installation methods are [available here](https://github.com/ayoisaiah/f2/wiki/Installation/).

## 💥 Why should you use F2?

- F2 helps you organise your filesystem through batch renaming so that your files and directories can have a consistent naming scheme.
- It offers a comprehensive set of renaming options, and scales well from trivial string replacements to more complex operations involving regular expressions.
- F2 prioritises correctness and safety by ensuring that a renaming operation does not result in conflicts or errors. It runs several [validations](https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection) before carrying out a renaming operation, and provides an easy way to automatically [fix any detected conflicts](https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection#auto-fixing-conflicts).
- F2 supports all the standard [renaming recipes](https://github.com/ayoisaiah/f2/wiki/Renaming-Recipes) including (but not limited to) string replacement, insertion of text as a prefix, suffix or other position in the file name, stripping a set of characters, changing the case of a set of letters, using auto incrementing numbers, swapping parts of the file name, e.t.c.
- F2 provides several [built-in variables](https://github.com/ayoisaiah/f2/wiki/Built-in-variables) for added flexibility in the renaming process. These variables are based on file attributes such as Exif information for images, and ID3 tags for audio files. F2 also supports utilising tags from the popular [exiftool](https://github.com/ayoisaiah/f2/wiki/Exiftool-variables) which should cover most use cases.
- F2 is very fast and won't waste your time. See [benchmarks](#benchmarks).
- F2 allows you to [revert any renaming operation](https://github.com/ayoisaiah/f2/wiki/Undoing-a-renaming-operation) performed with the program. This means you don't have to worry about making a mistake because you can always get back to the previous state without breaking a sweat.
- F2 has good test coverage with equal attention paid to all supported platforms (Linux, Windows and macOS).
- F2 is [well documented](https://github.com/ayoisaiah/f2/wiki) so that you won't have to scratch your head while figuring out what you can do with it. Lots of [realistic examples](https://github.com/ayoisaiah/f2/wiki/Real-world-examples) are provided to aid comprehension.

## ✨ Main features

- Safe and transparent. F2 uses a dry run mode by default so you can review the exact changes that will be made to your filesystem before making them.
- Cross-platform with full support for Linux, macOS, and Windows. It also runs on less commonly-used platforms, like Termux (Android).
- [Extremely fast](#benchmarks), even when working with a large amount of files.
- Supports the chaining of several consecutive renaming operations before a final output is produced.
- Automatically [detects potential conflicts](https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection) such as file collisions, or overrides and reports them to you.
- Provides several [built-in variables](https://github.com/ayoisaiah/f2/wiki/Built-in-variables) for the easier renaming of certain file types.
- Provides easy access to all ~25,000 tags in [exiftool](https://github.com/ayoisaiah/f2/wiki/Exiftool-variables) for maximum flexibility in renaming.
- Supports find and replace using [regular expressions](https://github.com/ayoisaiah/f2/wiki/Regular-expressions), including capture groups.
- Ignores hidden directories and files by default.
- Respects the [`NO_COLOR`](https://no-color.org/) environmental variable.
- Supports limiting the number of replaced matches, and you can start from the beginning or end of the file name.
- Supports recursive renaming for both files and directories.
- Supports renaming only files, or only directories, or both.
- Supports using an ascending integer for renaming (e.g 001, 002, 003, e.t.c.), and it can be formatted in several ways.
- Supports [undoing](https://github.com/ayoisaiah/f2/wiki/Undoing-a-renaming-operation) the last renaming operation in case of mistakes or errors.
- Supports renaming from a [CSV file](https://github.com/ayoisaiah/f2/wiki/Renaming-from-a-CSV-file).
- Extensive [documentation](https://github.com/ayoisaiah/f2/wiki) and examples for each option that is provided.

## 💻 Screenshots

![Screenshot of F2 in action on Linux](https://ik.imagekit.io/turnupdev/f2_EsdXrHHKt.png?tr:q-100)

![F2 can utilise exif attributes from a variety of image formats](https://ik.imagekit.io/turnupdev/f2-exif-example_Xn_Ko8ait.png?tr:q-100)

![F2 can utilise ID3 attributes to organise music files](https://ik.imagekit.io/turnupdev/f2-id3-example_Esb--IK6A.png?tr:q-100)

## 📃 Documentation

Visit the [wiki page](https://github.com/ayoisaiah/f2/wiki) to view usage examples and learn about all the renaming operations that can be achieved with F2.

## 🔥 Benchmarks

**Environment**
- **OS**: Ubuntu 20.04.2 LTS on Windows 10 x86_64
- **CPU**: Intel i7-7560U (4) @ 2.400GHz
- **Kernel**:  4.19.128-microsoft-standard

Renaming **10,000** MP3 files using their ID3 attributes (~1.6 seconds):

```bash
$ hyperfine --warmup 3 'f2 -f ".*" -r "{{id3.artist}}_{{id3.album}}_{{id3.track}}_{{r} }.mp3" -x'
Benchmark #1: f2 -f ".*" -r "{{id3.artist}}_{{id3.album}}_{{id3.track}}_{{r}}.mp3" -x
  Time (mean ± σ):      1.691 s ±  0.031 s    [User: 1.326 s, System: 0.744 s]
  Range (min … max):    1.634 s …  1.736 s    10 runs
```

Renaming **100,000** files using a random string and an auto incrementing
integer (~5 seconds):

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

### Windows

Renaming **10,000** MP3 files with an auto incrementing integer through native PowerShell commands (~30 seconds):

```bash
$ Measure-Command { Get-ChildItem *.mp3 | ForEach-Object -Begin { $count = 1 } -Process { Rename-Item $_ -NewName "music_$count.mp3"; $count++ } }
Days              : 0
Hours             : 0
Minutes           : 0
Seconds           : 29
Milliseconds      : 582
Ticks             : 295824810
TotalDays         : 0.000342389826388889
TotalHours        : 0.00821735583333333
TotalMinutes      : 0.49304135
TotalSeconds      : 29.582481
TotalMilliseconds : 29582.481
```

Renaming **10,000** MP3 files with an auto incrementing integer through F2 (~12 seconds):

```bash
$ Measure-Command { f2 -f ".*" -r "audio_%03d.mp3" -x }
Days              : 0
Hours             : 0
Minutes           : 0
Seconds           : 11
Milliseconds      : 634
Ticks             : 116342215
TotalDays         : 0.000134655341435185
TotalHours        : 0.00323172819444444
TotalMinutes      : 0.193903691666667
TotalSeconds      : 11.6342215
TotalMilliseconds : 11634.2215
```

## 🤝 Contribute

Bug reports and feature requests are much welcome! Please open an issue before creating a pull request.

## ⚖ Licence

Created by Ayooluwa Isaiah and released under the terms of the [MIT Licence](http://opensource.org/licenses/MIT).
