# TextoVista

[![License](https://img.shields.io/github/license/zmtcreative/markdown-reader)](./LICENSE.md)
![GitHub Release](https://img.shields.io/github/v/release/zmtcreative/markdown-reader?sort=semver)

Copyright &copy; 2026 ZMT Creative LLC. All rights reserved.

## About and Features

TextoVista is a Markdown reader written in Go using the Wails + Vue framework.

The reader supports the current CommonMark spec plus a wide range of Extension:

(*All options below are enabled by default unless noted otherwise*)

- GitHub Flavored Markdown and PHP Markdown Extensions
  - GFM: Tables
  - GFM: Strikethrough
  - GFM: Autolinks
  - GFM: Task List Items
  - PHP: Definition Lists
  - PHP: Footnotes
- Typographic Enhancements (*like [smartypants](https://daringfireball.net/projects/smartypants/)*)
  - Fancy quotes
  - Proper en-dash and em-dash
  - Proper ellipsis rendering
- Emojis
- Inline HTML  (*allows inclusion of bare HTML in Markdown -- __Potentially Unsafe__*)
- Sanitize HTML  (*removes some potential dangerous HTML elements and disables potentially dangerous URLs with executable file extensions in the URL -- __should be enabled when using Inline HTML__*)
- Frontmatter Parsing  (*currently can use the __Title:__ and __Date:__ entries in markdown frontmatter*)
- Mermaid Diagrams and Charts
- Image Figure Wrapping  (*provides a syntax for including captions with images*)
<!-- - [//]#: - Header Anchor Links -->
- Fenced DIVs  (*works similar to fenced code blocks but uses `:::` to surround sections you want to wrap with a DIV element*)
- Header Section Wrapper  (*nested wrapping of heading and all their content to keep section levels together since `<H?>` headings don't wrap the sections they represent*)
- Code Syntax Highlighting  (*provides syntax highlighting of fenced code blocks*)
- Pandoc-Style Fancy Lists  (*creating ordered lists using letters and roman numerals*)
- Alerts/Callouts  (*uses `> [!NOTE]` style markdown syntax to create GFM and/or Obsidian Callouts*)
- Block Attributes  (*allows you to add recognized HTML attributes to block elements using the `{.myclass #myid title="mytitle"}` syntax*)
- Bracketed Spans  (*allows you to wrap some text in brackets to create spans -- `this is [some text]{.myclass}` that would wrap the `some text` and render as `<span class="myclass">some text</span>`*)
- Proportional Font and Code Font Customization (*you can set the font family and font size you want the reader to use*)

## Status

> [!NOTE]
>
> Still in Beta stage.

The application is useable in its current form, with extensive settings available to control
which Markdown/Commonmark features and Extensions are displayed. The `File | Settings` dialog lets you change settings just for the current session or save the settings to a config file.

The CSS styling for the application is still in-progress, so the look of rendered markdown may
not be consistent, though it should be adequate for now.

__Very__ rudimentary printing support is available, but the CSS styling for printing has
not been tweaked, so print output will be useable but not necessarily ideal.

Development is still ongoing.

## Collaboration

This application is currently being privately developed and we are not (currently) accepting collaborators, feature requests or pull-requests.

The repository __is__ public and you are welcome to fork the project and modify it as you see fit, provided you adhere to the [LICENSE](LICENSE) and provide attribution in your `README.md` or a `NOTICE.md` file in your forked repository.
