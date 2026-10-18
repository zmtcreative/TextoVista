// Interface for data passed from Go backend via markdown-rendered event
export interface MarkdownRenderData {
  html: string;
  title: string;
  date: string;
  frontmatter_html?: string;
  sanitize?: boolean;
  // Future fields can be added here without breaking existing functionality
  // type?: string;
  // metadata?: Record<string, string>;
}
