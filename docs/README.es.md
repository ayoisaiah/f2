**Lea esto en otros idiomas:** [English](/README.md) | [Deutsch](/docs/README.de.md) | [Français](/docs/README.fr.md) | [Português](/docs/README.pt.md) | [Русский](/docs/README.ru.md) | [繁體中文](/docs/README.zh.md) | [Tiếng Việt](/docs/README.vi.md)

<p align="center">
  <img src="https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png" width="250" height="250" alt="f2">
</p>

<p align="center">
  <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-bienvenidas-brightgreen.svg?style=flat" alt=""></a>
  <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Acciones de Github"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Hecho%20con-Go-1f425f.svg" alt="hecho-con-Go"></a>
  <a href="https://goreportcard.com/report/github.com/ayoisaiah/f2"><img src="https://goreportcard.com/badge/github.com/ayoisaiah/f2" alt="GoReportCard"></a>
  <a href="https://github.com/ayoisaiah/f2"><img src="https://img.shields.io/github/go-mod/go-version/ayoisaiah/f2.svg" alt="Versión de Go.mod"></a>
  <a href="https://github.com/ayoisaiah/f2/blob/master/LICENCE"><img src="https://img.shields.io/github/license/ayoisaiah/f2.svg" alt="LICENCIA"></a>
  <a href="https://github.com/ayoisaiah/f2/releases/"><img src="https://img.shields.io/github/release/ayoisaiah/f2.svg" alt="Última versión"></a>
</p>

<h1 align="center">F2 - Renombrado por lotes en línea de comandos</h1>

**F2** es una herramienta de línea de comandos multiplataforma para renombrar
archivos y directorios por lotes de forma **rápida** y **segura**. ¡Escrito en
Go!

## ¿Qué hace F2 de manera diferente?

En comparación con otras herramientas de renombrado, F2 ofrece varias ventajas
clave:

- **Simulacro por defecto**: Por defecto, realiza una simulación para que pueda
  revisar los cambios de nombre antes de continuar.

- **Soporte de variables**: F2 le permite utilizar atributos de archivo, como
  datos EXIF para imágenes o etiquetas ID3 para archivos de audio, para
  brindarle la máxima flexibilidad en el renombrado.

- **Opciones completas**: Ya sea que se trate de reemplazos de cadenas simples o
  expresiones regulares complejas, F2 ofrece una gama completa de capacidades de
  renombrado.

- **La seguridad es lo primero**: Prioriza la precisión al garantizar que cada
  operación de renombrado esté libre de conflictos y errores mediante
  comprobaciones rigurosas.

- **Resolución de conflictos**: Cada operación de renombrado se valida antes de
  la ejecución y los conflictos detectados se pueden resolver automáticamente.

- **Alto rendimiento**: F2 es extremadamente rápido y eficiente, incluso al
  renombrar miles de archivos a la vez.

- **Funcionalidad de deshacer**: Cualquier operación de renombrado se puede
  deshacer fácilmente para permitir la corrección sencilla de errores.

- **Documentación extensa**: F2 está bien documentado con ejemplos claros y
  prácticos para ayudarlo a aprovechar al máximo sus funciones sin confusión.

## ⚡ Instalación

Si eres un desarrollador de Go, F2 se puede instalar con `go install` (requiere
v1.23 o posterior):

```bash
go install github.com/ayoisaiah/f2/v2/cmd/f2@latest
```

Otros métodos de instalación están
[documentados aquí](https://f2.freshman.tech/guide/getting-started.html) o
consulte la [página de versiones](https://github.com/ayoisaiah/f2/releases) para
descargar un binario precompilado para su sistema operativo.

## 📃 Enlaces rápidos

- [Instalación](https://f2.freshman.tech/guide/getting-started.html)
- [Tutorial de inicio](https://f2.freshman.tech/guide/tutorial.html)
- [Ejemplo del mundo real](https://f2.freshman.tech/guide/organizing-image-library.html)
- [Variables integradas](https://f2.freshman.tech/guide/how-variables-work.html)
- [Renombrado de pares de archivos](https://f2.freshman.tech/guide/pair-renaming.html)
- [Renombrado con un archivo CSV](https://f2.freshman.tech/guide/csv-renaming.html)
- [Clasificación](https://f2.freshman.tech/guide/sorting.html)
- [Resolución de conflictos](https://f2.freshman.tech/guide/conflict-detection.html)
- [Deshacer errores de renombrado](https://f2.freshman.tech/guide/undoing-mistakes.html)
- [REGISTRO DE CAMBIOS](https://f2.freshman.tech/reference/changelog.html)

## 💻 Capturas de pantalla

![F2 puede utilizar atributos Exif para organizar archivos de imagen](https://f2.freshman.tech/assets/2.D-uxLR9T.png)

## 🤝 Contribuir

¡Los informes de errores y las solicitudes de funciones son muy bienvenidos!
Abra un issue antes de crear una pull request.

## ⚖️ Licencia

Creado por Ayooluwa Isaiah y publicado bajo los términos de la
[Licencia MIT](https://github.com/ayoisaiah/f2/blob/master/LICENCE).
