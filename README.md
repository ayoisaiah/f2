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

<h1 align="center">F2 - Command-Line Batch Renaming</h1>

**F2** is a cross-platform command-line tool for batch renaming files and
directories **quickly** and **safely**. Written in Go!

## What does F2 do differently?

Compared to other renaming tools, F2 offers several key advantages:

- **Dry Run by Default**: It defaults to a dry run so that you can review the
  renaming changes before proceeding.

- **Variable Support**: F2 allows you to use file attributes, such as EXIF data
  for images or ID3 tags for audio files, to give you maximum flexibility in
  renaming.

- **Comprehensive Options**: Whether it's simple string replacements or complex
  regular expressions, F2 provides a full range of renaming capabilities.

- **Safety First**: It prioritizes accuracy by ensuring every renaming operation
  is conflict-free and error-proof through rigorous checks.

- **Conflict Resolution**: Each renaming operation is validated before execution
  and detected conflicts can be automatically resolved.

- **High Performance**: F2 is extremely fast and efficient, even when renaming
  thousands of files at once.

- **Undo Functionality**: Any renaming operation can be easily undone to allow
  the easy correction of mistakes.

- **Extensive Documentation**: F2 is well-documented with clear, practical
  examples to help you make the most of its features without confusion.

## ⚡ Installation

If you're a Go developer, F2 can be installed with `go install` (requires v1.23
or later):

```bash
go install github.com/ayoisaiah/f2/cmd/f2@latest
```

Other installation methods are
[documented here](https://f2.freshman.tech/guide/getting-started.html) or check
out the [releases page](https://github.com/ayoisaiah/f2/releases) to download a
pre-compiled binary for your operating system.

## 📃 Quick links

- [Installation](https://f2.freshman.tech/guide/getting-started.html)
- [Getting started tutorial](https://f2.freshman.tech/guide/tutorial.html)
- [Real-world example](https://f2.freshman.tech/guide/organizing-image-library.html)
- [Built-in variables](https://f2.freshman.tech/guide/how-variables-work.html)
- [File pair renaming](https://f2.freshman.tech/guide/pair-renaming.html)
- [Renaming with a CSV file](https://f2.freshman.tech/guide/csv-renaming.html)
- [Sorting](https://f2.freshman.tech/guide/sorting.html)
- [Resolving conflicts](https://f2.freshman.tech/guide/conflict-detection.html)
- [Undoing renaming mistakes](https://f2.freshman.tech/guide/undoing-mistakes.html)
- [CHANGELOG](https://f2.freshman.tech/reference/changelog.html)

## 💻 Screenshots

![F2 can utilise Exif attributes to organise image files](https://f2.freshman.tech/assets/2.D-uxLR9T.png)

## 🤝 Contribute

Bug reports and feature requests are much welcome! Please open an issue before
creating a pull request.

## ⚖ Licence

Created by Ayooluwa Isaiah and released under the terms of the
[MIT Licence](https://github.com/ayoisaiah/f2/blob/master/LICENCE).
