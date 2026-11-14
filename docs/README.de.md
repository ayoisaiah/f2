**Lesen Sie dies in anderen Sprachen:** [English](/README.md) | [Español](/docs/README.es.md) | [Français](/docs/README.fr.md) | [Português](/docs/README.pt.md) | [Русский](/docs/README.ru.md) | [繁體中文](/docs/README.zh.md) | [Tiếng Việt](/docs/README.vi.md)

<p align="center">
  <img src="https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png" width="250" height="250" alt="f2">
</p>

<p align="center">
  <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-willkommen-brightgreen.svg?style=flat" alt=""></a>
  <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Github-Aktionen"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Erstellt%20mit-Go-1f425f.svg" alt="erstellt-mit-Go"></a>
  <a href="https://goreportcard.com/report/github.com/ayoisaiah/f2"><img src="https://goreportcard.com/badge/github.com/ayoisaiah/f2" alt="GoReportCard"></a>
  <a href="https://github.com/ayoisaiah/f2"><img src="https://img.shields.io/github/go-mod/go-version/ayoisaiah/f2.svg" alt="Go.mod-Version"></a>
  <a href="https://github.com/ayoisaiah/f2/blob/master/LICENCE"><img src="https://img.shields.io/github/license/ayoisaiah/f2.svg" alt="LIZENZ"></a>
  <a href="https://github.com/ayoisaiah/f2/releases/"><img src="https://img.shields.io/github/release/ayoisaiah/f2.svg" alt="Neueste Version"></a>
</p>

<h1 align="center">F2 – Stapelumbenennung über die Befehlszeile</h1>

**F2** ist ein plattformübergreifendes Befehlszeilentool zum **schnellen** und
**sicheren** Stapelumbenennen von Dateien und Verzeichnissen. Geschrieben in Go!

## Was macht F2 anders?

Im Vergleich zu anderen Umbenennungstools bietet F2 mehrere wichtige Vorteile:

- **Standardmäßiger Probelauf**: Standardmäßig wird ein Probelauf durchgeführt,
  damit Sie die Umbenennungsänderungen vor dem Fortfahren überprüfen können.

- **Variablenunterstützung**: F2 ermöglicht die Verwendung von Dateiattributen
  wie EXIF-Daten für Bilder oder ID3-Tags für Audiodateien, um Ihnen maximale
  Flexibilität bei der Umbenennung zu bieten.

- **Umfassende Optionen**: Ob einfache Zeichenfolgenersetzungen oder komplexe
  reguläre Ausdrücke, F2 bietet eine vollständige Palette von
  Umbenennungsfunktionen.

- **Sicherheit geht vor**: Es legt Wert auf Genauigkeit, indem es durch strenge
  Prüfungen sicherstellt, dass jeder Umbenennungsvorgang konfliktfrei und
  fehlerfrei ist.

- **Konfliktlösung**: Jeder Umbenennungsvorgang wird vor der Ausführung
  validiert und erkannte Konflikte können automatisch gelöst werden.

- **Hohe Leistung**: F2 ist extrem schnell und effizient, selbst beim Umbenennen
  von Tausenden von Dateien auf einmal.

- **Rückgängig-Funktionalität**: Jeder Umbenennungsvorgang kann einfach
  rückgängig gemacht werden, um Fehler einfach zu korrigieren.

- **Umfangreiche Dokumentation**: F2 ist gut dokumentiert mit klaren,
  praktischen Beispielen, die Ihnen helfen, die Funktionen ohne Verwirrung
  optimal zu nutzen.

## ⚡ Installation

Wenn Sie ein Go-Entwickler sind, kann F2 mit `go install` installiert werden
(erfordert v1.23 oder höher):

```bash
go install github.com/ayoisaiah/f2/v2/cmd/f2@latest
```

Andere Installationsmethoden sind
[hier dokumentiert](https://f2.freshman.tech/guide/getting-started.html) oder
sehen Sie sich die
[Seite mit den Versionen](https://github.com/ayoisaiah/f2/releases) an, um eine
vorkompilierte Binärdatei für Ihr Betriebssystem herunterzuladen.

## 📃 Nützliche Links

- [Installation](https://f2.freshman.tech/guide/getting-started.html)
- [Tutorial für den Einstieg](https://f2.freshman.tech/guide/tutorial.html)
- [Praxisbeispiel](https://f2.freshman.tech/guide/organizing-image-library.html)
- [Integrierte Variablen](https://f2.freshman.tech/guide/how-variables-work.html)
- [Umbenennen von Dateipaaren](https://f2.freshman.tech/guide/pair-renaming.html)
- [Umbenennen mit einer CSV-Datei](https://f2.freshman.tech/guide/csv-renaming.html)
- [Sortierung](https://f2.freshman.tech/guide/sorting.html)
- [Konflikte lösen](https://f2.freshman.tech/guide/conflict-detection.html)
- [Umbenennungsfehler rückgängig machen](https://f2.freshman.tech/guide/undoing-mistakes.html)
- [ÄNDERUNGSPROTOKOLL](https://f2.freshman.tech/reference/changelog.html)

## 💻 Screenshots

![F2 kann Exif-Attribute verwenden, um Bilddateien zu organisieren](https://f2.freshman.tech/assets/2.D-uxLR9T.png)

## 🤝 Mitwirken

Fehlerberichte und Funktionswünsche sind herzlich willkommen! Bitte öffnen Sie
ein issue, bevor Sie eine pull request erstellen.

## ⚖️ Lizenz

Erstellt von Ayooluwa Isaiah und veröffentlicht unter den Bedingungen der
[MIT-Lizenz](https://github.com/ayoisaiah/f2/blob/master/LICENCE).
