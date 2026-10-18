package htmlsanitize

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	gohtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//go:embed assets/images/broken-file.svg
var BrokenFile string

// Constants for attribute handling
const (
	classDisallowed    = "disallowed"
	classBadHref       = "disallowed bad-href"
	classBadSrc        = "disallowed bad-src"
	attrTarget         = "target"
	attrTargetBlank    = "_blank"
	attrBadHref        = "bad-href"
	attrBadSrc         = "bad-src"
	attrClass          = "class"
	attrTitle          = "title"
	attrHref           = "href"
	attrSrc            = "src"
	titleDisallowedFmt = "(Link to '%s' is disallowed by policy)"
)

// HTML sanitizer comment template
const htmlSanitizerComment = "<!-- Removed by HTML sanitizer: %s -->"

// Set default list of forbidden HTML tags
// Note to Self: 'input' used to be on this list, but it was removed to ensure it doesn't break
//                 things like the Task List GFM extension (which uses bare input checkboxes). The 'input'
//                 element will still be removed if it's inside a 'form' (or one of the other
//                 elements on the list below)
var forbiddenTags = []string{
	"script", "dialog", "embed", "iframe",
	"form", "button", "select", "object",
	"applet", "base", "meta",
}

// Set default list of allowed file extensions
var allowedExts = []string{
	// Archive types
	".zip", ".rar", ".7z", ".tar", ".gz", "tgz", ".bz2",
	// Plain text Document types
	".css", "scss", ".htm", ".html", ".md", ".txt", ".xml",
	// Non-text Document types
	".docx", ".xlsx", ".pptx", ".pdf",
	// Images
	".png", ".apng", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp", ".ico", ".tiff", ".tif", ".icns",
	// Videos
	".mp4", ".ogg", ".webm", ".mov", ".avi", ".wmv", ".flv", ".mkv", ".m4v",
	// Audio
	".mp3", ".ogg", ".wav", ".flac", ".aac", ".m4a", ".wma", ".opus",
}


// SanitizeHTMLExtension implements Goldmark extension for markdown and raw HTML filtering.
// It provides security filtering for HTML content by:
// - Removing dangerous HTML tags (script, iframe, form, etc.)
// - Filtering URLs in href and src attributes based on file extensions
// - Adding target="_blank" to external links
// - Replacing disallowed content with safe alternatives or comments
type SanitizeHTMLExtension struct{}

// NewSanitizeHTMLExtension creates a new instance of the HTML sanitization extension.
// This follows Go best practices for constructor functions.
func NewSanitizeHTMLExtension() *SanitizeHTMLExtension {
	return &SanitizeHTMLExtension{}
}

// Extend registers our custom renderer with Goldmark.
// This follows the standard Goldmark extension pattern.
func (e *SanitizeHTMLExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&FilteredHTMLRenderer{}, 500),
	))
}

// FilteredHTMLRenderer handles HTML block and span rendering with security filtering
type FilteredHTMLRenderer struct{}

// RegisterFuncs registers rendering functions for different AST node types
func (r *FilteredHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindImage, r.renderImage)
}

// renderHTMLBlock processes HTML block elements
func (r *FilteredHTMLRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.HTMLBlock)
	var sb strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		sb.Write(line.Value(source))
	}

	filtered := filterHTML(sb.String())
	_, _ = w.WriteString(filtered)
	return ast.WalkContinue, nil
}

// renderRawHTML processes inline HTML elements
func (r *FilteredHTMLRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.RawHTML)
	var sb strings.Builder
	for i := 0; i < n.Segments.Len(); i++ {
		segment := n.Segments.At(i)
		sb.Write(segment.Value(source))
	}

	filtered := filterHTML(sb.String())
	_, _ = w.WriteString(filtered)
	return ast.WalkContinue, nil
}

// renderLink processes markdown links
func (r *FilteredHTMLRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	link := node.(*ast.Link)
	if !entering {
		_, _ = w.WriteString("</a>")
		return ast.WalkContinue, nil
	}

	newSVG := ""
	titleText := ""

	// Filter the destination URL
	dest := string(link.Destination)
	allowed := isAllowedAttributeValue(dest)

	// Check if link is internal anchor
	isInternalAnchor := strings.HasPrefix(dest, "#")

	// Check if link has existing target attribute
	hasTarget := false
	if link.Attributes() != nil {
		for _, attr := range link.Attributes() {
			if string(attr.Name) == attrTarget {
				hasTarget = true
				break
			}
		}
	}

	// Add start of <a> tag with href if URL is allowed
	if allowed {
		newSVG = ""
		_, _ = w.WriteString("<a href=\"")
		_, _ = w.Write(util.URLEscape(link.Destination, true))
		_ = w.WriteByte('"')
	} else {
		newSVG = resizeSVG(BrokenFile, 16, 16)
		_, _ = w.WriteString(`<a class="` + classBadHref + `" ` + attrBadHref + `="`)
		_, _ = w.Write(util.URLEscape(link.Destination, true))
		_ = w.WriteByte('"')
		titleText = fmt.Sprintf(titleDisallowedFmt, string(link.Destination))
	}

	// Add target="_blank" for external links without existing target
	if allowed && !isInternalAnchor && !hasTarget {
		_, _ = w.WriteString(` ` + attrTarget + `="` + attrTargetBlank + `"`)
	}

	// Render title if present
	if link.Title != nil {
		if titleText != "" {
			titleText = string(link.Title) + " " + titleText
		} else {
			titleText = string(link.Title)
		}
	}

	// Render title if present
	if titleText != "" {
		_, _ = w.WriteString(` ` + attrTitle + `="`)
		_, _ = w.Write(util.EscapeHTML([]byte(titleText)))
		_ = w.WriteByte('"')
	}

	// Render additional attributes
	if link.Attributes() != nil {
		for _, attr := range link.Attributes() {
			_ = w.WriteByte(' ')
			_, _ = w.Write(attr.Name)
			_, _ = w.WriteString(`="`)
			_, _ = w.Write(util.EscapeHTML(attr.Value.([]byte)))
			_ = w.WriteByte('"')
		}
	}
	_, _ = w.WriteString(`>`)
	if newSVG != "" {
		_, _ = w.WriteString(newSVG)
	}
	return ast.WalkContinue, nil
}

// renderImage processes markdown images
func (r *FilteredHTMLRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	img := node.(*ast.Image)
	_, _ = w.WriteString("<img")

	// Filter and render src attribute if URL is allowed
	dest := string(img.Destination)
	if isAllowedAttributeValue(dest) {
		_, _ = w.WriteString(` ` + attrSrc + `="`)
		_, _ = w.Write(util.URLEscape(img.Destination, true))
		_ = w.WriteByte('"')
	} else {
		// Render with bad-src attribute if URL is disallowed and set class for CSS
		_, _ = w.WriteString(` ` + attrClass + `="` + classBadSrc + `" ` + attrBadSrc + `="`)
		_, _ = w.Write(util.URLEscape(img.Destination, true))
		_ = w.WriteByte('"')
		_, _ = w.WriteString(` ` + attrTitle + `="`)
		_, _ = w.Write(util.EscapeHTML([]byte(fmt.Sprintf(titleDisallowedFmt, dest))))
		_ = w.WriteByte('"')
	}

	// Render alt text
	_, _ = w.WriteString(` alt="`)
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			_, _ = w.Write(textNode.Value(source))
		}
	}
	_ = w.WriteByte('"')

	// Render title if present
	if img.Title != nil {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(img.Title)
		_ = w.WriteByte('"')
	}

	// Render additional attributes
	if img.Attributes() != nil {
		for _, attr := range img.Attributes() {
			_ = w.WriteByte(' ')
			_, _ = w.Write(attr.Name)
			_, _ = w.WriteString(`="`)
			_, _ = w.Write(util.EscapeHTML(attr.Value.([]byte)))
			_ = w.WriteByte('"')
		}
	}
	_, _ = w.WriteString(" />")
	return ast.WalkSkipChildren, nil
}

// filterHTML processes raw HTML through our security filters
func filterHTML(raw string) string {
	// Handle incomplete HTML fragments (like just "<sup>" or "</sup>")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}

	// If it's just an opening or closing tag without content, pass it through
	// after security checks
	if strings.HasPrefix(raw, "<") && strings.HasSuffix(raw, ">") {
		// Check if it's a simple tag (opening or closing)
		if isSimpleHTMLTag(raw) {
			return filterSimpleTag(raw)
		}
	}

	contextNode := &gohtml.Node{
		Type:     gohtml.ElementNode,
		Data:     "html",
		DataAtom: atom.Html,
	}
	nodes, err := gohtml.ParseFragment(strings.NewReader(raw), contextNode)
	if err != nil {
		return raw // Return original on parse error
	}

	for _, n := range nodes {
		contextNode.AppendChild(n)
	}

	filterNode(contextNode)

	var buf bytes.Buffer
	if body := findBodyNode(contextNode); body != nil {
		for child := body.FirstChild; child != nil; child = child.NextSibling {
			_ = gohtml.Render(&buf, child)
		}
	} else {
		for child := contextNode.FirstChild; child != nil; child = child.NextSibling {
			_ = gohtml.Render(&buf, child)
		}
	}

	result := buf.String()

	// Post-process to fix URL escaping in href and src attributes
	// This reverses the unwanted &amp; -> & escaping for URLs that were allowed
	result = unescapeURLsInHTML(result)

	return result
}

// unescapeURLsInHTML reverses unwanted HTML escaping in href and src attributes
func unescapeURLsInHTML(html string) string {
	// Use regex to find href and src attributes and unescape URLs in them
	// This is safe because we only do this for URLs that passed our security filtering

	// Pattern to match href="..." or src="..." attributes
	hrefPattern := regexp.MustCompile(`((?:href|src)="[^"]*?)&amp;([^"]*")`)

	// TODO: CLEANUP - Consider using a more robust HTML attribute parser
	// instead of regex for better maintainability

	// Replace &amp; with & in URL attributes
	// Keep replacing until no more matches (handles multiple &amp; in one URL)
	for hrefPattern.MatchString(html) {
		html = hrefPattern.ReplaceAllString(html, "$1&$2")
	}

	return html
}// isSimpleHTMLTag checks if the string is a simple HTML tag (opening or closing)
func isSimpleHTMLTag(s string) bool {
	// Match patterns like <tag>, <tag attr="value">, </tag>
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "<") || !strings.HasSuffix(s, ">") {
		return false
	}

	// Remove < and >
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return false
	}

	// Check if it's a closing tag
	if strings.HasPrefix(inner, "/") {
		return true
	}

	// Check if it's an opening tag (with or without attributes)
	parts := strings.Fields(inner)
	if len(parts) == 0 {
		return false
	}

	// First part should be a valid tag name
	tagName := parts[0]
	return isValidTagName(tagName)
}

// filterSimpleTag applies security filtering to a simple HTML tag
func filterSimpleTag(tag string) string {
	// Extract tag name
	inner := strings.TrimSpace(tag[1 : len(tag)-1])

	// Handle closing tags
	if strings.HasPrefix(inner, "/") {
		tagName := strings.TrimSpace(inner[1:])
		if isForbiddenTag(tagName) {
			return fmt.Sprintf(htmlSanitizerComment, tagName)
		}
		return tag // Return as-is if allowed
	}

	// Handle opening tags
	parts := strings.Fields(inner)
	if len(parts) == 0 {
		return tag
	}

	tagName := parts[0]
	if isForbiddenTag(tagName) {
		return fmt.Sprintf(htmlSanitizerComment, tagName)
	}

	// For simple tags with just the tag name, return as-is
	if len(parts) == 1 {
		return tag
	}

	// For opening tags with attributes, manually filter attributes
	// without using gohtml.ParseFragment to avoid auto-closing
	return filterOpeningTagAttributes(tag, tagName)
}

// filterOpeningTagAttributes manually filters attributes in an opening tag
// without using gohtml.ParseFragment to avoid auto-closing the tag
func filterOpeningTagAttributes(tag, tagName string) string {
	// Parse attributes manually
	inner := strings.TrimSpace(tag[1 : len(tag)-1])

	// Check if it's a self-closing tag (XHTML style)
	isSelfClosing := strings.HasSuffix(inner, "/")
	if isSelfClosing {
		inner = strings.TrimSpace(inner[:len(inner)-1])
	}

	// Split on whitespace to separate tag name from attributes
	parts := strings.Fields(inner)
	if len(parts) <= 1 {
		// No attributes to filter - return original tag structure
		if isSelfClosing {
			return fmt.Sprintf("<%s/>", tagName)
		}
		return tag
	}

	// Reconstruct the tag with filtered attributes
	var result strings.Builder
	result.WriteString("<")
	result.WriteString(tagName)

	// Parse and filter each attribute
	attrString := strings.TrimSpace(inner[len(tagName):])
	filteredAttrs := parseAndFilterAttributes(attrString, tagName)

	for _, attr := range filteredAttrs {
		result.WriteString(" ")
		result.WriteString(attr)
	}

	// Close the tag appropriately
	if isSelfClosing {
		result.WriteString("/>")
	} else {
		result.WriteString(">")
	}
	return result.String()
}

// parseAndFilterAttributes parses attribute string and returns filtered attributes
func parseAndFilterAttributes(attrString, tagName string) []string {
	var attrs []string

	// Simple attribute parsing - this is a simplified version
	// In a production system, you might want more robust parsing
	parts := strings.Split(attrString, " ")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Look for key="value" pattern
		if strings.Contains(part, "=") {
			keyValue := strings.SplitN(part, "=", 2)
			if len(keyValue) == 2 {
				key := strings.TrimSpace(keyValue[0])
				value := strings.Trim(strings.TrimSpace(keyValue[1]), `"'`)

				// Disallow inline JavaScript event handlers
				if isEventHandler(key) {
					continue
				}

				// Filter the attribute
				if !isSensitiveAttribute(key) || isAllowedAttributeValue(value) {
					attrs = append(attrs, fmt.Sprintf(`%s="%s"`, key, value))
				} else {
					// Replace with bad- attributes using constants
					attrs = append(attrs, fmt.Sprintf(`bad-%s="%s"`, strings.ToLower(key), value))
					attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attrClass, classDisallowed+" bad-"+strings.ToLower(key)))
					attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attrTitle, fmt.Sprintf(titleDisallowedFmt, value)))
				}
			}
		} else {
			// Boolean attribute or malformed - just pass through
			attrs = append(attrs, part)
		}
	}

	// Add target="_blank" for external links (similar to addTargetBlank)
	hasTarget := false
	hasHref := false
	isInternalAnchor := false

	for _, attr := range attrs {
		if strings.HasPrefix(attr, `href="`) {
			hasHref = true
			value := strings.Trim(attr[6:], `"`)
			if strings.HasPrefix(value, "#") {
				isInternalAnchor = true
			}
		}
		if strings.HasPrefix(attr, `target="`) {
			hasTarget = true
		}
	}

	// Use the centralized function instead
	if strings.EqualFold(tagName, "a") && hasHref && !isInternalAnchor && !hasTarget {
		attrs = append(attrs, attrTarget+`="`+attrTargetBlank+`"`)
	}

	return attrs
}

// isValidTagName checks if a string is a valid HTML tag name
func isValidTagName(name string) bool {
	if name == "" {
		return false
	}

	// Simple check: alphanumeric characters, hyphens allowed
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			 (char >= 'A' && char <= 'Z') ||
			 (char >= '0' && char <= '9') ||
			 char == '-') {
			return false
		}
	}
	return true
}

// isForbiddenTag checks if a tag is in the forbidden list
func isForbiddenTag(tagName string) bool {
	// forbiddenTags := []string{
	// 	"script", "dialog", "embed", "iframe",
	// 	"form", "button", "select",
	// }

	for _, forbidden := range forbiddenTags {
		if strings.EqualFold(tagName, forbidden) {
			return true
		}
	}
	return false
}

// findBodyNode finds the body element in an HTML document structure
func findBodyNode(node *gohtml.Node) *gohtml.Node {
	if node.Type == gohtml.ElementNode && node.Data == "body" {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if body := findBodyNode(child); body != nil {
			return body
		}
	}
	return nil
}

// filterNodeList processes a list of HTML nodes
func filterNodeList(nodes []*gohtml.Node) []*gohtml.Node {
	var result []*gohtml.Node
	for _, node := range nodes {
		if filtered := filterNode(node); filtered != nil {
			result = append(result, filtered)
		}
	}
	return result
}

// filterNode applies security filters to an HTML node
func filterNode(node *gohtml.Node) *gohtml.Node {
	// Remove forbidden elements and their contents
	if node.Type == gohtml.ElementNode {
		// forbiddenTags := []string{
		// 	"script", "dialog", "embed", "iframe",
		// 	"form", "button", "select",
		// }
		for _, tag := range forbiddenTags {
			if strings.EqualFold(node.Data, tag) {
				return nil // Completely remove the element
			}
		}
	}

	// Filter attributes in allowed elements
	if node.Type == gohtml.ElementNode {
		filterAttributes(node)
	}

	// Add target="_blank" to external <a> tags
	if node.Type == gohtml.ElementNode && node.Data == "a" {
		addTargetBlank(node)
	}

	// For text nodes and other non-element nodes, return as-is
	if node.Type != gohtml.ElementNode {
		return node
	}

	// Process child nodes recursively
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling // Save next before filtering might change structure

		filtered := filterNode(child)
		if filtered == nil {
			// Child was removed, replace with comment
			comment := &gohtml.Node{
				Type: gohtml.CommentNode,
				Data: fmt.Sprintf(" Removed by HTML sanitizer: %s ", child.Data),
			}
			node.InsertBefore(comment, child)
			node.RemoveChild(child)
		}
		// If filtered is not nil and is the same node, it stays in place
		// If filtered is different, we would need to replace, but that's not expected in our case

		child = next
	}

	return node
}

// addTargetBlank adds target="_blank" to external links
func addTargetBlank(node *gohtml.Node) {
	// Check if this is an internal anchor link
	isInternalAnchor := false
	hasTarget := false
	hasHref := false

	for _, attr := range node.Attr {
		switch attr.Key {
		case attrHref:
			hasHref = true
			if strings.HasPrefix(attr.Val, "#") {
				isInternalAnchor = true
			}
		case attrTarget:
			hasTarget = true
		}
	}

	// Add target="_blank" for external links without existing target
	if hasHref && !isInternalAnchor && !hasTarget {
		node.Attr = append(node.Attr, gohtml.Attribute{
			Key: attrTarget,
			Val: attrTargetBlank,
		})
	}
}

// isEventHandler checks if an attribute is an inline event handler (e.g. onload, onerror, onclick)
func isEventHandler(attrName string) bool {
	lower := strings.ToLower(strings.TrimSpace(attrName))
	return strings.HasPrefix(lower, "on") && len(lower) > 2
}

// filterAttributes removes unsafe attributes from elements
func filterAttributes(node *gohtml.Node) {
	var attrs []gohtml.Attribute
	for _, attr := range node.Attr {
		// Disallow inline JavaScript event handlers
		if isEventHandler(attr.Key) {
			continue
		}
		if !isSensitiveAttribute(attr.Key) || isAllowedAttributeValue(attr.Val) {
			attrs = append(attrs, attr)
		} else {
			// TODO: CLEANUP - This logic is similar to parseAndFilterAttributes
			// Replace with bad- prefixed attributes using constants
			attrs = append(attrs, gohtml.Attribute{
				Key: "bad-" + strings.ToLower(attr.Key),
				Val: attr.Val,
			})
			attrs = append(attrs, gohtml.Attribute{
				Key: attrClass,
				Val: classDisallowed + " bad-" + strings.ToLower(attr.Key),
			})
			attrs = append(attrs, gohtml.Attribute{
				Key: attrTitle,
				Val: fmt.Sprintf(titleDisallowedFmt, attr.Val),
			})
		}
	}
	node.Attr = attrs
}

// isSensitiveAttribute checks if attribute needs filtering
func isSensitiveAttribute(attrName string) bool {
	sensitiveAttrs := []string{"action", "cite", "data", "href", "src"}
	for _, attr := range sensitiveAttrs {
		if strings.EqualFold(attrName, attr) {
			return true
		}
	}
	return false
}

// isAllowedAttributeValue determines if attribute value is safe
func isAllowedAttributeValue(value string) bool {
	if value == "" {
		return true
	}

	trimmedVal := strings.TrimSpace(value)
	lowerVal := strings.ToLower(trimmedVal)

	// Explicitly disallow pseudo-protocols and script execution schemes
	if strings.HasPrefix(lowerVal, "javascript:") ||
		strings.HasPrefix(lowerVal, "vbscript:") {
		return false
	}

	// Check data: URIs - allow safe raster image data URIs while blocking unsafe types (html, svg, script)
	if strings.HasPrefix(lowerVal, "data:") {
		return isAllowedDataURI(lowerVal)
	}

	// Allow fragment-only identifiers
	// (this only checks for values starting with #, not full URLs)
	if trimmedVal[0] == '#' {
		return isValidFragment(trimmedVal[1:])
	}

	// Allow directory paths
	if strings.HasSuffix(trimmedVal, "/") {
		return true
	}

	// Extract and analyze URL components
	u, err := url.Parse(trimmedVal)
	if err != nil {
		// Fallback to simple path extraction for invalid URLs
		_, ext := extractExtension(trimmedVal)
		return isExtensionAllowed(ext)
	}

	// Disallow any scheme other than http, https, mailto, tel, file
	if u.Scheme != "" {
		scheme := strings.ToLower(u.Scheme)
		if scheme != "http" && scheme != "https" && scheme != "mailto" && scheme != "tel" && scheme != "file" {
			return false
		}
	}

	// Handle special case: domain-only URLs (e.g., "example.com")
	if isDomainOnlyURL(u) {
		return true
	}

	// Check query string parameters for disallowed extensions
	if u.RawQuery != "" {
		if !areQueryParamsAllowed(u.RawQuery) {
			return false
		}
	}

	// Check to see if the URL has a fragment
	if u.Fragment != "" && !isValidFragment(u.Fragment) {
		return false
	}

	// Extract file extension from path
	base := path.Base(u.Path)
	ext := path.Ext(base)

	// If no extension found, allow the URL
	if ext == "" {
		return true
	}

	// Check if extension is allowed
	return isExtensionAllowed(ext)
}

// isAllowedDataURI checks if a data: URI is a safe raster image
func isAllowedDataURI(lowerVal string) bool {
	if !strings.HasPrefix(lowerVal, "data:") {
		return false
	}

	// Strictly allow only safe raster image MIME types (png, jpeg, jpg, gif, webp, bmp, ico)
	// SVG is intentionally excluded from user-provided data URIs because SVG can contain script/XML payloads.
	allowedPrefixes := []string{
		"data:image/png;",
		"data:image/jpeg;",
		"data:image/jpg;",
		"data:image/gif;",
		"data:image/webp;",
		"data:image/bmp;",
		"data:image/x-icon;",
		"data:image/vnd.microsoft.icon;",
	}

	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(lowerVal, prefix) {
			commaIdx := strings.Index(lowerVal, ",")
			if commaIdx == -1 {
				return false
			}
			meta := lowerVal[len("data:"):commaIdx]
			// Ensure metadata only specifies encoding/parameters and no script injection
			if strings.Contains(meta, "script") || strings.Contains(meta, "javascript") {
				return false
			}
			return true
		}
	}
	return false
}

// isDomainOnlyURL checks if URL represents a domain only (no path or file)
func isDomainOnlyURL(u *url.URL) bool {
	// Valid domain-only cases:
	// 1. Empty path (e.g., "http://example.com")
	// 2. Single slash path with no query parameters (e.g., "http://example.com/")
	return (u.Path == "" || u.Path == "/") && u.RawQuery == ""
}

// areQueryParamsAllowed checks if query string parameters contain any disallowed file extensions
func areQueryParamsAllowed(rawQuery string) bool {
	// Parse the query string
	queryValues, err := url.ParseQuery(rawQuery)
	if err != nil {
		// If we can't parse the query, allow it (conservative approach)
		return true
	}

	// Check each parameter value for file extensions
	for _, values := range queryValues {
		for _, value := range values {
			if containsDisallowedExtension(value) {
				return false
			}
		}
	}

	return true
}

// containsDisallowedExtension checks if a string contains a file path with disallowed extension
func containsDisallowedExtension(value string) bool {
	// Look for potential file paths in the value
	// This handles cases like: "file.exe", "path/to/file.dll", "download?file=virus.bat"

	// Split on common separators to find potential filenames
	separators := []string{"/", "\\", "=", "&", "?", " ", ",", ";"}
	parts := []string{value}

	// Split the value by each separator
	for _, sep := range separators {
		var newParts []string
		for _, part := range parts {
			subParts := strings.Split(part, sep)
			newParts = append(newParts, subParts...)
		}
		parts = newParts
	}

	// Check each part for file extensions
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Extract extension from this part
		ext := path.Ext(part)
		if ext != "" && !isExtensionAllowed(ext) {
			return true // Found disallowed extension
		}
	}

	return false
}

// isExtensionAllowed checks if a file extension is in the allowed list
func isExtensionAllowed(ext string) bool {
	ext = strings.ToLower(ext)
	if (ext == "") || (ext == ".") {
		return true
	}

	// If the extension is in the list of allowed extensions, return true
	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}

	return false // Forbid all other extensions by default
}

// extractExtension retrieves file extension from a string
func extractExtension(rawValue string) (string, string) {
	base := path.Base(rawValue)
	ext := path.Ext(base)
	return base, ext
}

// isValidFragment checks fragment identifier format
func isValidFragment(fragment string) bool {
	if fragment == "" {
		return true
	}
	validFragment := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validFragment.MatchString(fragment)
}

// resizeSVG modifies SVG content to set specific width and height attributes.
// If either width or height is 0, it uses the other dimension to maintain square aspect ratio.
func resizeSVG(svg string, width, height int) string {
	if width == 0 || height == 0 {
		if width == 0 {
			width = height
		}
		if height == 0 {
			height = width
		}
	}
	newSVG := regexp.MustCompile(`\s+width="[^"]+"`).ReplaceAllString(svg, fmt.Sprintf(` width="%dpx"`, width))
	newSVG = regexp.MustCompile(`\s+height="[^"]+"`).ReplaceAllString(newSVG, fmt.Sprintf(` height="%dpx"`, height))
	return newSVG
}