<template>
  <div class="app-toolbar">
    <div class="toolbar-content">
      <!-- Left side - reserved for future toolbar items -->
      <div class="toolbar-left">
        <!-- Future toolbar items can go here -->
      </div>

      <!-- Right side - theme toggle and other action items -->
      <div class="toolbar-right">
        <button
          v-if="canGoBack"
          @click="$emit('back')"
          class="toolbar-btn back-btn"
          title="Back to previous document"
        >
          <Icon
            name="arrow-left"
            :size="18"
            class="toolbar-icon"
          />
        </button>

        <button
          @click="$emit('refresh')"
          class="toolbar-btn refresh-btn"
          title="Refresh"
        >
          <Icon
            name="refresh"
            :size="18"
            class="toolbar-icon"
          />
        </button>

        <button
          @click="$emit('toggleFrontmatter')"
          class="toolbar-btn frontmatter-toggle-btn"
          :title="showFrontmatter ? 'Hide Frontmatter' : 'Show Frontmatter'"
        >
          <Icon
            name="list-tree"
            :size="18"
            class="toolbar-icon"
          />
        </button>

        <button
          @click="$emit('toggleTheme')"
          class="toolbar-btn theme-toggle-btn"
          :title="currentTheme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'"
        >
          <Icon
            v-if="currentTheme === 'light'"
            name="sun"
            :size="18"
            class="toolbar-icon"
          />
          <Icon
            v-if="currentTheme === 'dark'"
            name="moon"
            :size="18"
            class="toolbar-icon"
          />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from './Icon.vue'

interface Props {
  currentTheme: 'light' | 'dark'
  showFrontmatter: boolean
  canGoBack?: boolean
}

withDefaults(defineProps<Props>(), {
  canGoBack: false
})

defineEmits<{
  back: []
  refresh: []
  toggleTheme: []
  toggleFrontmatter: []
}>()
</script>

<style scoped>
.app-toolbar {
  flex-shrink: 0;
  position: sticky;
  top: 0;
  left: 0;
  right: 0;
  height: 28px;
  background-color: var(--toolbar-bg, #f8f9fa);
  border-bottom: 3px solid var(--toolbar-border, #e9ecef);
  border-top:  3px solid var(--toolbar-border, #e9ecef);
  z-index: 1000;
  /* box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15); */
}

.toolbar-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
  padding: 0 12px;
  max-width: 100%;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s ease;
  color: var(--toolbar-text, #495057);
  min-width: 32px;
  min-height: 28px;
}

.toolbar-btn:hover {
  background-color: var(--toolbar-hover, rgba(0, 0, 0, 0.1));
}

.toolbar-btn:hover .toolbar-icon {
  transform: scale(1.1);
}

.toolbar-icon {
  transition: transform 0.2s ease;
}

/* Dark theme styles */
:global(.dark) .app-toolbar {
  --toolbar-bg: #000000;
  --toolbar-border: var(--bg-color);
  --toolbar-text: #ffffff;
  --toolbar-hover: rgba(255, 255, 255, 0.25);
}

/* Light theme styles */
:global(.light) .app-toolbar {
  /* --toolbar-bg: #ffffff;
  --toolbar-border: #e2e8f0;
  --toolbar-text: #4a5568;
  --toolbar-hover: rgba(0, 0, 0, 0.05); */
  --toolbar-bg: #2b2b2b;
  --toolbar-border: var(--bg-color);
  --toolbar-text: #ffffff;
  --toolbar-hover: rgba(255, 255, 255, 0.25);
}
</style>
