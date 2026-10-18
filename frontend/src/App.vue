<template>
  <div class="app-container">
    <!-- Application Toolbar -->
    <Toolbar
      :currentTheme="currentTheme"
      :showFrontmatter="showFrontmatter"
      @refresh="refreshCurrentDocument"
      @toggleTheme="toggleTheme"
      @toggleFrontmatter="toggleFrontmatter"
    />

    <!-- FrontMatter Section (between toolbar and header) -->
    <FrontMatter
      :frontmatterHTML="frontmatterHTML"
      :isVisible="showFrontmatter"
    />

    <!-- Fixed Header -->
    <header class="app-header" :class="{ 'with-frontmatter': showFrontmatter && frontmatterHTML }">
      <h1 v-html="docHTMLTitle" class="document-title"></h1>
      <p v-html="docHTMLDate" class="document-dates"></p>
      <!-- Display error message if preset -->
      <p v-if="errorMessage" class="error-message">Error: {{ errorMessage }}</p>
    </header>

    <!-- Scrollable Main Content Area -->
    <main class="content-area">
      <div v-if="errorMessage" id="content" class="markdown-body error-content">
        <h1>An error occurred:</h1>
        <p>{{ errorMessage }}</p>
        <p>Please try opening another file or check the file path.</p>
      </div>
      <article v-else v-html="renderedHTML" id="content" class="markdown-body"></article>
    </main>
  </div>

  <!-- Help Dialog -->
  <Help
    ref="helpRef"
    :show="showHelpModal"
    @close="hideHelpModal"
  />

  <!-- Settings Dialog -->
  <Settings
    :show="showSettingsDialog"
    @close="hideSettingsDialog"
    @saved="onSettingsSaved"
  />
</template>

<script setup lang="ts">
import { ref, watch, nextTick,onMounted, onUnmounted } from 'vue';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import { GetTheme, SetTheme, GetCurrentFont, GetCurrentMonospaceFont, HasCurrentFile, ReloadCurrentDocument } from '../wailsjs/go/main/App';
import Settings from './components/Settings.vue';
import Help from './components/Help.vue';
import Toolbar from './components/Toolbar.vue';
import FrontMatter from './components/FrontMatter.vue';
import type { MarkdownRenderData } from './types/markdown';
import mermaid from 'mermaid';
import katex from 'katex';
import DOMPurify from 'dompurify';

// const renderedHTML = ref('');
const renderedHTML = ref('<h3 style="text-align:center;color:green; border:0;">Loading document...</h3>');
const docHTMLTitle = ref('');
const docHTMLDate = ref('');
const frontmatterHTML = ref('');
const showFrontmatter = ref(false);
const errorMessage = ref('');

// Modal reactive variables
const showHelpModal = ref(false);

// Settings dialog reactive variables
const showSettingsDialog = ref(false);

// Help component reference
const helpRef = ref<InstanceType<typeof Help>>();

const currentTheme = ref<'light' | 'dark'>('light');

// Function to add class to html and body elements
function addDocClass(thisClass: string) {
  document.documentElement.classList.add(thisClass);
  document.body.classList.add(thisClass);
}

// Function to remove class from html and body elements
function removeDocClass(thisClass: string) {
  document.documentElement.classList.remove(thisClass);
  document.body.classList.remove(thisClass);
}

// Function to toggle class on html and body elements
function toggleDocClass(thisClass: string) {
  const hasClass = document.documentElement.classList.contains(thisClass);
  if (hasClass) {
    removeDocClass(thisClass);
  } else {
    addDocClass(thisClass);
  }
}

// Function to toggle the theme
async function toggleTheme() {
  const newTheme = currentTheme.value === 'light' ? 'dark' : 'light';
  await SetTheme(newTheme); // Call Go backend to set the new theme
  currentTheme.value = newTheme;
}

// Function to toggle the frontmatter visibility
function toggleFrontmatter() {
  showFrontmatter.value = !showFrontmatter.value;
}

async function refreshCurrentDocument() {
  try {
    await ReloadCurrentDocument();
  } catch (error) {
    console.error('Error refreshing current document:', error);
    errorMessage.value = String(error);
    renderedHTML.value = '';
  }
}

// Watch for changes in the theme and update the body class
watch(currentTheme, (newTheme, oldTheme) => {
  if (oldTheme) {
    document.body.classList.remove(oldTheme);
  }
  document.body.classList.add(newTheme);
  // Also update the <html> element if needed
  document.documentElement.className = newTheme;

  // Re-initialize Mermaid with the correct theme
  mermaid.initialize({
    startOnLoad: false,
    theme: newTheme === 'dark' ? 'dark' : 'default',
  });
}, { immediate: true }); // immediate: true runs the watcher on component mount

// Print function
function printContent() {
    window.print();
}

// Save as PDF function (using browser's print to PDF)
function saveAsPDF(filePath: string) {
    // For web browsers, we can trigger print dialog with PDF option
    // The actual file saving needs to be handled differently in Wails
    window.print();
}

// Function to hide the modal
function hideHelpModal() {
    showHelpModal.value = false;
}

// Function to show the modal
function showHelpModalDialog(helpTitle: string, helpText: string) {
    if (helpRef.value) {
        helpRef.value.showHelpModalDialog(helpTitle, helpText);
        showHelpModal.value = true;
    }
}

// Settings dialog functions
function hideSettingsDialog() {
    showSettingsDialog.value = false;
}

function openSettingsDialog() {
    showSettingsDialog.value = true;
}

function onSettingsSaved() {
    // Settings have been saved successfully
    // You might want to emit an event or update the UI here
    console.log('Settings saved successfully');

    // Call setRootFontSizeNum after settings are saved to update CSS custom variables
    nextTick(() => {
        updateFontSettings();
        setRootFontSizeNum();
    });
}

async function updateFontSettings() {
    try {
        // Get current font settings from backend
        const fontSettings = await GetCurrentFont();
        const monospaceFontSettings = await GetCurrentMonospaceFont();

        if (fontSettings) {
            const root = document.documentElement;

            // Update font family CSS variable
            if (fontSettings.fontFamily) {
                root.style.setProperty('--font-family-base', `${fontSettings.fontFamily}`);
                root.style.setProperty('font-family', `${fontSettings.fontFamily}`);
                console.log(`Updated --font-family-base to: ${fontSettings.fontFamily}`);
            }

            // Update font size CSS variables
            if (fontSettings.fontSize) {
                root.style.setProperty('--font-size-base', `${fontSettings.fontSize}px`);
                root.style.setProperty('font-size', `${fontSettings.fontSize}px`);
                root.style.setProperty('--font-size-base-px', `${fontSettings.fontSize}px`);
                console.log(`Updated --font-size-base to: ${fontSettings.fontSize}px`);
            }
        }

        if (monospaceFontSettings) {
            const root = document.documentElement;

            // Update monospace font family CSS variable
            if (monospaceFontSettings.fontFamily) {
                root.style.setProperty('--font-family-code', `${monospaceFontSettings.fontFamily}`);
                console.log(`Updated --font-family-code to: ${monospaceFontSettings.fontFamily}`);
            }

            // Update monospace font size CSS variable
            if (monospaceFontSettings.fontSize) {
                root.style.setProperty('--font-size-code', `${monospaceFontSettings.fontSize}px`);
                console.log(`Updated --font-size-code to: ${monospaceFontSettings.fontSize}px`);
            }
        }
    } catch (error) {
        console.error('Error updating font settings:', error);
    }
}

function setRootFontSizeNum() {
    const root = document.documentElement;

    // 1. Retrieve the current value of the CSS custom property.
    // We use getComputedStyle() to get the live, computed value of the property.
    const currentSizeWithUnit = getComputedStyle(root).getPropertyValue('--font-size-base-px').trim();

    // 2. Strip the unit to get the number.
    // Use parseFloat for better handling of decimal values, then convert to string
    const currentSizeNumber = parseFloat(currentSizeWithUnit);

    // 3. Validate the parsed number and update the custom property
    if (!isNaN(currentSizeNumber) && currentSizeNumber > 0) {
        root.style.setProperty('--font-size-base-num', currentSizeNumber.toString());
        console.log(`Updated --font-size-base-num to: ${currentSizeNumber}`);
    } else {
        console.warn(`Invalid font size value: "${currentSizeWithUnit}", parsed as: ${currentSizeNumber}`);
    }
};

// Function to render KaTeX expressions in the content
function renderKaTeXExpressions() {
  try {
    // Find the content container
    const contentElement = document.querySelector('#content');
    if (!contentElement) {
      console.warn('Content element not found for KaTeX rendering');
      return;
    }

    // Render inline math expressions \(...\)
    const inlineMathElements = contentElement.querySelectorAll('p, div, span, td, th, li');
    inlineMathElements.forEach(element => {
      if (element.innerHTML.includes('\\(') && element.innerHTML.includes('\\)')) {
        element.innerHTML = element.innerHTML.replace(
          /\\\((.*?)\\\)/g,
          (match, latex) => {
            try {
              return katex.renderToString(latex, {
                throwOnError: false,
                displayMode: false
              });
            } catch (error) {
              console.error('KaTeX inline rendering error:', error);
              return match; // Return original if error
            }
          }
        );
      }
    });

    // Render display math expressions \[...\]
    const displayMathElements = contentElement.querySelectorAll('div');
    displayMathElements.forEach(element => {
      if (element.innerHTML.includes('\\[') && element.innerHTML.includes('\\]')) {
        element.innerHTML = element.innerHTML.replace(
          /\\\[(.*?)\\\]/gs, // 's' flag for multiline matching
          (match, latex) => {
            try {
              return `<div>${katex.renderToString(latex.trim(), {
                throwOnError: false,
                displayMode: true
              })}</div>`;
            } catch (error) {
              console.error('KaTeX display rendering error:', error);
              return match; // Return original if error
            }
          }
        );
      }
    });

    console.log('KaTeX expressions rendered successfully');
  } catch (error) {
    console.error('Error in renderKaTeXExpressions:', error);
  }
}


onMounted(async () => {
  // Get initial theme from Go backend
  currentTheme.value = (await GetTheme()) as 'light' | 'dark';

  // Set up a timeout to check if we receive the 'markdown-rendered' event
  // If no event is received within a reasonable time, check if there's actually a file loaded
  let renderTimeout = setTimeout(async () => {
    try {
      const hasFile = await HasCurrentFile();
      if (!hasFile) {
        // Only show the "no file specified" message if there's truly no file
        console.log('No file detected after timeout, showing no file message');
        renderedHTML.value = '<h3>No markdown file specified. Please open a markdown file using File > Open.</h3>';
      } else {
        // There's a file but no render event came - this might indicate an issue
        console.log('File detected but no render event received, keeping loading message');
        // Keep the loading message - user can manually reload if needed
      }
    } catch (error) {
      console.error('Error checking current file status:', error);
      // If we can't check, assume no file and show the appropriate message
      renderedHTML.value = '<h3>No markdown file specified. Please open a markdown file using File > Open.</h3>';
    }
  }, 5000); // Wait 5 seconds for the markdown-rendered event

  // Listen for theme changes initiated from the backend
  EventsOn('theme:changed', (newTheme: string) => {
    if (newTheme && (newTheme === 'light' || newTheme === 'dark')) {
      currentTheme.value = newTheme;
    }
  });

  // Listen for print events from Go backend
  EventsOn('print-content', () => {
      printContent();
  });

  EventsOn('save-as-pdf', (filePath: string) => {
      saveAsPDF(filePath);
  });

  // Listen for techdoc class manipulation events from Go backend
  EventsOn('add-doc-class', (thisClass: string) => {
    addDocClass(thisClass);
  });

  EventsOn('remove-doc-class', (thisClass: string) => {
    removeDocClass(thisClass);
  });

  EventsOn('toggle-doc-class', (thisClass: string) => {
    toggleDocClass(thisClass);
  });

  // Initialize Mermaid.js
  mermaid.initialize({
    startOnLoad: false,
    theme: 'default',
  });

  // Listen for the 'markdown-rendered' event from the Go backend
  EventsOn('markdown-rendered', (data: MarkdownRenderData) => {
    console.log('Received markdownLoaded event. Updating HTML content.');
    clearTimeout(renderTimeout); // Cancel the timeout since we got the event
    renderedHTML.value = DOMPurify.sanitize(data.html || '', {
      USE_PROFILES: { html: true, svg: true, mathMl: true },
      ADD_ATTR: ['target']
    });
    docHTMLTitle.value = DOMPurify.sanitize(data.title || '', {
      ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'code', 'span'],
      ALLOWED_ATTR: ['class']
    });
    docHTMLDate.value = DOMPurify.sanitize(data.date || '', {
      ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'code', 'span'],
      ALLOWED_ATTR: ['class']
    });
    frontmatterHTML.value = data.frontmatter_html || '';
    errorMessage.value = ''; // Clear any previous error message
    // Ensure the DOM is updated before running Mermaid, KaTeX and other updates
    nextTick(() => {
      console.log('Next tick after setting renderedHTML');

      // After the content is set, initialize Mermaid diagrams
      mermaid.run({
        nodes: document.querySelectorAll('.markdown-body .mermaid'),
      });

      // Render KaTeX expressions
      renderKaTeXExpressions();

      document.title = data.title; // Set the document title

      // Call setRootFontSizeNum after document is loaded and DOM is updated
      updateFontSettings();
      setRootFontSizeNum();
    });
  });
  EventsOn('error', (message: string) => {
    console.error('Received error event:', message);
    errorMessage.value = message; // Update the reactive error message.
    renderedHTML.value = '';
  });

  // Listen for an event from the Go backend to show the help dialog
  EventsOn("show-help", (helpTitle: string, helpText: string) => {
      console.log('Received show-help event:', helpTitle, helpText); // Debug log
      showHelpModalDialog(helpTitle, helpText);
  });

  // Listen for settings dialog events from Go backend
  EventsOn("show-settings", () => {
    console.log('Received show-settings event'); // Debug log
    openSettingsDialog();
  });
});

onUnmounted(() => {
  // Clean up the event listener when the component is destroyed
  EventsOff('markdown-rendered');
  EventsOff('error');
  EventsOff('add-doc-class');
  EventsOff('remove-doc-class');
  EventsOff('toggle-doc-class');
  EventsOff('print-content');
  EventsOff('save-as-pdf');
  EventsOff('show-help');
  EventsOff('show-settings');
});

</script>

<style>
/* App.vue component-specific styles */
/* Note: .app-header and .content-area styles have been moved to _app.scss */
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

/* Error message styling - remains here as it's component-specific */
.error-message {
  color: #dc3545;
  font-weight: bold;
  margin-left: 20px;
  padding: 5px 10px;
  background-color: #ffeaea;
  border-radius: 4px;
}

/* Adjust header positioning when frontmatter is visible */
.app-header.with-frontmatter {
  /* Add a small top margin to account for frontmatter section shadow */
  margin-top: 4px;
}

/* Print-specific overrides to ensure proper printing */
@media print {
  .app-container {
    display: block !important;
    height: auto !important;
    overflow: visible !important;
  }
}

</style>
