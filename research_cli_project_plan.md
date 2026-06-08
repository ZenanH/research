# Research CLI 项目方案

## 1. 项目定位

`research` 是一个面向文献调研的独立命令行工具。用户在 Terminal 中输入：

```bash
research
```

即可进入一个简洁、漂亮的 TUI 交互界面，选择期刊、文章数量、关键词和导出路径，然后从 OpenAlex 获取论文元数据，并导出适合后续 AI 调研的 Markdown 文件。

项目目标不是做一个一次性脚本，而是做成可长期维护、可发布、可通过 Homebrew 安装的 CLI app。

## 2. 核心约束

- 不改动用户本地 Python 环境。
- 不要求用户安装或管理 Python 包、虚拟环境、Conda 环境。
- 用户通过 `research` 命令进入 TUI。
- 在 TUI 中强制要求用户提供 OpenAlex API key。
- OpenAlex 作为第一版唯一数据库。
- 第一版重点是稳定导出期刊论文题目、作者、日期、摘要。
- 默认导出一个合并 Markdown，方便后续交给 AI 做调研。
- 后续可扩展为多数据源、多格式、多任务的 research CLI。

## 3. OpenAlex API Key 策略

OpenAlex 旧文档中曾说明 API 免费且无需认证；但当前官方文档已经更新为：API 免费，但需要免费的 API key。OpenAlex 文档还注明，从 2026-02-13 起，使用 API 需要 API key。

因此本项目采用更明确、更稳定的产品策略：

- TUI 启动后检查是否已有 OpenAlex API key。
- 如果没有，必须要求用户输入 key。
- 不提供“跳过 key 并继续”的默认路径。
- key 可以临时使用，也可以由用户选择保存到本机配置。
- 保存 key 必须是用户显式选择，不能悄悄写入配置文件。
- CLI 参数模式也必须提供 key，或从配置 / 环境变量读取 key。

推荐优先级：

1. 命令行参数：`--openalex-key`
2. 环境变量：`OPENALEX_API_KEY`
3. 本机配置文件
4. TUI 中用户输入

官方参考：

- [OpenAlex API Overview](https://docs.openalex.org/how-to-use-the-api/api-overview)
- [OpenAlex Rate limits and authentication](https://docs.openalex.org/how-to-use-the-api/rate-limits-and-authentication)
- [OpenAlex Authentication & Pricing](https://developers.openalex.org/api-reference/authentication)

## 4. 推荐技术路线

推荐使用 Go 实现。

原因：

- 编译后是单个可执行文件。
- 不依赖 Python 环境。
- 不依赖 Node.js / npm。
- 适合通过 Homebrew 发布。
- 适合跨平台发布 macOS / Linux / Windows。
- Go 的 CLI 和 TUI 生态成熟。
- 对本项目的网络请求、表单输入、Markdown 输出来说复杂度适中。

建议技术栈：

- 语言：Go
- CLI 框架：`cobra`
- TUI / 表单：`charmbracelet/huh`
- 样式：`charmbracelet/lipgloss`
- Spinner / progress：`charmbracelet/bubbles`
- HTTP：Go 标准库 `net/http`
- 配置：XDG config path / macOS Application Support
- 发布：GoReleaser
- 安装：Homebrew tap

## 5. 命令设计

### 5.1 交互模式

默认入口：

```bash
research
```

启动后进入 TUI：

```text
Research
OpenAlex journal paper exporter

? Choose a workflow
  Recent papers from journal
  Keyword search in journal
  Settings
  Exit
```

### 5.2 非交互模式

为了方便自动化和批处理，也保留命令参数模式。

功能 1：按期刊导出最近文章：

```bash
research journal \
  --name "computers and geotechnics" \
  --count 100 \
  --output ./computers_and_geotechnics_recent_100.md
```

功能 2：按期刊和关键词导出文章：

```bash
research search \
  --journal "computers and geotechnics" \
  --count 100 \
  --keywords "machine learning,DEM,slope stability" \
  --keyword-mode any \
  --output ./computers_and_geotechnics_keywords_100.md
```

查看 / 设置配置：

```bash
research config
research config set openalex-key
research config unset openalex-key
```

列出期刊候选项：

```bash
research sources "computers and geotechnics"
```

## 6. TUI 功能设计

### 6.1 首次启动

如果没有检测到 OpenAlex API key：

```text
OpenAlex API key required

OpenAlex is free, but current API access requires a free API key.
Get one at: https://openalex.org/settings/api

? Enter OpenAlex API key:
```

用户输入后：

```text
? Save this key for future runs?
  Yes, save to local config
  No, use only this session
```

如果用户选择保存，写入本机配置文件。

### 6.2 功能 1：按期刊最近文章导出

用户输入：

- 期刊名称
- 文章数量
- 导出路径

TUI 示例：

```text
Recent papers from journal

Journal name
Computers and Geotechnics

Number of papers
100

Output path
./outputs/computers_and_geotechnics_recent_100.md
```

执行流程：

1. 校验 API key。
2. 用期刊名称搜索 OpenAlex source。
3. 如果有多个候选 source，要求用户选择。
4. 使用选定 source ID 查询 works。
5. 按 `publication_date:desc` 排序。
6. 获取前 X 篇。
7. 生成 Markdown。

### 6.3 功能 2：按期刊 + 多关键词导出

用户输入：

- 期刊名称
- 文章数量
- 多个关键词
- 关键词匹配模式
- 导出路径

TUI 示例：

```text
Keyword search in journal

Journal name
Computers and Geotechnics

Number of papers
100

Keywords
machine learning, DEM, slope stability

Keyword mode
any

Output path
./outputs/computers_and_geotechnics_keywords_100.md
```

关键词模式：

- `any`：标题或摘要命中任意关键词即可。
- `all`：标题或摘要必须命中全部关键词。

推荐语义：

> 在指定期刊中，找到最近的 X 篇满足关键词条件的论文。

这比“先取最近 X 篇再过滤关键词”更符合用户预期。否则用户输入 100 篇，最终可能只得到 8 篇。

实现方式：

1. 优先使用 OpenAlex 查询能力缩小候选范围。
2. 拉取候选论文。
3. 在本地对 `title + abstract` 做精确关键词过滤。
4. 直到收集到 X 篇，或 OpenAlex 没有更多结果。

## 7. OpenAlex 查询设计

### 7.1 期刊匹配

输入期刊名后，先查询 OpenAlex sources。

期刊候选信息应显示：

- display name
- ISSN-L
- ISSN
- works count
- OpenAlex source ID

如果只有一个高置信匹配，可以直接使用；如果多个候选项接近，应让用户选择。

后续也可以支持用户直接输入 ISSN：

```bash
research journal --issn 0266-352X --count 100
```

### 7.2 最近文章查询

推荐查询逻辑：

- endpoint：OpenAlex works
- filter：`primary_location.source.id:<source_id>`
- filter：`type:article`
- sort：`publication_date:desc`
- pagination：cursor paging
- per page：遵循 OpenAlex 当前文档限制

默认只取 `article`，后续可以支持：

```bash
--types article,review
```

### 7.3 摘要处理

OpenAlex 的摘要通常以 `abstract_inverted_index` 形式返回，需要本地重建为正常文本。

如果没有摘要，Markdown 中明确写：

```md
_No abstract available from OpenAlex._
```

不要编造摘要，不要自动用标题生成摘要。

## 8. Markdown 输出设计

默认输出一个合并 Markdown 文件，因为这最适合后续交给 AI 调研。

建议结构：

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
- Generated at: ...

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

关键词导出时增加：

```md
## Query

- Keywords: machine learning, DEM, slope stability
- Keyword mode: any
```

### 8.1 输出模式

第一版 TUI 中默认只暴露合并输出。

后续 CLI 参数可以支持：

- `combined`：一个完整 Markdown，默认。
- `split`：`index.md` + 每篇论文一个 Markdown。
- `both`：同时生成合并文件和拆分文件。

建议第一版先实现 `combined`，保留扩展点即可。

## 9. 配置策略

不能修改用户 Python 环境，也不依赖 Python 配置。

配置文件建议位置：

macOS：

```text
~/Library/Application Support/research/config.toml
```

Linux：

```text
~/.config/research/config.toml
```

Windows：

```text
%AppData%\research\config.toml
```

配置内容示例：

```toml
[openalex]
api_key = ""

[output]
default_dir = "./research-outputs"

[export]
mode = "combined"
```

安全原则：

- 只有用户显式选择保存时才写入 API key。
- 不在日志里打印完整 API key。
- 错误信息里不回显 API key。
- 可以提供 `research config unset openalex-key` 删除 key。

## 10. 项目结构建议

```text
research/
  cmd/
    research/
      main.go
  internal/
    app/
      app.go
    config/
      config.go
    openalex/
      client.go
      source.go
      works.go
      abstract.go
    tui/
      home.go
      key.go
      journal.go
      search.go
    exporter/
      markdown.go
      filename.go
    model/
      paper.go
      source.go
  testdata/
    openalex_source_response.json
    openalex_works_response.json
  README.md
  LICENSE
  go.mod
  go.sum
  .goreleaser.yaml
  Formula/
    research.rb
```

## 11. Homebrew 发布路线

目标安装方式：

```bash
brew tap <owner>/research
brew install research
research
```

推荐使用 GoReleaser 自动化：

1. GitHub repo：`research`
2. 配置 `.goreleaser.yaml`
3. GitHub Actions 在 tag 时构建 release
4. 自动产出 macOS / Linux 二进制
5. 自动更新 Homebrew tap formula

发布流程：

```bash
git tag v0.1.0
git push origin v0.1.0
```

Homebrew formula 核心逻辑：

```ruby
class Research < Formula
  desc "Fetch OpenAlex journal papers and export AI-friendly Markdown"
  homepage "https://github.com/<owner>/research"
  url "https://github.com/<owner>/research/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"research", "./cmd/research"
  end

  test do
    system "#{bin}/research", "--version"
  end
end
```

## 12. MVP 范围

第一版建议只做这些：

- Go 单二进制 CLI。
- 命令名：`research`。
- TUI 首页。
- 首次启动强制输入 OpenAlex API key。
- 可选保存 key 到本机配置。
- 功能 1：期刊名 + 数量 + 导出路径。
- 功能 2：期刊名 + 数量 + 多关键词 + 导出路径。
- 多 source 候选选择。
- 默认导出合并 Markdown。
- 摘要缺失时明确标注。
- README。
- 基础单元测试。
- GoReleaser / Homebrew 发布预留配置。

暂不做：

- GUI。
- PDF 下载。
- 全文解析。
- 多数据库来源。
- AI 自动总结。
- 文献引用网络分析。
- Web dashboard。

## 13. 后续扩展

后续可以逐步增加：

- Semantic Scholar / Crossref fallback。
- DOI 去重。
- 按日期范围筛选。
- 按 review / article / proceeding 等类型筛选。
- `split` / `both` 输出模式。
- JSON / CSV / BibTeX 输出。
- 直接生成 AI 调研 prompt。
- 直接生成主题聚类报告。
- 下载开放获取 PDF。
- 接入本地 LLM 或 OpenAI API 做摘要聚合。

## 14. 推荐最终产品体验

用户第一次运行：

```bash
research
```

看到：

```text
Research
OpenAlex journal paper exporter

OpenAlex API key required.
Get a free key at https://openalex.org/settings/api

? Enter OpenAlex API key:
? Save this key for future runs?
```

然后：

```text
? Choose a workflow
  Recent papers from journal
  Keyword search in journal
  Settings
  Exit
```

选择功能 1：

```text
Journal name: Computers and Geotechnics
Number of papers: 100
Output path: ./computers_and_geotechnics_recent_100.md

Resolving journal...
Matched: Computers and Geotechnics
Fetching papers... 100/100
Writing Markdown...
Done.
```

选择功能 2：

```text
Journal name: Computers and Geotechnics
Number of papers: 100
Keywords: machine learning, DEM, slope stability
Keyword mode: any
Output path: ./computers_and_geotechnics_keywords_100.md

Resolving journal...
Matched: Computers and Geotechnics
Searching papers...
Filtering by keywords...
Writing Markdown...
Done.
```

## 15. 一句话总结

把它做成一个 Go 编写的独立 CLI：`research`。用户在 Terminal 输入 `research` 进入漂亮 TUI，强制提供 OpenAlex API key，然后选择“按期刊最近文章导出”或“按期刊 + 多关键词导出”，最终生成适合 AI 文献调研的 Markdown 文件。项目不依赖 Python，不修改本地 Python 环境，并预留 Homebrew 发布路径。
