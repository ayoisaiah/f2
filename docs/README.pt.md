<p align="center">
  <img src="https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png" width="250" height="250" alt="f2">
</p>

<p align="center">
  <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-bem--vindos-brightgreen.svg?style=flat" alt=""></a>
  <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Ações do Github"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Feito%20com-Go-1f425f.svg" alt="feito-com-Go"></a>
  <a href="https://goreportcard.com/report/github.com/ayoisaiah/f2"><img src="https://goreportcard.com/badge/github.com/ayoisaiah/f2" alt="GoReportCard"></a>
  <a href="https://github.com/ayoisaiah/f2"><img src="https://img.shields.io/github/go-mod/go-version/ayoisaiah/f2.svg" alt="Versão do Go.mod"></a>
  <a href="https://github.com/ayoisaiah/f2/blob/master/LICENCE"><img src="https://img.shields.io/github/license/ayoisaiah/f2.svg" alt="LICENÇA"></a>
  <a href="https://github.com/ayoisaiah/f2/releases/"><img src="https://img.shields.io/github/release/ayoisaiah/f2.svg" alt="Última versão"></a>
</p>

<h1 align="center">F2 - Renomeação em Lote na Linha de Comando</h1>

**F2** é uma ferramenta de linha de comando multiplataforma para renomear
arquivos e diretórios em lote de forma **rápida** e **segura**. Escrito em Go!

## O que o F2 faz de diferente?

Em comparação com outras ferramentas de renomeação, o F2 oferece várias
vantagens importantes:

- **Execução de Teste por Padrão**: Por padrão, ele executa um teste para que
  você possa revisar as alterações de renomeação antes de prosseguir.

- **Suporte a Variáveis**: O F2 permite que você use atributos de arquivo, como
  dados EXIF para imagens ou tags ID3 para arquivos de áudio, para lhe dar a
  máxima flexibilidade na renomeação.

- **Opções Abrangentes**: Seja para substituições simples de strings ou
  expressões regulares complexas, o F2 oferece uma gama completa de recursos de
  renomeação.

- **Segurança em Primeiro Lugar**: Ele prioriza a precisão, garantindo que cada
  operação de renomeação seja livre de conflitos e à prova de erros por meio de
  verificações rigorosas.

- **Resolução de Conflitos**: Cada operação de renomeação é validada antes da
  execução e os conflitos detectados podem ser resolvidos automaticamente.

- **Alto Desempenho**: O F2 é extremamente rápido e eficiente, mesmo ao renomear
  milhares de arquivos de uma só vez.

- **Funcionalidade de Desfazer**: Qualquer operação de renomeação pode ser
  facilmente desfeita para permitir a correção fácil de erros.

- **Documentação Extensa**: O F2 é bem documentado com exemplos claros e
  práticos para ajudá-lo a aproveitar ao máximo seus recursos sem confusão.

## ⚡ Instalação

Se você é um desenvolvedor Go, o F2 pode ser instalado com `go install` (requer
v1.23 ou posterior):

```bash
go install github.com/ayoisaiah/f2/v2/cmd/f2@latest
```

Outros métodos de instalação estão
[documentados aqui](https://f2.freshman.tech/guide/getting-started.html) ou
confira a [página de lançamentos](https://github.com/ayoisaiah/f2/releases) para
baixar um binário pré-compilado para o seu sistema operacional.

## 📃 Links rápidos

- [Instalação](https://f2.freshman.tech/guide/getting-started.html)
- [Tutorial de introdução](https://f2.freshman.tech/guide/tutorial.html)
- [Exemplo do mundo real](https://f2.freshman.tech/guide/organizing-image-library.html)
- [Variáveis incorporadas](https://f2.freshman.tech/guide/how-variables-work.html)
- [Renomeação de pares de arquivos](https://f2.freshman.tech/guide/pair-renaming.html)
- [Renomeando com um arquivo CSV](https://f2.freshman.tech/guide/csv-renaming.html)
- [Classificação](https://f2.freshman.tech/guide/sorting.html)
- [Resolvendo conflitos](https://f2.freshman.tech/guide/conflict-detection.html)
- [Desfazendo erros de renomeação](https://f2.freshman.tech/guide/undoing-mistakes.html)
- [REGISTRO DE ALTERAÇÕES](https://f2.freshman.tech/reference/changelog.html)

## 💻 Capturas de tela

![O F2 pode utilizar atributos Exif para organizar arquivos de imagem](https://f2.freshman.tech/assets/2.D-uxLR9T.png)

## 🤝 Contribuir

Relatórios de bugs e solicitações de recursos são muito bem-vindos! Por favor,
abra uma issue antes de criar um pull request.

## ⚖️ Licença

Criado por Ayooluwa Isaiah e lançado sob os termos da
[Licença MIT](https://github.com/ayoisaiah/f2/blob/master/LICENCE).
