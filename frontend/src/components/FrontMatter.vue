<template>
  <div v-if="isVisible" class="frontmatter-section">
    <div v-html="sanitizedFrontmatterHTML" class="frontmatter-content"></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DOMPurify from 'dompurify'

interface Props {
  frontmatterHTML?: string
  isVisible: boolean
}

const props = defineProps<Props>()

// Computed property to determine if we should show the section
const isVisible = computed(() => {
  return props.isVisible && props.frontmatterHTML && props.frontmatterHTML.trim().length > 0
})

const sanitizedFrontmatterHTML = computed(() => {
  if (!props.frontmatterHTML) return ''
  return DOMPurify.sanitize(props.frontmatterHTML, {
    ALLOWED_TAGS: ['div', 'span', 'p', 'b', 'strong', 'i', 'em', 'code', 'pre', 'br'],
    ALLOWED_ATTR: ['class', 'style'],
    ALLOW_DATA_ATTR: false
  })
})
</script>

<style scoped>
.frontmatter-section {
  flex-shrink: 0;
  position: sticky;
  top: 28px; /* Position below the toolbar (28px height) */
  left: 0;
  right: 0;
  background-color: var(--frontmatter-bg, #f8f9fa);
  border-bottom: 1px solid var(--frontmatter-border, #e9ecef);
  z-index: 999;
  /* max-height: 300px; */
  max-height: 25%;
  overflow-y: auto;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.frontmatter-content {
  padding: 12px 20px;
  font-family: var(--font-family-code, 'Monaco', 'Menlo', 'Ubuntu Mono', monospace);
  /* font-size: var(--font-size-code, 13px); */
  font-size: clamp(0.7rem, calc(var(--font-header-ratio) * 0.7), 1.125rem);
  line-height: 1;
}

/* Frontmatter syntax highlighting styles */
.frontmatter-content :deep(.frontmatter-container) {
  background-color: var(--frontmatter-content-bg, #ffffff);
  border-radius: 6px;
  border: 1px solid var(--frontmatter-content-border, #e1e4e8);
  padding: 3px 12px;
}

.frontmatter-content :deep(.frontmatter-header),
.frontmatter-content :deep(.frontmatter-footer) {
  color: var(--fm-delimiter-color, #6f42c1);
  font-weight: bold;
  margin: 0;
  padding: 4px 0;
}

.frontmatter-content :deep(.frontmatter-line) {
  margin: 2px 0;
  min-height: 1.25em;
}

.frontmatter-content :deep(.frontmatter-indent) {
  margin-left: 20px;
}

.frontmatter-content :deep(.fm-key) {
  color: var(--fm-key-color, #005cc5);
  font-weight: 500;
}

.frontmatter-content :deep(.fm-string) {
  color: var(--fm-string-color, #032f62);
}

.frontmatter-content :deep(.fm-number) {
  color: var(--fm-number-color, #005cc5);
}

.frontmatter-content :deep(.fm-boolean) {
  color: var(--fm-boolean-color, #e36209);
}

.frontmatter-content :deep(.fm-null) {
  color: var(--fm-null-color, #6f42c1);
  font-style: italic;
}

.frontmatter-content :deep(.fm-datetime) {
  color: var(--fm-datetime-color, #22863a);
}

.frontmatter-content :deep(.fm-array-marker) {
  color: var(--fm-array-marker-color, #6f42c1);
  margin-right: 4px;
}

/* Dark theme support */
:global(.dark) .frontmatter-section {
  --frontmatter-bg: #1a1a1a;
  --frontmatter-border: #404040;
  --frontmatter-content-bg: #2d2d2d;
  --frontmatter-content-border: #404040;
  --fm-delimiter-color: #bb86fc;
  --fm-key-color: #79c0ff;
  --fm-string-color: #a5d6ff;
  --fm-number-color: #79c0ff;
  --fm-boolean-color: #ffa657;
  --fm-null-color: #bb86fc;
  --fm-datetime-color: #56d364;
  --fm-array-marker-color: #bb86fc;
}

/* Light theme support */
:global(.light) .frontmatter-section {
  --frontmatter-bg: #f6f8fa;
  --frontmatter-border: #d1d9e0;
  --frontmatter-content-bg: #ffffff;
  --frontmatter-content-border: #e1e4e8;
  --fm-delimiter-color: #6f42c1;
  --fm-key-color: #005cc5;
  --fm-string-color: #032f62;
  --fm-number-color: #005cc5;
  --fm-boolean-color: #e36209;
  --fm-null-color: #6f42c1;
  --fm-datetime-color: #22863a;
  --fm-array-marker-color: #6f42c1;
}

/* Scrollbar styling for frontmatter section */
.frontmatter-section::-webkit-scrollbar {
  width: 8px;
}

.frontmatter-section::-webkit-scrollbar-track {
  background: var(--frontmatter-bg);
}

.frontmatter-section::-webkit-scrollbar-thumb {
  background: var(--frontmatter-border);
  border-radius: 4px;
}

.frontmatter-section::-webkit-scrollbar-thumb:hover {
  background: var(--fm-key-color);
}
</style>
