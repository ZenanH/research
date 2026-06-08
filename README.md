# Research

`research` is a small Go CLI for exporting OpenAlex journal papers into AI-friendly Markdown.

The first MVP focuses on one reliable workflow:

- resolve a journal/source in OpenAlex
- fetch recent journal articles
- optionally filter by keywords in title and abstract
- rebuild OpenAlex inverted-index abstracts
- export one combined Markdown file for downstream AI literature review

## Install From Source

```bash
go build -o research ./cmd/research
```

Then move the binary somewhere on your `PATH`.

## OpenAlex API Key

OpenAlex API access requires a free API key. `research` looks for the key in this order:

1. `--openalex-key`
2. `OPENALEX_API_KEY`
3. local config file
4. interactive prompt

You can save a key explicitly:

```bash
research config set openalex-key
```

You can remove it:

```bash
research config unset openalex-key
```

Config locations:

- macOS: `~/Library/Application Support/research/config.toml`
- Linux: `~/.config/research/config.toml`
- Windows: `%AppData%\research\config.toml`

## Interactive Mode

```bash
research
```

The interactive entry asks for an OpenAlex API key if none is available, then shows:

```text
Choose a workflow
  1. Recent papers from journal
  2. Keyword search in journal
  3. Settings
  4. Exit
```

When multiple OpenAlex source candidates are found, interactive mode asks you to choose one.

## Non-Interactive Commands

Export recent papers from a journal:

```bash
research journal \
  --name "computers and geotechnics" \
  --count 100 \
  --output ./computers_and_geotechnics_recent_100.md
```

Export recent papers from a journal that match any keyword:

```bash
research search \
  --journal "computers and geotechnics" \
  --count 100 \
  --keywords "machine learning,DEM,slope stability" \
  --keyword-mode any \
  --output ./computers_and_geotechnics_keywords_100.md
```

List source candidates:

```bash
research sources "computers and geotechnics"
```

Show config:

```bash
research config
```

## Markdown Output

The default output is one combined Markdown file:

```md
# Computers and Geotechnics: Recent 100 Papers

## Metadata

- Journal: Computers and Geotechnics
- Source: OpenAlex
- Source ID: ...
- ISSN-L: ...
- Requested count: 100
- Retrieved count: 100
- Query type: recent
- Sort: publication_date:desc

## Index

| # | Date | Title | Authors |
|---:|---|---|---|

## Papers

### 1. Paper title

- Date: 2026-...
- Authors: ...
- DOI: ...
- OpenAlex: ...
- Publisher page: ...

#### Abstract

...
```

If OpenAlex does not provide an abstract, the exporter writes:

```md
_No abstract available from OpenAlex._
```

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build -o research ./cmd/research
```

Format:

```bash
gofmt -w .
```

This MVP intentionally uses only the Go standard library. The project plan recommends Cobra and Charm TUI packages for a richer future UI; the current internal package boundaries are designed so that those can be introduced later without rewriting the OpenAlex or export layers.

