import type { app } from '../../wailsjs/go/models';

export function createConfigFixture(overrides: Partial<app.Config> = {}): app.Config {
  return {
    application: {
      use_inline_html: true,
      use_sanitize_html: true,
      use_strip_h1: false,
      use_frontmatter_title: true,
      use_auto_refresh: true,
      font_family: 'Verdana',
      font_size: 16,
      font_family_mono: 'Consolas',
      font_size_mono: 14,
      use_advanced_font_detection: false,
      ...(overrides.application ?? {}),
    },
    markdown: {
      use_gfm: true,
      use_php_md_ext: false,
      use_emoji: true,
      use_mermaid: true,
      use_figure: true,
      use_anchor: true,
      use_fences: true,
      use_sections: true,
      use_highlighting: true,
      use_fancylists: true,
      use_attributes: true,
      use_abbreviations: false,
      use_typographic: true,
      use_katex: true,
      use_d2_diagrams: true,
      use_supersubscript: true,
      ...(overrides.markdown ?? {}),
    },
    alert_callouts: {
      use_alertcallouts: true,
      alertcallout_style: 'GFMPlus',
      ...(overrides.alert_callouts ?? {}),
    },
  } as app.Config;
}