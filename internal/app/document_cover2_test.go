package app

// This file was created to expand test coverage without adding more tests to document_test.go.

import (
	"context"
	"reflect"
	"testing"
)

func TestCleanupHTMLContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "paragraph gets CRLF after preceding tag",
			input:    `<div>Intro</div>   <p>Body</p>`,
			expected: "<div>Intro</div>\r\n<p>Body</p>",
		},
		{
			name:     "closing code tag trims stray body and html whitespace",
			input:    `<pre><code>fmt.Println(1)</body> </html> </code></pre>`,
			expected: "<pre>\r\n<code>fmt.Println(1)</code></pre>",
		},
		{
			name:     "pre code html wrapper is collapsed with CRLF before code",
			input:    `<pre><code><html>value</code></pre>`,
			expected: "<pre>\r\n<code>value</code></pre>",
		},
		{
			name:     "pre code body wrapper is collapsed with CRLF before code",
			input:    `<pre>   <code>   <body class="x">value</code></pre>`,
			expected: "<pre>\r\n<code>value</code></pre>",
		},
		{
			name: "code content starts immediately after inserted CRLF",
			input: `<pre><code>
value</code></pre>`,
			expected: "<pre>\r\n<code>value</code></pre>",
		},
		{
			name:     "section gets CRLF after preceding tag",
			input:    `<div>Intro</div> <section class="section-h1">Body</section>`,
			expected: "<div>Intro</div>\r\n<section class=\"section-h1\">Body</section>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(CleanupHTMLContent([]byte(tt.input)))
			if got != tt.expected {
				t.Fatalf("CleanupHTMLContent() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestUpdateDocumentClasses(t *testing.T) {
	dp := &DocumentProcessor{
		ctx:      context.Background(),
		docTypes: []string{"techdoc", "api"},
	}
	events := captureDocumentRuntime(t)

	dp.updateDocumentClasses("guide tutorial")

	if !reflect.DeepEqual(dp.docTypes, []string{"guide", "tutorial"}) {
		t.Fatalf("docTypes = %#v, want %#v", dp.docTypes, []string{"guide", "tutorial"})
	}

	var got []string
	for _, event := range *events {
		classes, ok := event.data[0].([]string)
		if !ok {
			t.Fatalf("event payload type = %T, want []string", event.data[0])
		}
		if len(classes) != 1 {
			t.Fatalf("event classes = %#v, want single class entry", classes)
		}
		got = append(got, event.name+":"+classes[0])
	}

	want := []string{
		"remove-doc-class:techdoc",
		"remove-doc-class:api",
		"add-doc-class:guide",
		"add-doc-class:tutorial",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event sequence = %#v, want %#v", got, want)
	}
}

func TestUpdateDocumentClassesClearsStateWhenDocTypeEmpty(t *testing.T) {
	dp := &DocumentProcessor{
		ctx:      context.Background(),
		docTypes: []string{"release-notes"},
	}
	events := captureDocumentRuntime(t)

	dp.updateDocumentClasses("")

	if len(dp.docTypes) != 0 {
		t.Fatalf("docTypes length = %d, want 0", len(dp.docTypes))
	}
	if len(*events) != 1 {
		t.Fatalf("event count = %d, want 1", len(*events))
	}
	if (*events)[0].name != "remove-doc-class" {
		t.Fatalf("event name = %q, want %q", (*events)[0].name, "remove-doc-class")
	}
	classes, ok := (*events)[0].data[0].([]string)
	if !ok || len(classes) != 1 || classes[0] != "release-notes" {
		t.Fatalf("event payload = %#v, want []string{release-notes}", (*events)[0].data[0])
	}
}
