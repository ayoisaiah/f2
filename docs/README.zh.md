**以其他語言閱讀此文檔：**[English](/README.md) | [Deutsch](docs/README.de.md) | [Español](docs/README.es.md) | [Français](docs/README.fr.md) | [Português](docs/README.pt.md) | [Русский](docs/README.ru.md)

<p align="center">
   <img src="https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png" width="250" height="250" alt="f2">
</p>
<p align="center">
   <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat" alt="歡迎PR"></a>
   <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Github Actions"></a>
   <a href="https://golang.org"><img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="Go語言開發"></a>
   <a href="https://goreportcard.com/report/github.com/ayoisaiah/f2"><img src="https://goreportcard.com/badge/github.com/ayoisaiah/f2" alt="GoReportCard"></a>
   <a href="https://github.com/ayoisaiah/f2"><img src="https://img.shields.io/github/go-mod/go-version/ayoisaiah/f2.svg" alt="Go.mod版本"></a>
   <a href="https://github.com/ayoisaiah/f2/blob/master/LICENCE"><img src="https://img.shields.io/github/license/ayoisaiah/f2.svg" alt="MIT許可證"></a>
   <a href="https://github.com/ayoisaiah/f2/releases/"><img src="https://img.shields.io/github/release/ayoisaiah/f2.svg" alt="最新版本"></a>
</p>
<h1 align="center">F2 - 命令行批次重新命名工具</h1>

**F2** 是一款跨平台的命令行工具，用於**快速**、**安全**地批次重新命名檔案和目錄。採用 Go 語言編寫！

## F2 有何不同？

相比其他重新命名工具，F2 具備以下核心優勢：

* **預設模擬執行**：預設開啟模擬運行模式，方便你在執行前預覽更改。
* **支援檔案屬性變數**：允許使用檔案屬性（如圖片的 EXIF 數據、音頻的 ID3 標籤）進行重新命名，靈活性極高。
* **功能全面**：無論是簡單的字串替換還是複雜的正則表達式匹配，F2 都能滿足。
* **安全第一**：通過嚴格檢查確保每次操作無衝突、防誤操作，優先保證準確性。
* **衝突自動處理**：執行前驗證操作，可自動檢測並解決衝突。
* **性能卓越**：處理成千上萬檔案依然快速高效。
* **操作可撤銷**：輕鬆回退任何重新命名操作，修正錯誤更方便。
* **文件詳盡**：提供清晰實用的示例文件，助你快速掌握功能，避免困惑。

## ⚡ 安裝

Go 開發者可通過 `go install` 安裝（需要 Go v1.23 或更高版本）：

```bash
go install github.com/ayoisaiah/f2/v2/cmd/f2@latest
```

其他安裝方式[請查閱文檔](https://f2.freshman.tech/guide/getting-started.html)，或前往 [Release](https://github.com/ayoisaiah/f2/releases) 下載對應作業系統的預編譯二進位文件。

## 📃 快速連結

* [安裝指南](https://f2.freshman.tech/guide/getting-started.html)
* [入門教程](https://f2.freshman.tech/guide/tutorial.html)
* [實戰案例](https://f2.freshman.tech/guide/organizing-image-library.html)
* [內建變數說明](https://f2.freshman.tech/guide/how-variables-work.html)
* [配對檔案重新命名](https://f2.freshman.tech/guide/pair-renaming.html)
* [CSV 檔案批次重新命名](https://f2.freshman.tech/guide/csv-renaming.html)
* [檔案排序](https://f2.freshman.tech/guide/sorting.html)
* [衝突解決](https://f2.freshman.tech/guide/conflict-detection.html)
* [撤銷操作](https://f2.freshman.tech/guide/undoing-mistakes.html)
* [更新日誌](https://f2.freshman.tech/reference/changelog.html)

## 💻 效果截圖

![F2 利用 Exif 屬性整理圖片檔案](https://f2.freshman.tech/assets/2.D-uxLR9T.png)

## 🤝 參與貢獻

歡迎提交 Bug 報告和功能建議！提交 Pull Request 前請先創建 Issue 討論。

## ⚖ 許可協議

由 Ayooluwa Isaiah 創建，採用 [MIT 許可證](https://github.com/ayoisaiah/f2/blob/master/LICENCE)發布。

由 [Karlbaey101](//github.com/Karlbaey101) 翻譯，翻譯的文字遵循 [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/) 協議。
