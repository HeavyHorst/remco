package main

import (
	"strings"
	"testing"
)

func TestAddLineNumbersToCodeBlocksDoesNotInsertInterLineWhitespace(t *testing.T) {
	html := `<pre><code>first
second
</code></pre>`

	got := addLineNumbersToCodeBlocks(html)

	if strings.Contains(got, "</span>\n<span") {
		t.Fatalf("expected adjacent line spans without preserved newlines, got %q", got)
	}
	if !strings.Contains(got, `<span class="line">first</span><span class="line">second</span>`) {
		t.Fatalf("expected code lines to stay wrapped in order, got %q", got)
	}
}

func TestAddLineNumbersToCodeBlocksClassifiesCommandsAndPreservesAttrs(t *testing.T) {
	html := `<pre><code class="language-bash">remco -c config
</code></pre>`

	got := addLineNumbersToCodeBlocks(html)

	if !strings.Contains(got, `class="code-block code-block-command"`) {
		t.Fatalf("expected command classification on pre block, got %q", got)
	}
	if !strings.Contains(got, `<code class="language-bash">`) {
		t.Fatalf("expected original code attrs to be preserved, got %q", got)
	}
}

func TestAddLineNumbersToCodeBlocksPreservesEscapedHTML(t *testing.T) {
	html := `<pre><code>&lt;div&gt;safe &amp; sound&lt;/div&gt;
</code></pre>`

	got := addLineNumbersToCodeBlocks(html)

	if strings.Contains(got, `<span class="line"><div>`) {
		t.Fatalf("expected escaped HTML to stay escaped, got %q", got)
	}
	if !strings.Contains(got, `<span class="line">&lt;div&gt;safe &amp; sound&lt;/div&gt;</span>`) {
		t.Fatalf("expected escaped code to be preserved, got %q", got)
	}
}

func TestAddLineNumbersToCodeBlocksHonorsManualOverride(t *testing.T) {
	html := `<!-- remco-manual:block=diagram -->
<pre><code>┌── remco ──┐
</code></pre>`

	got := addLineNumbersToCodeBlocks(html)

	if !strings.Contains(got, `class="code-block code-block-diagram"`) {
		t.Fatalf("expected manual diagram classification, got %q", got)
	}
	if strings.Contains(got, `remco-manual:block=diagram`) {
		t.Fatalf("expected override comment to be consumed, got %q", got)
	}
}

func TestAddHeadingAnchorsAndNumbersReplacesWholeHeading(t *testing.T) {
	html := `<h2 id="build">Build</h2><h3 id="with-code">Use <code>remco</code></h3>`
	entries := []tocEntry{
		{level: 2, text: "Build", id: "build"},
		{level: 3, text: "Use remco", id: "with-code"},
	}

	got := addHeadingAnchorsAndNumbers(html, numberTOCEntries(13, entries))

	if strings.Contains(got, `</h2></h2>`) || strings.Contains(got, `</h3></h3>`) {
		t.Fatalf("expected headings to have a single closing tag, got %q", got)
	}
	if !strings.Contains(got, `<h2 id="build"><a class="heading-anchor" href="#build">13.1 Build</a></h2>`) {
		t.Fatalf("expected numbered h2 heading, got %q", got)
	}
	if !strings.Contains(got, `<h3 id="with-code"><a class="heading-anchor" href="#with-code">13.1.1 Use <code>remco</code></a></h3>`) {
		t.Fatalf("expected inline heading markup to be preserved, got %q", got)
	}
}

func TestPreprocessAdmonitionsTrimsIndentedLines(t *testing.T) {
	content := "!!! note\n    First line.\n    Second line.\n"

	got := preprocessAdmonitions(content, makeGoldmark())

	if strings.Contains(got, "<pre><code>") {
		t.Fatalf("expected admonition body to render as markdown, got %q", got)
	}
	if strings.Contains(got, "    Second line.") {
		t.Fatalf("expected admonition indentation to be trimmed from every line, got %q", got)
	}
	if !strings.Contains(got, "First line.") || !strings.Contains(got, "Second line.") {
		t.Fatalf("expected admonition body to contain both lines, got %q", got)
	}
}

func TestPreprocessAdmonitionsSupportsTip(t *testing.T) {
	content := "!!! tip\n    Use templates.\n"

	got := preprocessAdmonitions(content, makeGoldmark())

	if !strings.Contains(got, `class="admonition tip"`) {
		t.Fatalf("expected tip admonition class, got %q", got)
	}
}

func TestResolveLinksRewritesDocLinks(t *testing.T) {
	anchorMap := map[string]string{
		"config/configuration-options.md": "doc-config-configuration-options",
	}

	got := resolveLinks("[options](../config/configuration-options.md)", "details/backends.md", anchorMap)
	if got != "[options](#doc-config-configuration-options)" {
		t.Fatalf("unexpected rewritten link: %q", got)
	}

	got = resolveLinks("[fragment](../config/configuration-options.md#backend)", "details/backends.md", anchorMap)
	if got != "[fragment](#backend)" {
		t.Fatalf("unexpected rewritten fragment link: %q", got)
	}
}

func TestNextUniqueIDHandlesGeneratedAndExplicitCollisions(t *testing.T) {
	counter := map[string]int{}
	ids := []string{
		nextUniqueID("bar", counter),
		nextUniqueID("bar", counter),
		nextUniqueID("bar-1", counter),
		nextUniqueID("bar", counter),
	}
	want := []string{"bar", "bar-1", "bar-1-1", "bar-2"}

	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("id %d = %q, want %q", i, ids[i], want[i])
		}
	}
}

func TestSectionAnchorUsesPath(t *testing.T) {
	first := sectionAnchor("details/backends.md")
	second := sectionAnchor("config/backends.md")

	if first == second {
		t.Fatalf("expected anchors to differ for different paths, got %q", first)
	}
	if first != "doc-details-backends" {
		t.Fatalf("unexpected first anchor: %q", first)
	}
	if second != "doc-config-backends" {
		t.Fatalf("unexpected second anchor: %q", second)
	}
}

func TestExtractFirstHeadingTrimsCRLF(t *testing.T) {
	content := "# Hello world\r\n\r\nBody\r\n"

	if got := extractFirstHeading(content); got != "Hello world" {
		t.Fatalf("extractFirstHeading() = %q, want %q", got, "Hello world")
	}
}

func TestNumberTOCEntriesNormalizesSectionsThatStartAtH3(t *testing.T) {
	entries := []tocEntry{{level: 3, text: "exists", id: "exists"}}

	numbered := numberTOCEntries(16, entries)
	if len(numbered) != 1 {
		t.Fatalf("expected one numbered entry, got %d", len(numbered))
	}
	if numbered[0].number != "16.1" {
		t.Fatalf("numbered[0].number = %q, want %q", numbered[0].number, "16.1")
	}
}

func TestBuildTOCAddsGroupsAndUsesValidNestedLists(t *testing.T) {
	sections := []section{
		{title: "remco", anchor: "doc-index", sourcePath: "index.md", group: groupForSection("index.md")},
		{title: "Template functions", anchor: "doc-template-template-functions", sourcePath: "template/template-functions.md", group: groupForSection("template/template-functions.md")},
	}
	entries := [][]numberedTOCEntry{
		{{level: 2, text: "Sections", id: "sections", number: "1.1"}},
		{{level: 2, text: "exists", id: "exists", number: "2.1"}, {level: 3, text: "default value", id: "default-value", number: "2.1.1"}},
	}

	got := buildTOC(sections, entries)

	if !strings.Contains(got, `class="toc-group-title">Overview</span>`) {
		t.Fatalf("expected overview group heading, got %q", got)
	}
	if !strings.Contains(got, `class="toc-group-title">Template Authoring</span>`) {
		t.Fatalf("expected template authoring group heading, got %q", got)
	}
	if strings.Contains(got, `<ul class="toc-entries">
<ul>`) {
		t.Fatalf("expected nested lists to be inside list items, got %q", got)
	}
	if !strings.Contains(got, `<li><div class="toc-entry-line"><span class="toc-num">2.1</span><a href="#exists">exists</a></div>
<ul>`) {
		t.Fatalf("expected child list to nest under parent list item, got %q", got)
	}
}
