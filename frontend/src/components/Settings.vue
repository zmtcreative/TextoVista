<template>
  <Teleport to="body">
    <div
      id="settings-overlay"
      class="settings-overlay"
      v-show="show"
    >
      <div
        id="settings-dialog"
        class="settings-dialog"
        ref="settingsWindow"
      >
        <div class="settings-titlebar">
          <div
            class="window-title"
            @mousedown="startDrag"
          >
            <span>⚙️ Settings - TextoVista</span>
          </div>
          <div class="window-controls">
            <button
              id="settings-close"
              class="settings-close-button window-control-button"
              @click.stop="closeDialog"
              title="Close Settings"
            >
              &times;
            </button>
          </div>
        </div>
        <div class="settings-body">
          <h3 id="settings-title">Settings</h3>

          <!-- Tab Navigation - Fixed at top -->
          <div class="tab-navigation">
            <button
              type="button"
              class="tab-button"
              :class="{ 'active': activeTab === 'application' }"
              @click="activeTab = 'application'"
            >
              🏠 Application
            </button>
            <button
              type="button"
              class="tab-button"
              :class="{ 'active': activeTab === 'markdown' }"
              @click="activeTab = 'markdown'"
            >
              📝 Markdown
            </button>
            <button
              type="button"
              class="tab-button"
              :class="{ 'active': activeTab === 'alerts' }"
              @click="activeTab = 'alerts'"
            >
              ⚠️ Alerts/Callouts
            </button>
          </div>

          <form @submit.prevent="saveSettings" class="settings-form">
            <!-- Tab Content Area -->
            <div class="tab-content">

              <!-- Application Tab -->
              <div v-show="activeTab === 'application'" class="tab-panel">
                <fieldset id="fieldset-application" class="settings-section">
                  <legend>Application Settings</legend>

                  <!-- Inline HTML -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.application.use_inline_html"
                      />
                      <span class="checkbox-custom"></span>
                      Allow Inline HTML
                    </label>
                    <p class="setting-description">
                      Allow raw HTML tags to be rendered in markdown content (automatically sanitized for safety).
                    </p>
                  </div>

                  <!-- Strip First H1 -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.application.use_strip_h1"
                      />
                      <span class="checkbox-custom"></span>
                      Strip First H1 Heading
                    </label>
                    <p class="setting-description">
                      Remove the first H1 Heading when displaying the content.
                    </p>
                  </div>

                  <!-- Use Frontmatter Title -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.application.use_frontmatter_title"
                      />
                      <span class="checkbox-custom"></span>
                      Prefer Frontmatter Title
                    </label>
                    <p class="setting-description">
                      Prefer the "Title:" from the frontmatter instead of the first H1 heading.
                    </p>
                  </div>

                  <!-- Auto Refresh -->
                  <div class="setting-group">
                    <label class="checkbox-label" for="application-auto-refresh">
                      <input
                        id="application-auto-refresh"
                        type="checkbox"
                        v-model="localSettings.application.use_auto_refresh"
                      />
                      <span class="checkbox-custom"></span>
                      Auto Refresh Current File
                    </label>
                    <p class="setting-description">
                      Automatically reload and rerender the current markdown file when it changes on disk.
                    </p>
                  </div>

                  <!-- Separator -->
                  <div class="setting-group separator">&nbsp;</div>

                  <!-- Font Family Selection -->
                  <div class="setting-group">
                    <label class="setting-label">Font Family</label>
                    <select v-model="localSettings.application.font_family" class="setting-select">
                      <option v-for="font in availableFonts" :key="font" :value="font">
                        {{ font }}
                      </option>
                    </select>
                    <p class="setting-description">
                      Choose the font family for the application interface and content.
                    </p>
                  </div>

                  <!-- Font Size Selection -->
                  <div class="setting-group">
                    <label class="setting-label">Font Size</label>
                    <div class="font-size-control">
                      <input
                        type="range"
                        v-model.number="localSettings.application.font_size"
                        min="12"
                        max="24"
                        step="0.5"
                        class="setting-range"
                      />
                      <span class="font-size-value">{{ localSettings.application.font_size }}px</span>
                    </div>
                    <p class="setting-description">
                      Adjust the font size for better readability (12px - 24px).
                    </p>
                  </div>

                  <!-- Separator -->
                  <div class="setting-group separator">&nbsp;</div>

                  <!-- Monospace Font Family Selection -->
                  <div class="setting-group">
                    <label class="setting-label">Monospace Font Family</label>
                    <select v-model="localSettings.application.font_family_mono" class="setting-select">
                      <option v-for="font in availableMonospaceFonts" :key="font" :value="font">
                        {{ font }}
                      </option>
                    </select>
                    <p class="setting-description">
                      Choose the monospace font family for code blocks and inline code.
                    </p>
                  </div>

                  <!-- Monospace Font Size Selection -->
                  <div class="setting-group">
                    <label class="setting-label">Monospace Font Size</label>
                    <div class="font-size-control">
                      <input
                        type="range"
                        v-model.number="localSettings.application.font_size_mono"
                        min="10"
                        max="20"
                        step="0.5"
                        class="setting-range"
                      />
                      <span class="font-size-value">{{ localSettings.application.font_size_mono }}px</span>
                    </div>
                    <p class="setting-description">
                      Adjust the monospace font size for code elements (10px - 20px).
                    </p>
                  </div>

                  <!-- Advanced Font Detection -->
                  <div class="setting-group flex-full-width">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.application.use_advanced_font_detection"
                      />
                      <span class="checkbox-custom"></span>
                      Advanced Monospace Font Detection
                    </label>
                    <p class="setting-description">
                      Enable advanced detection to find more monospace fonts by analyzing font names and files. May take longer to load.
                    </p>
                  </div>
                </fieldset>
              </div>

              <!-- Markdown Options Tab -->
              <div v-show="activeTab === 'markdown'" class="tab-panel">
                <fieldset id="fieldset-markdown" class="settings-section">
                  <legend>Markdown Processing Options</legend>

                  <!-- GitHub Flavored Markdown -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_gfm"
                      />
                      <span class="checkbox-custom"></span>
                      GitHub Flavored Markdown
                    </label>
                    <p class="setting-description">
                      GFM Extensions (tables, strikethroughs, task lists & auto links).
                    </p>
                  </div>

                  <!-- PHP Markdown Extensions -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_php_md_ext"
                      />
                      <span class="checkbox-custom"></span>
                      PHP Markdown Extensions
                    </label>
                    <p class="setting-description">
                      PHP Markdown Extensions (footnotes & definition lists).
                    </p>
                  </div>

                  <!-- Fenced Code Highlighting -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_highlighting"
                      />
                      <span class="checkbox-custom"></span>
                      Fenced Code Highlighting
                    </label>
                    <p class="setting-description">
                      Enable syntax highlighting for code blocks.
                    </p>
                  </div>

                  <!-- Fancy Lists -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_fancylists"
                      />
                      <span class="checkbox-custom"></span>
                      Pandoc-Style Fancy Lists
                    </label>
                    <p class="setting-description">
                      Allow advanced list formatting with custom numbering.
                    </p>
                  </div>

                  <!-- Emoji Support -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_emoji"
                      />
                      <span class="checkbox-custom"></span>
                      Emoji Support
                    </label>
                    <p class="setting-description">
                      Convert emoji codes (like :smile:) to emoji characters.
                    </p>
                  </div>

                  <!-- Anchor Links on Headings -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_anchor"
                      />
                      <span class="checkbox-custom"></span>
                      Anchor Links on Headings
                    </label>
                    <p class="setting-description">
                      Add clickable anchor links to heading elements.
                    </p>
                  </div>

                  <!-- Mermaid Diagrams -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_mermaid"
                      />
                      <span class="checkbox-custom"></span>
                      Mermaid Diagrams
                    </label>
                    <p class="setting-description">
                      Enable Mermaid diagram rendering for charts and graphs.
                    </p>
                  </div>

                  <!-- D2 Diagrams -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_d2_diagrams"
                      />
                      <span class="checkbox-custom"></span>
                      D2 Diagrams
                    </label>
                    <p class="setting-description">
                      Enable D2 diagram rendering for flowcharts and diagrams.
                    </p>
                  </div>

                  <!-- KaTeX Math Support -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_katex"
                      />
                      <span class="checkbox-custom"></span>
                      KaTeX Math Support
                    </label>
                    <p class="setting-description">
                      Enable KaTeX math rendering for LaTeX formulas.
                    </p>
                  </div>

                  <!-- Image Figure Wrapping -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_figure"
                      />
                      <span class="checkbox-custom"></span>
                      Image Figure Wrapping
                    </label>
                    <p class="setting-description">
                      Wrap images in figure elements with captions.
                    </p>
                  </div>

                  <!-- Section Wrapping -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_sections"
                      />
                      <span class="checkbox-custom"></span>
                      Wrap Headings in SECTION Elements
                    </label>
                    <p class="setting-description">
                      Wrap headings and their content in HTML section elements.
                    </p>
                  </div>

                  <!-- Fenced DIV Blocks -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_fences"
                      />
                      <span class="checkbox-custom"></span>
                      Fenced DIVs
                    </label>
                    <p class="setting-description">
                      Enable fenced div blocks for custom content containers.
                    </p>
                  </div>

                  <!-- Custom Attributes -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_attributes"
                      />
                      <span class="checkbox-custom"></span>
                      Custom Attributes
                    </label>
                    <p class="setting-description">
                      Allow custom attributes using '{.myclass}' syntax.
                    </p>
                  </div>

                  <!-- Typographic Extensions -->
                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_typographic"
                      />
                      <span class="checkbox-custom"></span>
                      Typographic Extensions
                    </label>
                    <p class="setting-description">
                      Use fancy quotes and typographic enhancements.
                    </p>
                  </div>

                  <!-- Abbreviations Handling (Experimental) -->
                  <div class="setting-group experimental">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_abbreviations"
                      />
                      <span class="checkbox-custom"></span>
                      Abbreviations
                    </label>
                    <p class="setting-description">
                      Support for abbreviations (using '*[ABBR]: Abbreviations' syntax).
                    </p>
                  </div>
                </fieldset>
              </div>

              <!-- Alert Callouts Tab -->
              <div v-show="activeTab === 'alerts'" class="tab-panel">
                <fieldset id="fieldset-alerts" class="settings-section">
                  <legend>GFM/Obsidian - Alerts/Callouts</legend>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.alert_callouts.use_alertcallouts"
                      />
                      <span class="checkbox-custom"></span>
                      Alerts/Callouts
                    </label>
                    <p class="setting-description">
                      Use GitHub/Obsidian style alert/callout blocks.
                    </p>
                  </div>

                  <div class="setting-group alert-callout-style">
                    <label for="alert-callout-style" class="select-label">
                      Alert Callout Style:
                    </label>
                    <select
                      id="alert-callout-style"
                      v-model="localSettings.alert_callouts.alertcallout_style"
                      class="setting-select"
                    >
                      <option
                        v-for="(description, key) in alertCalloutStyles"
                        :key="key"
                        :value="key"
                      >
                        {{ description }}
                      </option>
                    </select>
                    <p class="setting-description">
                      Choose the style for rendering alert callouts (notes, warnings, etc.).
                    </p>
                  </div>
                </fieldset>
              </div>

            </div>

            <!-- Buttons -->
            <div class="settings-buttons">
              <button type="button" @click="resetToDefaults" class="button button-secondary">
                Reset to Defaults
              </button>
              <button type="button" @click="closeDialog" class="button button-secondary">
                Cancel
              </button>
              <button type="button" @click="saveSettingsSessionOnly" class="button button-info">
                Apply for Session
              </button>
              <button type="submit" class="button button-primary">
                Save to Disk
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue';
import { GetSettings, GetAlertCalloutStyles, SaveSettings, SaveSettingsSessionOnly, GetAvailableFonts, GetAvailableMonospaceFonts, GetAdvancedFontDetectionStatus, SetAdvancedFontDetection } from '../../wailsjs/go/main/App';
import { app } from '../../wailsjs/go/models';

// Define the settings interface to match the generated TypeScript model
type Settings = app.Config;

// Component props
const props = defineProps<{
  show: boolean;
}>();

// Component emits
const emit = defineEmits<{
  close: [];
  saved: [];
}>();

// Create default settings structure matching the new nested format
const createDefaultSettings = (): Settings => {
  return app.Config.createFrom({
    application: {
      use_inline_html: true,
      use_sanitize_html: true,
      use_strip_h1: true,
      use_frontmatter_title: true,
      use_auto_refresh: true,
      font_family: "Verdana, Arial, Helvetica, Tahoma, Geneva, sans-serif",
      font_size: 16.0,
      font_family_mono: "Consolas, Monaco, DejaVu Sans Mono, Liberation Mono, Courier New, Courier, monospace",
      font_size_mono: 14.0,
      use_advanced_font_detection: true
    },
    markdown: {
      use_gfm: true,
      use_php_md_ext: true,
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
      use_d2_diagrams: true
    },
    alert_callouts: {
      use_alertcallouts: true,
      alertcallout_style: 'GFMPlus'
    }
  });
};

// Reactive state
const localSettings = ref<Settings>(createDefaultSettings());
const alertCalloutStyles = ref<Record<string, string>>({});
const availableFonts = ref<string[]>([]);
const availableMonospaceFonts = ref<string[]>([]);
const loading = ref(false);

// Tab functionality - Default to application tab
const activeTab = ref<'application' | 'markdown' | 'alerts'>('application');

// Dialog window functionality
const settingsWindow = ref<HTMLElement>();
const isDragging = ref(false);
const dragOffset = ref({ x: 0, y: 0 });

// Load settings when dialog is shown
watch(() => props.show, async (newShow) => {
  if (newShow) {
    // Reset to the first tab when dialog opens
    activeTab.value = 'application';
    await loadSettings();
  }
});

// Synchronize sanitization setting with inline HTML setting for backwards compatibility
watch(() => localSettings.value.application.use_inline_html, (newInlineHTML) => {
  localSettings.value.application.use_sanitize_html = Boolean(newInlineHTML);
});

// Load current settings and alert callout styles
async function loadSettings() {
  try {
    loading.value = true;

    // Load current settings
    const currentSettings = await GetSettings();
    localSettings.value = app.Config.createFrom(currentSettings);

    // Load available alert callout styles
    alertCalloutStyles.value = await GetAlertCalloutStyles();

    // Load available fonts
    availableFonts.value = await GetAvailableFonts();

    // Load available monospace fonts
    availableMonospaceFonts.value = await GetAvailableMonospaceFonts();
  } catch (error) {
    console.error('Error loading settings:', error);
  } finally {
    loading.value = false;
  }
}

// Save settings
async function saveSettings() {
  try {
    loading.value = true;

    // Ensure font_size is a number
    if (localSettings.value.application.font_size) {
      localSettings.value.application.font_size = parseFloat(localSettings.value.application.font_size as any);
    }

    // Ensure font_size_mono is a number
    if (localSettings.value.application.font_size_mono) {
      localSettings.value.application.font_size_mono = parseFloat(localSettings.value.application.font_size_mono as any);
    }

    // Synchronize sanitization with inline HTML
    localSettings.value.application.use_sanitize_html = Boolean(localSettings.value.application.use_inline_html);

    // Create proper Config object before saving
    const configToSave = app.Config.createFrom(localSettings.value);
    await SaveSettings(configToSave);

    emit('saved');
    closeDialog();
  } catch (error) {
    console.error('Error saving settings:', error);
    // You might want to show an error message to the user here
    alert('Error saving settings: ' + error);
  } finally {
    loading.value = false;
  }
}

// Save settings for session only (don't write to disk)
async function saveSettingsSessionOnly() {
  try {
    loading.value = true;

    // Ensure font_size is a number
    if (localSettings.value.application.font_size) {
      localSettings.value.application.font_size = parseFloat(localSettings.value.application.font_size as any);
    }

    // Ensure font_size_mono is a number
    if (localSettings.value.application.font_size_mono) {
      localSettings.value.application.font_size_mono = parseFloat(localSettings.value.application.font_size_mono as any);
    }

    // Synchronize sanitization with inline HTML
    localSettings.value.application.use_sanitize_html = Boolean(localSettings.value.application.use_inline_html);

    // Create proper Config object before saving
    const configToSave = app.Config.createFrom(localSettings.value);
    await SaveSettingsSessionOnly(configToSave);

    emit('saved');
    closeDialog();
  } catch (error) {
    console.error('Error applying settings for session:', error);
    // You might want to show an error message to the user here
    alert('Error applying settings for session: ' + error);
  } finally {
    loading.value = false;
  }
}

// Reset to default values
function resetToDefaults() {
  localSettings.value = createDefaultSettings();

  // Fix font selection dropdowns to show proper defaults
  // Parse the default font family strings to extract the first font name
  const defaultFontFamily = "Verdana, Arial, Helvetica, Tahoma, Geneva, sans-serif";
  const defaultMonoFontFamily = "Consolas, Monaco, DejaVu Sans Mono, Liberation Mono, Courier New, Courier, monospace";

  // Extract first font name from the default font family string
  const firstDefaultFont = defaultFontFamily.split(',')[0].trim();
  const firstDefaultMonoFont = defaultMonoFontFamily.split(',')[0].trim();

  // Set the font family to the first available matching font, or fall back to first available
  if (availableFonts.value.length > 0) {
    // Try to find the default font in the available fonts
    const matchingFont = availableFonts.value.find(font =>
      font.toLowerCase().includes(firstDefaultFont.toLowerCase()) ||
      firstDefaultFont.toLowerCase().includes(font.toLowerCase())
    );
    localSettings.value.application.font_family = matchingFont || availableFonts.value[0];
  }

  if (availableMonospaceFonts.value.length > 0) {
    // Try to find the default mono font in the available monospace fonts
    const matchingMonoFont = availableMonospaceFonts.value.find(font =>
      font.toLowerCase().includes(firstDefaultMonoFont.toLowerCase()) ||
      firstDefaultMonoFont.toLowerCase().includes(font.toLowerCase())
    );
    localSettings.value.application.font_family_mono = matchingMonoFont || availableMonospaceFonts.value[0];
  }
}

// Close dialog
function closeDialog() {
  emit('close');
}

function startDrag(event: MouseEvent) {
  isDragging.value = true;

  if (settingsWindow.value) {
    const rect = settingsWindow.value.getBoundingClientRect();
    dragOffset.value = {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top
    };
  }

  document.addEventListener('mousemove', drag);
  document.addEventListener('mouseup', stopDrag);
  event.preventDefault();
  event.stopPropagation();
}

function drag(event: MouseEvent) {
  if (!isDragging.value || !settingsWindow.value) return;

  const x = event.clientX - dragOffset.value.x;
  const y = event.clientY - dragOffset.value.y;

  settingsWindow.value.style.left = `${x}px`;
  settingsWindow.value.style.top = `${y}px`;
}

function stopDrag() {
  isDragging.value = false;
  document.removeEventListener('mousemove', drag);
  document.removeEventListener('mouseup', stopDrag);
}

// Initialize when component mounts
loadSettings();
</script>

<style scoped>
/* CSS Variables - Define all the custom properties used in this component */
/* .settings-overlay {
  --dialog-color-bg-primary: #ffffff;
  --dialog-color-bg-secondary: #f8fafc;
  --dialog-color-bg-tertiary: #f1f5f9;
  --dialog-color-bg-hover: #f8fafc;
  --dialog-color-border: #e2e8f0;
  --dialog-color-border-light: #f1f5f9;
  --dialog-color-text-primary: #1e293b;
  --dialog-color-text-secondary: #64748b;
  --dialog-color-accent: #3b82f6;
  --dialog-color-accent-hover: #2563eb;
  --font-family-sans: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
} */

/* Settings overlay - covers the entire screen transparently */
.settings-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  z-index: 1000;
  display: flex;
  align-items: flex-start;
  justify-content: flex-start;
  backdrop-filter: blur(3px);
}

/* Settings dialog window */
.settings-dialog {
  background: var(--dialog-color-bg-primary);
  border: 1px solid var(--dialog-color-border);
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  min-width: 560px;
  max-width: 800px;
  width: 90vw;
  min-height: 420px;
  max-height: 85vh;
  position: relative;
  display: flex;
  flex-direction: column;
  font-family: var(--dialog-font-family-sans);
  top: 20px;
  left: 20px;
}

/* Titlebar */
.settings-titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--dialog-color-bg-secondary);
  border-bottom: 1px solid var(--dialog-color-border);
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  cursor: move;
  user-select: none;
}

.window-title {
  color: var(--dialog-color-text-primary);
  font-weight: 600;
  font-size: 14px;
  flex: 1;
  cursor: move;
}

.window-controls {
  display: flex;
  gap: 8px;
}

.window-control-button {
  background: none;
  border: none;
  color: var(--dialog-color-text-secondary);
  font-size: 18px;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.window-control-button:hover {
  background: var(--dialog-color-bg-tertiary);
  color: var(--dialog-color-text-primary);
}

.settings-close-button:hover {
  background: #e74c3c;
  color: white;
}

/* Settings body */
.settings-body {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden; /* Prevent body scrolling */
}

#settings-title {
  margin: 20px 20px 0 20px;
  color: var(--dialog-color-text-primary);
  font-size: 18px;
  font-weight: 600;
  display: none;
}

/* Tab navigation */
.tab-navigation {
  display: flex;
  margin: 20px 16px 0 16px;
  padding-bottom: 0px;
  border-bottom: 4px solid var(--dialog-color-border);
  gap: 1px;
  flex-shrink: 0; /* Don't shrink this section */
}

.tab-button {
  background: var(--dialog-color-bg-secondary);
  border: none;
  color: var(--dialog-color-text-secondary);
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s ease;
  border-radius: 8px 8px 0 0;
  position: relative;
  bottom: 1px;
}

.tab-button:hover {
  background: var(--dialog-color-bg-tertiary);
  color: var(--dialog-color-text-primary);
}

.tab-button.active {
  background: var(--dialog-color-bg-primary);
  color: var(--dialog-color-tab-accent);
  border: 2px solid var(--dialog-color-bg-secondary);
  border-bottom: 2px solid var(--dialog-color-accent);

}

/* Tab content */
.tab-content {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
  margin: 4px 0px 4px 0px;
  min-height: 240px; /* Allow shrinking below content size */
}

.tab-panel {
  animation: fadeIn 0.2s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Settings form */
.settings-form {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden; /* Let child elements handle scrolling */
  padding: 0 20px;
}

.settings-section {
  border: 1px solid var(--dialog-color-border);
  border-radius: 6px;
  margin-bottom: 20px;
  padding: 0;
  background: var(--dialog-color-bg-primary);
  display: flex;
  flex-flow: row wrap;
  gap: 0px 0px;
}

.settings-section legend {
  color: var(--dialog-color-text-primary);
  font-weight: 600;
  font-size: 16px;
  padding: 0 6px;
  margin-left: 6px;
}

.settings-section > div:first-of-type {
  padding-top: 8px;
}

.settings-section > div:last-of-type {
  padding-bottom: 8px;
}

/* Setting groups */
.setting-group.separator {
  /* border-bottom: 1px solid var(--dialog-color-border-light);
  border-right: 1px solid var(--dialog-color-border-light); */
  border: 0;
  border-bottom: 1px solid var(--dialog-color-border-light);
  height: 1px;
  background-color: var(--dialog-color-border-light);
  flex: 1 0 90%;
}

.setting-group {
  padding: 4px 10px 4px 20px;
  border-bottom: 1px solid var(--dialog-color-border-light);
  border-right: 1px solid var(--dialog-color-border-light);
  flex: 1 0 40%;
}

.setting-group.flex-full-width {
  flex: 1 0 90%;
}

#fieldset-alerts .setting-group {
  flex: 1 0 50%;
}

.setting-group:last-child {
  border-bottom: none;
}

.setting-group:hover {
  background: var(--dialog-color-bg-hover);
}

.setting-group.disabled-setting {
  opacity: 0.5;
  pointer-events: none;
}

.setting-group.disabled-setting .checkbox-label {
  cursor: not-allowed;
}

.setting-group.disabled-setting input[type="checkbox"]:disabled {
  cursor: not-allowed;
}

.setting-group.experimental p::before {
  content: '[Experimental] ';
  color: red;
}

/* Checkbox styling */
.checkbox-label {
  display: flex;
  align-items: center;
  color: var(--dialog-color-text-primary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  gap: 12px;
  line-height: 1.15;
}

.checkbox-label input[type="checkbox"] {
  display: none;
}

.checkbox-custom {
  width: 15px;
  height: 15px;
  border: 2px solid var(--dialog-color-border);
  border-radius: 3px;
  background: var(--dialog-color-bg-primary);
  position: relative;
  top: 5px;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom {
  background: var(--dialog-color-accent);
  border-color: var(--dialog-color-accent);
}

.checkbox-custom::after {
  content: '';
  position: absolute;
  width: 4px;
  height: 9px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
  top: 0px;
  left: 5px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom::after {
  opacity: 1;
}

.checkbox-custom:hover {
  border-color: var(--dialog-color-tab-accent);
}

/* Select styling */
.select-label {
  display: block;
  color: var(--dialog-color-text-primary);
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 8px;
}

.setting-select {
  width: 100%;
  padding: 5px 8px;
  border: 1px solid var(--dialog-color-border);
  border-radius: 4px;
  background: var(--dialog-color-bg-primary);
  color: var(--dialog-color-text-primary);
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  transition: border-color 0.2s ease;
}

.setting-select:hover {
  border-color: var(--dialog-color-accent);
}

.setting-select:focus {
  outline: none;
  border-color: var(--dialog-color-accent);
  box-shadow: 0 0 0 2px rgba(74, 144, 226, 0.2);
}

/* Font controls styling */
.font-size-control {
  display: flex;
  align-items: center;
  gap: 12px;
}

.setting-range {
  flex: 1;
  height: 6px;
  background: var(--dialog-color-border);
  border-radius: 3px;
  outline: none;
  cursor: pointer;
  transition: background 0.2s ease;
  -webkit-appearance: none;
  appearance: none;
}

.setting-range::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  background: var(--dialog-color-accent);
  border-radius: 50%;
  cursor: pointer;
  transition: background 0.2s ease;
}

.setting-range::-webkit-slider-thumb:hover {
  background: var(--dialog-color-accent-hover);
}

.setting-range::-moz-range-thumb {
  width: 18px;
  height: 18px;
  background: var(--dialog-color-accent);
  border-radius: 50%;
  cursor: pointer;
  border: none;
  transition: background 0.2s ease;
}

.setting-range::-moz-range-thumb:hover {
  background: var(--dialog-color-accent-hover);
}

.font-size-value {
  min-width: 45px;
  text-align: right;
  font-size: 12px;
  font-weight: 500;
  color: var(--dialog-color-text-secondary);
}

.setting-label {
  display: block;
  color: var(--dialog-color-text-primary);
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 8px;
}

/* Setting descriptions */
.setting-description {
  margin: 0.0625rem 0 0 33px;
  color: var(--dialog-color-text-secondary);
  font-size: 11px;
  line-height: 1.15;
}

#fieldset-alerts .setting-group.alert-callout-style {
  padding-left: 50px;
}

#fieldset-alerts .alert-callout-style p.setting-description {
  margin-top: 0.15rem;
  margin-left: 0.75rem;
}

/* Buttons */
.settings-buttons {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  padding: 12px 0;
  border-top: 4px solid var(--dialog-color-border);
  flex-shrink: 0; /* Don't shrink this section */
}

.button {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 80px;
}

.button-primary {
  background: var(--dialog-color-accent);
  color: white;
}

.button-primary:hover {
  background: var(--dialog-color-accent-hover);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.3);
}

.button-secondary {
  background: var(--dialog-color-bg-secondary);
  color: var(--dialog-color-text-primary);
  border: 1px solid var(--dialog-color-border);
}

.button-secondary:hover {
  background: var(--dialog-color-bg-tertiary);
  transform: translateY(-1px);
}

.button-info {
  background: #10b981;
  color: white;
  border: 1px solid #059669;
}

.button-info:hover {
  background: #059669;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.button:active {
  transform: translateY(0);
}

/* Scrollbar styling */
.tab-content::-webkit-scrollbar {
  width: 8px;
}

.tab-content::-webkit-scrollbar-track {
  background: var(--dialog-color-bg-secondary);
  border-radius: 4px;
}

.tab-content::-webkit-scrollbar-thumb {
  background: var(--dialog-color-border);
  border-radius: 4px;
}

.tab-content::-webkit-scrollbar-thumb:hover {
  background: var(--dialog-color-text-secondary);
}

@media (min-height: 580px) {
  .settings-dialog {
    min-height: 520px;
  }
  .tab-content {
    min-height:342px;
  }
}

/* Responsive adjustments */
@media (max-width: 560px) {
  .settings-dialog {
    min-width: unset;
    width: 95vw;
    margin: 20px;
  }

  .setting-group {
    padding: 8px 10px 8px 20px;
    border-bottom: 1px solid var(--dialog-color-border-light);
    border-right: 1px solid var(--dialog-color-border-light);
    flex: 1 0 100%;
  }

  .tab-navigation {
    flex-direction: column;
  }

  .tab-button {
    border-radius: 4px;
    margin-bottom: 2px;
  }

  .settings-buttons {
    flex-direction: column-reverse;
  }

  .button {
    width: 100%;
  }
}
</style>
