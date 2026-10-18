package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	dateparse "github.com/araddon/dateparse"
	encodingUnicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"texto-vista/internal/markdown"
	"texto-vista/internal/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
)

var (
    documentEventsEmit = runtime.EventsEmit
    documentReadFile   = os.ReadFile
)

// MarkdownRenderData contains all data needed for the markdown-rendered event
type MarkdownRenderData struct {
	HTML            string `json:"html"`
	Title           string `json:"title"`
	Date            string `json:"date"`
	FrontmatterHTML string `json:"frontmatter_html,omitempty"`
	Sanitize        bool   `json:"sanitize"`
	// Future fields can be added here without breaking existing functionality
	// Type     string `json:"type,omitempty"`
	// Metadata map[string]string `json:"metadata,omitempty"`
}

// DocumentProcessor handles document processing and rendering
type DocumentProcessor struct {
    ctx               context.Context
    configManager     *ConfigManager
    docTypes          []string
    mdConverter       goldmark.Markdown
}

// NewDocumentProcessor creates a new DocumentProcessor
func NewDocumentProcessor(ctx context.Context, configManager *ConfigManager) *DocumentProcessor {
    return NewDocumentProcessorWithStyle(ctx, configManager)
}

// NewDocumentProcessorWithStyle creates a new DocumentProcessor with configuration manager
func NewDocumentProcessorWithStyle(ctx context.Context, configManager *ConfigManager) *DocumentProcessor {
    return &DocumentProcessor{
        ctx:               ctx,
        configManager:     configManager,
        docTypes:          []string{},
        mdConverter:       markdown.CreateGoldmarkInstance(configManager),
    }
}

// LoadAndDisplayMarkdown reads a Markdown file from the given path,
// converts its content to HTML using Goldmark, and then emits the HTML
// to the frontend via the "markdownLoaded" event.
func (dp *DocumentProcessor) LoadAndDisplayMarkdown(filePath string) error {
    mdContent, err := documentReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("file not found: %s", filePath)
        }
        if os.IsPermission(err) {
            return fmt.Errorf("permission denied to read file: %s", filePath)
        }
        return fmt.Errorf("failed to read file %s: %w", filePath, err)
    }

    // Note-to-Self:
    // Emit the initial loading state -- this will show a "Loading document..." message
    //   This will flash quickly for most document loads and probably won't be visible.
    //   HOWEVER, complex documents and those with D2 Diagrams may take longer to load, so
    //   this message will be visible.
    tmpData := MarkdownRenderData{
        HTML:        `<h3 style="text-align:center;color:green; border:0;">Loading document...</h3>`,
        Title:       "",
        Date:        "",
        FrontmatterHTML: "",
    }
    documentEventsEmit(dp.ctx, "markdown-rendered", tmpData)
    // End-of-Note

    // Detect and handle UTF-16 BOMs, and convert to UTF-8 if necessary
    if len(mdContent) >= 2 {
        bom := mdContent[:2]
        var transformer transform.Transformer
        if bom[0] == 0xFE && bom[1] == 0xFF { // UTF-16 BE
            transformer = encodingUnicode.UTF16(encodingUnicode.BigEndian, encodingUnicode.IgnoreBOM).NewDecoder()
        } else if bom[0] == 0xFF && bom[1] == 0xFE { // UTF-16 LE
            transformer = encodingUnicode.UTF16(encodingUnicode.LittleEndian, encodingUnicode.IgnoreBOM).NewDecoder()
        }

        if transformer != nil {
            utf8Content, _, err := transform.Bytes(transformer, mdContent)
            if err == nil {
                mdContent = utf8Content
            }
        }
    }

    mdContent = bytes.TrimPrefix(mdContent, []byte{0xEF, 0xBB, 0xBF})

    // Normalize line endings to Unix-style (LF)
    mdContent = []byte(strings.ReplaceAll(string(mdContent), "\r\n", "\n"))

    // Convert Markdown content to HTML
    htmlContent, docFrontmatter, thisDocumentH1Title, err := markdown.ConvertMarkdownToHTML(dp.mdConverter, mdContent, dp.configManager)
    if err != nil {
        return fmt.Errorf("failed to convert Markdown to HTML: %w", err)
    }

    // Process document metadata
    docTitle, docDate, docType := dp.processDocumentMetadata(filePath, thisDocumentH1Title, docFrontmatter)

    // Format frontmatter as HTML
    formatter := NewFrontmatterHTMLFormatter()
    frontmatterHTML := formatter.FormatAsHTML(docFrontmatter)

    // Cleanup HTML content
    htmlContent = CleanupHTMLContent(htmlContent)

    // Strip the first H1 element if configured
    if dp.configManager.UseStripH1() {
        htmlContent = stripFirstH1(htmlContent)
    }

    if dp.configManager.UseAbbreviations() {
        if abbrDefs, ok := docFrontmatter["__ABBR__"].(map[string]string); ok {
            htmlContent = markdown.ReplaceAbbreviationsInHTML(htmlContent, abbrDefs)
        }
    }

    // Resolve local relative image references relative to document directory
    htmlContent = resolveLocalImageSources(htmlContent, filepath.Dir(filePath))

    // Create structured data for the frontend
    renderData := MarkdownRenderData{
        HTML:            string(htmlContent),
        Title:           docTitle,
        Date:            docDate,
        FrontmatterHTML: frontmatterHTML,
        Sanitize:        dp.configManager.UseSanitize(),
    }

    // Emit the converted HTML to the frontend
    documentEventsEmit(dp.ctx, "markdown-rendered", renderData)

    // Handle document type classes
    dp.updateDocumentClasses(docType)

    return nil
}

// processDocumentMetadata extracts and formats document metadata
func (dp *DocumentProcessor) processDocumentMetadata(filePath, extractedTitle string, docFrontmatter map[string]any) (string, string, string) {
    var docTitle, docDate, fmDocTitle, fmDocDate, fmDocType string
    // docFileTitle := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
    docFileTitle := filepath.Base(filePath)
    timeLayout := time.DateTime + " MST"
    docDateLM := ""
    docDateDD := ""

    // Get file modification date
    if fileInfo, err := os.Stat(filePath); err == nil {
        fileModDate := fileInfo.ModTime()
        tz := fileModDate.Location()
        fmtLastModified := `<span class="date-label last-modified">Last Modified:</span> <span class="date-value last-modified">%s</span>`
        docDateLM = fmt.Sprintf(fmtLastModified, fileModDate.In(tz).Format(timeLayout))
    }

    // Extract frontmatter data - now using map[string]any the utils.GetValue generic to handle possible nil values
    if docFrontmatter != nil {
        if title, ok := utils.GetValue[string](docFrontmatter, "title"); ok {
            fmDocTitle = html.EscapeString(title)
        }
        if date, ok := utils.GetValue[time.Time](docFrontmatter, "date"); ok {
            fmDocDate = date.String()
        }
        if docType, ok := utils.GetValue[string](docFrontmatter, "doctype"); ok {
            fmDocType = strings.ToLower(docType)
        }
    }

    // Determine document title
    if fmDocTitle != "" && dp.configManager.UseFrontmatterTitle() {
        docTitle = fmDocTitle
    } else if extractedTitle != "" {
        docTitle = html.EscapeString(extractedTitle)
    } else {
        docTitle = fmt.Sprintf("File: %s", html.EscapeString(docFileTitle))
    }

    // Process document date from frontmatter
    if fmDocDate != "" {
        fmtDocDate := `<span class="date-label document-date">Document Date:</span> <span class="date-value document-date">%s</span>`
        tz := time.Now().Local().Location()
        dateString, err := dateparse.ParseIn(fmDocDate, tz)
        if err == nil {
            docDateDD = fmt.Sprintf(fmtDocDate, dateString.Format(timeLayout))
        } else {
            docDateDD = fmt.Sprintf(fmtDocDate, html.EscapeString(fmDocDate))
        }
    }

    // Combine dates
    if docDateDD != "" {
        docDate = docDateDD
    }
    if docDateLM != "" {
        if docDate == "" {
            docDate = docDateLM
        } else {
            docDate = docDate + "<br>" + docDateLM
        }
    }

    return docTitle, docDate, fmDocType
}

// updateDocumentClasses manages CSS classes on the document
func (dp *DocumentProcessor) updateDocumentClasses(docType string) {
    // Remove existing document type classes
    if dp.docTypes != nil {
        for _, dt := range dp.docTypes {
            dp.RemoveDocClass(dt)
        }
    }

    // Reset docTypes slice
    dp.docTypes = []string{}

    // Add new document type classes
    if docType != "" {
        docTypeArray := strings.Fields(docType)
        for _, dt := range docTypeArray {
            dp.AddDocClass(dt)
            dp.docTypes = append(dp.docTypes, dt)
        }
    }
}

// AddDocClass adds the class to html and body elements
func (dp *DocumentProcessor) AddDocClass(thisClass ...string) {
    documentEventsEmit(dp.ctx, "add-doc-class", thisClass)
}

// RemoveDocClass removes the class from html and body elements
func (dp *DocumentProcessor) RemoveDocClass(thisClass ...string) {
    documentEventsEmit(dp.ctx, "remove-doc-class", thisClass)
}

// ToggleDocClass toggles the class on html and body elements
func (dp *DocumentProcessor) ToggleDocClass(thisClass ...string) {
    documentEventsEmit(dp.ctx, "toggle-doc-class", thisClass)
}

// CleanupHTMLContent refines the generated HTML for better rendering.
func CleanupHTMLContent(htmlContent []byte) []byte {
    htmlString := string(htmlContent)

    re := regexp.MustCompile("(?si)" + `(>)\s*(<p>|<p\s+[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(?:(?:</body>|</html>)?\s)+(</code>)`)
    // htmlString = re.ReplaceAllString(htmlString, "\r\n$1")  // we do NOT want extra CRLF before closing </code>
    htmlString = re.ReplaceAllString(htmlString, "$1")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)(<code[^>]*>)(?:<html>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)\s*(<code[^>]*>)\s*(?:<body[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)\s*(<code[^>]*>)(?:\r\n|\n)*(\S+)`)
    // htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2\r\n$3") // we do NOT want extra CRLF after opening <code>
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2$3")

    re = regexp.MustCompile("(?si)" + `(>)\s*(<section[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    return []byte(htmlString)
}

// stripFirstH1 removes the first <h1> element and its contents from HTML content.
// If any other header elements (h2-h6) are encountered before the first h1,
// the original content is returned unchanged.
func stripFirstH1(htmlContent []byte) []byte {
    htmlString := string(htmlContent)

    // First pass: check if there are any h2-h6 elements before the first h1
    // This regex finds the first occurrence of any header element (h1-h6)
    headerPattern := regexp.MustCompile(`(?i)<h([1-6])[^>]*>`)
    headerMatch := headerPattern.FindStringSubmatch(htmlString)

    if headerMatch == nil {
        // No headers found, return original content
        return htmlContent
    }

    // Check if the first header found is not h1
    if headerMatch[1] != "1" {
        // First header is h2-h6, return original content unchanged
        return htmlContent
    }

    // Second pass: find and remove the first h1 element and its contents
    // This regex matches the opening h1 tag, its contents, and the closing tag
    // The (?s) flag makes . match newline characters as well
    h1Pattern := regexp.MustCompile(`(?is)<h1[^>]*>.*?</h1>`)

    // Find the first h1 match
    h1Match := h1Pattern.FindStringIndex(htmlString)
    if h1Match == nil {
        // No h1 found (shouldn't happen based on first pass, but safety check)
        return htmlContent
    }

    // Remove the first h1 element by concatenating the parts before and after it
    modifiedHTML := htmlString[:h1Match[0]] + htmlString[h1Match[1]:]

    return []byte(modifiedHTML)
}

var imgSrcRegex = regexp.MustCompile(`(?i)(<img\b[^>]*?\bsrc\s*=\s*)(?:"([^"]+)"|'([^']+)')([^>]*>)`)

// resolveLocalImageSources converts relative local image paths in <img> tags
// to embedded base64 data URIs so that WebView2/Chromium can display them without
// encountering 404s or browser local filesystem origin restrictions.
func resolveLocalImageSources(htmlContent []byte, docDir string) []byte {
    if len(htmlContent) == 0 {
        return htmlContent
    }

    return imgSrcRegex.ReplaceAllFunc(htmlContent, func(match []byte) []byte {
        submatches := imgSrcRegex.FindSubmatch(match)
        if len(submatches) < 5 {
            return match
        }

        quote := `"`
        srcVal := string(submatches[2])
        if srcVal == "" {
            quote = `'`
            srcVal = string(submatches[3])
        }
        lowerSrc := strings.ToLower(strings.TrimSpace(srcVal))

        // Ignore external schemes, data URIs, mailto, tel, javascript, and anchors
        if lowerSrc == "" ||
            strings.HasPrefix(lowerSrc, "http://") ||
            strings.HasPrefix(lowerSrc, "https://") ||
            strings.HasPrefix(lowerSrc, "//") ||
            strings.HasPrefix(lowerSrc, "data:") ||
            strings.HasPrefix(lowerSrc, "mailto:") ||
            strings.HasPrefix(lowerSrc, "tel:") ||
            strings.HasPrefix(lowerSrc, "javascript:") ||
            strings.HasPrefix(lowerSrc, "#") {
            return match
        }

        // Clean query parameters and fragments (e.g., "icon.svg#hash" or "foo.png?v=1")
        cleanSrc := srcVal
        if idx := strings.IndexAny(cleanSrc, "?#"); idx != -1 {
            cleanSrc = cleanSrc[:idx]
        }
        if unescaped, err := url.PathUnescape(cleanSrc); err == nil {
            cleanSrc = unescaped
        }

        var targetPath string
        if filepath.IsAbs(cleanSrc) {
            targetPath = cleanSrc
        } else {
            targetPath = filepath.Join(docDir, cleanSrc)
        }
        targetPath = filepath.Clean(targetPath)

        // Stat the target file
        info, err := os.Stat(targetPath)
        if err != nil || info.IsDir() {
            return match
        }

        // 20MB file size ceiling to prevent OOM / DoS
        if info.Size() > 20*1024*1024 {
            return match
        }

        ext := strings.ToLower(filepath.Ext(targetPath))
        var mimeType string
        switch ext {
        case ".svg":
            mimeType = "image/svg+xml"
        case ".png":
            mimeType = "image/png"
        case ".jpg", ".jpeg":
            mimeType = "image/jpeg"
        case ".gif":
            mimeType = "image/gif"
        case ".webp":
            mimeType = "image/webp"
        case ".bmp":
            mimeType = "image/bmp"
        case ".ico":
            mimeType = "image/x-icon"
        default:
            return match
        }

        data, err := documentReadFile(targetPath)
        if err != nil {
            return match
        }

        dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
        var sb strings.Builder
        sb.Write(submatches[1])
        sb.WriteString(quote)
        sb.WriteString(dataURI)
        sb.WriteString(quote)
        sb.Write(submatches[4])
        return []byte(sb.String())
    })
}

