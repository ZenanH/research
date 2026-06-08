# Research

`research` is a small Go CLI for exporting OpenAlex journal papers into AI-friendly Markdown.

The first MVP focuses on one reliable workflow:

- resolve a journal/source in OpenAlex
- fetch recent journal articles
- optionally filter by keywords in title and abstract
- rebuild OpenAlex inverted-index abstracts
- enrich missing abstracts from Crossref, then Semantic Scholar with a configured key or anonymous access
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

Semantic Scholar enrichment is optional. `research` can use Semantic Scholar as a second abstract fallback after Crossref. A key is recommended for better rate limits, but not required; if no key is configured, enrichment falls back to anonymous public access.

`research` looks for the Semantic Scholar key in this order:

1. `--semantic-scholar-key`
2. `SEMANTIC_SCHOLAR_API_KEY`
3. local config file
4. interactive prompt, where pressing Enter uses anonymous access

You can save a key explicitly:

```bash
research config set semantic-scholar-key
```

You can remove it:

```bash
research config unset semantic-scholar-key
```

Or set it for one shell session:

```bash
export SEMANTIC_SCHOLAR_API_KEY="..."
```

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
  --output ./research-outputs
```

By default, missing abstracts are enriched from Crossref, then Semantic Scholar. If no Semantic Scholar key is configured, the CLI uses anonymous public access with lower rate limits:

```bash
research journal \
  --name "computers and geotechnics" \
  --count 100 \
  --enrich-abstracts \
  --output ./research-outputs
```

Only export papers that have abstracts available from OpenAlex, Crossref, or Semantic Scholar:

```bash
research journal \
  --name "computers and geotechnics" \
  --count 100 \
  --require-abstract \
  --output ./research-outputs
```

Disable enrichment when you only want OpenAlex-native abstracts:

```bash
research journal \
  --name "computers and geotechnics" \
  --count 100 \
  --enrich-abstracts=false \
  --require-abstract \
  --output ./openalex_abstracts_only.md
```

Export recent papers from a journal that match any keyword:

```bash
research search \
  --journal "computers and geotechnics" \
  --count 100 \
  --keywords "machine learning,DEM,slope stability" \
  --keyword-mode any \
  --output ./research-outputs
```

`--output` can be either a Markdown file path or a directory. When it is omitted or points to a directory, `research` writes into the configured default output directory and generates a short filename from the journal abbreviation and requested count:

```text
./research-outputs/cag_100.md
./research-outputs/cag_keywords_100.md
```

If the target file already exists, `research` keeps the old file and writes to the next available name, such as `cag_100_2.md`.

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
- Papers with abstracts: ...
- Abstract coverage: ...
- Abstract sources: OpenAlex 61, Crossref 22, Semantic Scholar 0, Missing 17
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
- Abstract source: OpenAlex
- Publisher page: ...

#### Abstract

...
```

If available sources do not provide an abstract, the exporter writes:

```md
_No abstract available from available sources._
```

OpenAlex does not have abstracts for every paper. The Markdown metadata includes `Papers with abstracts`, `Abstract coverage`, and `Abstract sources` so you can quickly judge whether an export is suitable for AI-assisted reading. Use `--require-abstract` when you want the CLI to skip papers without abstracts from OpenAlex, Crossref, or Semantic Scholar and continue paging until it collects the requested count or OpenAlex has no more matching results.

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

## Release Setup

GitHub Releases use the built-in `GITHUB_TOKEN` provided by GitHub Actions. You do not need to create a repository secret named `GITHUB_TOKEN`.

Homebrew tap updates write to a separate repository:

```text
ZenanH/homebrew-research
```

Create that repository before the first tagged release. Then create a fine-grained GitHub Personal Access Token with access to `ZenanH/homebrew-research` and these permissions:

```text
Contents: Read and write
Metadata: Read
```

Add the token to the main `ZenanH/research` repository:

```text
Settings -> Secrets and variables -> Actions -> New repository secret
```

Use this secret name:

```text
HOMEBREW_TAP_GITHUB_TOKEN
```

After that, push a tag to publish:

```bash
git tag v0.2.0
git push origin v0.2.0
```
