**Прочитайте это на других языках:** [English](/README.md) | [Deutsch](/docs/README.de.md) | [Español](/docs/README.es.md) | [Français](/docs/README.fr.md) | [Português](/docs/README.pt.md) | [繁體中文](/docs/README.zh.md) | [Tiếng Việt](/docs/README.vi.md)

<p align="center">
  <img src="https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png" width="250" height="250" alt="f2">
</p>

<p align="center">
  <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-приветствуются-brightgreen.svg?style=flat" alt=""></a>
  <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Действия Github"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Сделано%20на-Go-1f425f.svg" alt="сделано-на-Go"></a>
  <a href="https://goreportcard.com/report/github.com/ayoisaiah/f2"><img src="https://goreportcard.com/badge/github.com/ayoisaiah/f2" alt="GoReportCard"></a>
  <a href="https://github.com/ayoisaiah/f2"><img src="https://img.shields.io/github/go-mod/go-version/ayoisaiah/f2.svg" alt="Версия Go.mod"></a>
  <a href="https://github.com/ayoisaiah/f2/blob/master/LICENCE"><img src="https://img.shields.io/github/license/ayoisaiah/f2.svg" alt="ЛИЦЕНЗИЯ"></a>
  <a href="https://github.com/ayoisaiah/f2/releases/"><img src="https://img.shields.io/github/release/ayoisaiah/f2.svg" alt="Последняя версия"></a>
</p>

<h1 align="center">F2 - Пакетное переименование в командной строке</h1>

**F2** — это кроссплатформенный инструмент командной строки для пакетного
переименования файлов и каталогов **быстро** и **безопасно**. Написан на Go!

## Что F2 делает по-другому?

По сравнению с другими инструментами переименования, F2 предлагает несколько
ключевых преимуществ:

- **Тестовый запуск по умолчанию**: по умолчанию выполняется тестовый запуск,
  чтобы вы могли просмотреть изменения в переименовании перед продолжением.

- **Поддержка переменных**: F2 позволяет использовать атрибуты файлов, такие как
  данные EXIF для изображений или теги ID3 для аудиофайлов, чтобы обеспечить
  максимальную гибкость при переименовании.

- **Комплексные параметры**: будь то простая замена строк или сложные регулярные
  выражения, F2 предоставляет полный спектр возможностей переименования.

- **Безопасность прежде всего**: он отдает приоритет точности, гарантируя, что
  каждая операция переименования не содержит конфликтов и ошибок благодаря
  строгим проверкам.

- **Разрешение конфликтов**: каждая операция переименования проверяется перед
  выполнением, и обнаруженные конфликты могут быть разрешены автоматически.

- **Высокая производительность**: F2 чрезвычайно быстр и эффективен даже при
  переименовании тысяч файлов одновременно.

- **Функциональность отмены**: любую операцию переименования можно легко
  отменить, чтобы легко исправить ошибки.

- **Обширная документация**: F2 хорошо документирован с четкими, практическими
  примерами, которые помогут вам максимально эффективно использовать его функции
  без путаницы.

## ⚡ Установка

Если вы разработчик Go, F2 можно установить с помощью `go install` (требуется
версия 1.23 или новее):

```bash
go install github.com/ayoisaiah/f2/v2/cmd/f2@latest
```

Другие способы установки
[задокументированы здесь](https://f2.freshman.tech/guide/getting-started.html)
или ознакомьтесь со
[страницей выпусков](https://github.com/ayoisaiah/f2/releases), чтобы загрузить
предварительно скомпилированный двоичный файл для вашей операционной системы.

## 📃 Быстрые ссылки

- [Установка](https://f2.freshman.tech/guide/getting-started.html)
- [Учебное пособие по началу работы](https://f2.freshman.tech/guide/tutorial.html)
- [Пример из реальной жизни](https://f2.freshman.tech/guide/organizing-image-library.html)
- [Встроенные переменные](https://f2.freshman.tech/guide/how-variables-work.html)
- [Переименование пары файлов](https://f2.freshman.tech/guide/pair-renaming.html)
- [Переименование с помощью CSV-файла](https://f2.freshman.tech/guide/csv-renaming.html)
- [Сортировка](https://f2.freshman.tech/guide/sorting.html)
- [Разрешение конфликтов](https://f2.freshman.tech/guide/conflict-detection.html)
- [Отмена ошибок переименования](https://f2.freshman.tech/guide/undoing-mistakes.html)
- [СПИСОК ИЗМЕНЕНИЙ](https://f2.freshman.tech/reference/changelog.html)

## 💻 Скриншоты

![F2 может использовать атрибуты Exif для организации файлов изображений](https://f2.freshman.tech/assets/2.D-uxLR9T.png)

## 🤝 Внести свой вклад

Сообщения об ошибках и пожелания приветствуются! Пожалуйста, откройте issue,
прежде чем создавать pull request.

## ⚖️ Лицензия

Создано Ayooluwa Isaiah и выпущено на условиях
[лицензии MIT](https://github.com/ayoisaiah/f2/blob/master/LICENCE).
