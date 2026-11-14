**以其他語言閱讀此文檔：**[English](/README.md) | [Deutsch](/docs/README.de.md) | [Español](/docs/README.es.md) | [Português](/docs/README.pt.md) | [Русский](/docs/README.ru.md) | [繁體中文](/docs/README.zh.md) | [Tiếng Việt](/docs/README.vi.md)

<p align="center">
  <img src="https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png" width="250" height="250" alt="f2">
</p>

<p align="center">
  <a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat" alt=""></a>
  <a href="https://github.com/ayoisaiah/F2/actions"><img src="https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg" alt="Actions Github"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="Fait avec Go"></a>
  <a href="https://goreportcard.com/report/github.com/ayoisaiah/f2"><img src="https://goreportcard.com/badge/github.com/ayoisaiah/f2" alt="GoReportCard"></a>
  <a href="https://github.com/ayoisaiah/f2"><img src="https://img.shields.io/github/go-mod/go-version/ayoisaiah/f2.svg" alt="Version Go.mod"></a>
  <a href="https://github.com/ayoisaiah/f2/blob/master/LICENCE"><img src="https://img.shields.io/github/license/ayoisaiah/f2.svg" alt="LICENCE"></a>
  <a href="https://github.com/ayoisaiah/f2/releases/"><img src="https://img.shields.io/github/release/ayoisaiah/f2.svg" alt="Dernière version"></a>
</p>

<h1 align="center">F2 - Renommage par lots en ligne de commande</h1>

**F2** est un outil en ligne de commande multiplateforme pour renommer des
fichiers et des répertoires par lots **rapidement** et **en toute sécurité**.
Écrit en Go!

## Qu'est-ce que F2 fait différemment ?

Comparé à d'autres outils de renommage, F2 offre plusieurs avantages clés:

- **Simulation par défaut**: Il effectue par défaut une simulation afin que vous
  puissiez examiner les changements de nom avant de continuer.

- **Prise en charge des variables**: F2 vous permet d'utiliser les attributs des
  fichiers, tels que les données EXIF pour les images ou les balises ID3 pour
  les fichiers audio, pour vous offrir une flexibilité maximale lors du
  renommage.

- **Options complètes**: Qu'il s'agisse de simples remplacements de chaînes de
  caractères ou d'expressions régulières complexes, F2 offre une gamme complète
  de fonctionnalités de renommage.

- **La sécurité d'abord**: Il privilégie l'exactitude en s'assurant que chaque
  opération de renommage est exempte de conflits et d'erreurs grâce à des
  contrôles rigoureux.

- **Résolution des conflits**: Chaque opération de renommage est validée avant
  son exécution et les conflits détectés peuvent être résolus automatiquement.

- **Haute performance**: F2 est extrêmement rapide et efficace, même lors du
  renommage de milliers de fichiers à la fois.

- **Fonctionnalité d'annulation**: Toute opération de renommage peut être
  facilement annulée pour permettre la correction facile des erreurs.

- **Documentation complète**: F2 est bien documenté avec des exemples clairs et
  pratiques pour vous aider à tirer le meilleur parti de ses fonctionnalités
  sans confusion.

## ⚡ Installation

Si vous êtes un développeur Go, F2 peut être installé avec `go install`
(nécessite la v1.23 ou une version ultérieure):

```bash
go install github.com/ayoisaiah/f2/v2/cmd/f2@latest
```

D'autres méthodes d'installation sont
[documentées ici](https://f2.freshman.tech/guide/getting-started.html) ou
consultez la [page des versions](https://github.com/ayoisaiah/f2/releases) pour
télécharger un binaire pré-compilé pour votre système d'exploitation.

## 📃 Liens rapides

- [Installation](https://f2.freshman.tech/guide/getting-started.html)
- [Tutoriel de démarrage](https://f2.freshman.tech/guide/tutorial.html)
- [Exemple concret](https://f2.freshman.tech/guide/organizing-image-library.html)
- [Variables intégrées](https://f2.freshman.tech/guide/how-variables-work.html)
- [Renommage de paires de fichiers](https://f2.freshman.tech/guide/pair-renaming.html)
- [Renommage avec un fichier CSV](https://f2.freshman.tech/guide/csv-renaming.html)
- [Tri](https://f2.freshman.tech/guide/sorting.html)
- [Résolution des conflits](https://f2.freshman.tech/guide/conflict-detection.html)
- [Annuler les erreurs de renommage](https://f2.freshman.tech/guide/undoing-mistakes.html)
- [CHANGELOG](https://f2.freshman.tech/reference/changelog.html)

## 💻 Captures d'écran

![F2 peut utiliser les attributs Exif pour organiser les fichiers image](https://f2.freshman.tech/assets/2.D-uxLR9T.png)

## 🤝 Contribuer

Les rapports de bogues et les demandes de fonctionnalités sont les bienvenus !
Veuillez ouvrir une issue avant de créer une pull request.

## ⚖️ Licence

Créé par Ayooluwa Isaiah et publié sous les termes de la
[Licence MIT](https://github.com/ayoisaiah/f2/blob/master/LICENCE).
