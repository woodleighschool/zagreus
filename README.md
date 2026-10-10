# zagreus

[![Release](https://img.shields.io/github/v/release/woodleighschool/zagreus?display_name=tag&sort=semver)](https://github.com/woodleighschool/zagreus/releases/latest)
[![CI](https://github.com/woodleighschool/zagreus/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/woodleighschool/zagreus/actions/workflows/ci.yaml)
[![Go](https://img.shields.io/github/go-mod/go-version/woodleighschool/zagreus?logo=go)](https://github.com/woodleighschool/zagreus/blob/main/go.mod)
[![License](https://img.shields.io/github/license/woodleighschool/zagreus)](https://github.com/woodleighschool/zagreus/blob/main/LICENSE)

> **zagreus** — the head of security in Hades, the underworld

Imports and exports any changes in a scheduled Nessus scan to a designated Trello board for easy management

## 🚀 Usage

Create a `.env` file, then start the published image:

```bash
docker compose up -d
```

## ⚙️ Configuration

Configuration comes from a yaml file using the following schema:

[Zagreus Schema](zagreus.schema.json)

## 🧑‍💻 Development

Mise owns the toolchain and repository commands:

```bash
mise install
mise run deps
mise run build
mise run test
mise run lint
```

Run `mise tasks` for the available checks and generation commands

## 📄 License

Licensed under the [Apache License 2.0](LICENSE)
