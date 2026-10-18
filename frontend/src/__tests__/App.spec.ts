import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import App from '../App.vue';
import type { MarkdownRenderData } from '../types/markdown';
import { createConfigFixture } from '../test-utils/config-fixtures';

const runtimeHandlers = new Map<string, (...args: any[]) => void>();

async function waitForRuntimeHandlers(): Promise<void> {
  await vi.waitFor(() => {
    expect(runtimeHandlers.has('markdown-rendered')).toBe(true);
    expect(runtimeHandlers.has('theme:changed')).toBe(true);
    expect(runtimeHandlers.has('show-help')).toBe(true);
    expect(runtimeHandlers.has('show-settings')).toBe(true);
  });
}

const runtimeMocks = vi.hoisted(() => ({
  EventsOn: vi.fn((eventName: string, callback: (...args: any[]) => void) => {
    runtimeHandlers.set(eventName, callback);
    return vi.fn();
  }),
  EventsOff: vi.fn((...eventNames: string[]) => {
    for (const eventName of eventNames) {
      runtimeHandlers.delete(eventName);
    }
  }),
}));

const appApiMocks = vi.hoisted(() => ({
  GetTheme: vi.fn(),
  SetTheme: vi.fn(),
  GetSettings: vi.fn(),
  GetCurrentFont: vi.fn(),
  GetCurrentMonospaceFont: vi.fn(),
  HasCurrentFile: vi.fn(),
  ReloadCurrentDocument: vi.fn(),
  GetAlertCalloutStyles: vi.fn(),
  GetAvailableFonts: vi.fn(),
  GetAvailableMonospaceFonts: vi.fn(),
  GetAdvancedFontDetectionStatus: vi.fn(),
  SetAdvancedFontDetection: vi.fn(),
  SaveSettings: vi.fn(),
  SaveSettingsSessionOnly: vi.fn(),
}));

const mermaidMocks = vi.hoisted(() => ({
  initialize: vi.fn(),
  run: vi.fn(),
}));

const katexMocks = vi.hoisted(() => ({
  renderToString: vi.fn((latex: string) => `<span class="katex">${latex}</span>`),
}));

vi.mock('../../wailsjs/runtime/runtime', () => runtimeMocks);
vi.mock('../../wailsjs/go/main/App', () => appApiMocks);
vi.mock('mermaid', () => ({ default: mermaidMocks }));
vi.mock('katex', () => ({ default: katexMocks }));

describe('App.vue', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    runtimeHandlers.clear();

    appApiMocks.GetTheme.mockResolvedValue('light');
    appApiMocks.SetTheme.mockResolvedValue(undefined);
    appApiMocks.GetSettings.mockResolvedValue(createConfigFixture());
    appApiMocks.GetCurrentFont.mockResolvedValue({ fontFamily: 'Verdana', fontSize: 16 });
    appApiMocks.GetCurrentMonospaceFont.mockResolvedValue({ fontFamily: 'Consolas', fontSize: 14 });
    appApiMocks.HasCurrentFile.mockResolvedValue(false);
    appApiMocks.ReloadCurrentDocument.mockResolvedValue(undefined);
    appApiMocks.GetAlertCalloutStyles.mockResolvedValue({ GFMPlus: 'GitHub Flavored Markdown Plus' });
    appApiMocks.GetAvailableFonts.mockResolvedValue(['Verdana', 'Tahoma']);
    appApiMocks.GetAvailableMonospaceFonts.mockResolvedValue(['Consolas', 'Courier New']);
    appApiMocks.GetAdvancedFontDetectionStatus.mockResolvedValue(false);
    appApiMocks.SetAdvancedFontDetection.mockResolvedValue(undefined);
    appApiMocks.SaveSettings.mockResolvedValue(undefined);
    appApiMocks.SaveSettingsSessionOnly.mockResolvedValue(undefined);

    mermaidMocks.initialize.mockClear();
    mermaidMocks.run.mockClear();
    katexMocks.renderToString.mockClear();
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('falls back to the no-file message when no markdown file is loaded', async () => {
    const wrapper = mount(App, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: true,
        },
      },
    });

    await vi.runAllTimersAsync();
    await vi.waitFor(() => expect(appApiMocks.HasCurrentFile).toHaveBeenCalled());

    expect(wrapper.get('#content').text()).toContain('No markdown file specified');
  });

  it('updates rendered content, title, frontmatter, and diagram processing on markdown-rendered', async () => {
    const wrapper = mount(App, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: true,
        },
      },
    });
    await waitForRuntimeHandlers();
    const markdownRendered = runtimeHandlers.get('markdown-rendered');

    expect(markdownRendered).toBeTypeOf('function');

    const payload: MarkdownRenderData = {
      html: '<p>Inline math \\(x+y\\)</p><div>\\[x^2\\]</div><pre class="mermaid">graph TD;A-->B;</pre>',
      title: 'Rendered Title',
      date: '2026-06-03',
      frontmatter_html: '<div class="frontmatter-container"><span class="fm-key">title</span></div>',
    };

    markdownRendered!(payload);
    await wrapper.vm.$nextTick();
    await vi.waitFor(() => expect(mermaidMocks.run).toHaveBeenCalled());

    expect(document.title).toBe('Rendered Title');
    expect(wrapper.get('.document-title').html()).toContain('Rendered Title');
    await wrapper.get('.frontmatter-toggle-btn').trigger('click');
    expect(wrapper.find('.frontmatter-section').exists()).toBe(true);
  });

  it('sanitizes malicious markup and event handlers in markdown-rendered payload', async () => {
    const wrapper = mount(App, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: true,
        },
      },
    });
    await waitForRuntimeHandlers();
    const markdownRendered = runtimeHandlers.get('markdown-rendered');

    const maliciousPayload: MarkdownRenderData = {
      html: '<p>Safe paragraph</p><script>alert("pwned")</script><img src="x" onerror="alert(1)"><a href="javascript:alert(2)">bad link</a>',
      title: '<script>alert(3)</script><b>Safe Title</b>',
      date: '<img src=x onerror=alert(4)>2026-10-18',
      frontmatter_html: '',
    };

    markdownRendered!(maliciousPayload);
    await wrapper.vm.$nextTick();

    const contentHtml = wrapper.get('#content').html();
    expect(contentHtml).toContain('Safe paragraph');
    expect(contentHtml).not.toContain('<script>');
    expect(contentHtml).not.toContain('onerror');
    expect(contentHtml).not.toContain('javascript:alert');

    const titleHtml = wrapper.get('.document-title').html();
    expect(titleHtml).not.toContain('<script>');
    expect(titleHtml).toContain('<b>Safe Title</b>');

    const dateHtml = wrapper.get('.document-dates').html();
    expect(dateHtml).not.toContain('<img');
    expect(dateHtml).not.toContain('onerror');
    expect(dateHtml).toContain('2026-10-18');
  });

  it('responds to theme, help, settings, and error events from the runtime', async () => {
    const wrapper = mount(App, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: true,
        },
      },
    });
    await waitForRuntimeHandlers();

    runtimeHandlers.get('theme:changed')!('dark');
    await wrapper.vm.$nextTick();

    expect(document.documentElement.className).toBe('dark');
    expect(document.body.classList.contains('dark')).toBe(true);

    runtimeHandlers.get('show-help')!('About', '<p>Help body</p>');
    await wrapper.vm.$nextTick();
    expect(wrapper.get('#help-modal-overlay').isVisible()).toBe(true);
    expect(wrapper.get('#help-modal-text').html()).toContain('Help body');

    runtimeHandlers.get('show-settings')!();
    await wrapper.vm.$nextTick();
    expect(wrapper.get('#settings-overlay').isVisible()).toBe(true);

    runtimeHandlers.get('error')!('<img src=x onerror="alert(1)">Unable to load file');
    await wrapper.vm.$nextTick();
    expect(wrapper.get('.error-message').text()).toContain('Unable to load file');
    expect(wrapper.get('#content').text()).toContain('An error occurred');
    expect(wrapper.get('#content').text()).toContain('<img src=x onerror="alert(1)">Unable to load file');
    expect(wrapper.get('#content').html()).not.toContain('<img src=x onerror="alert(1)">');
    expect(wrapper.get('#content').find('img').exists()).toBe(false);
  });

  it('toggles theme from the toolbar and delegates persistence to the backend', async () => {
    const wrapper = mount(App, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: true,
        },
      },
    });
    await waitForRuntimeHandlers();

    await wrapper.get('.theme-toggle-btn').trigger('click');
    await vi.waitFor(() => expect(appApiMocks.SetTheme).toHaveBeenCalledWith('dark'));

    expect(document.documentElement.className).toBe('dark');
  });

  it('manually reloads the current document from the toolbar refresh button', async () => {
    const wrapper = mount(App, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: true,
        },
      },
    });
    await waitForRuntimeHandlers();

    await wrapper.get('.refresh-btn').trigger('click');

    await vi.waitFor(() => expect(appApiMocks.ReloadCurrentDocument).toHaveBeenCalledTimes(1));
  });

  it('shows an error when manual refresh fails', async () => {
    appApiMocks.ReloadCurrentDocument.mockRejectedValueOnce(new Error('reload failed'));

    const wrapper = mount(App, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: true,
        },
      },
    });
    await waitForRuntimeHandlers();

    await wrapper.get('.refresh-btn').trigger('click');

    await vi.waitFor(() => {
      expect(wrapper.get('.error-message').text()).toContain('reload failed');
    });
  });

  it('renders application layout with proper container hierarchy for scrolling', async () => {
    const wrapper = mount(App, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: true,
        },
      },
    });
    await waitForRuntimeHandlers();

    const container = wrapper.find('.app-container');
    expect(container.exists()).toBe(true);
    expect(container.find('.app-toolbar').exists()).toBe(true);
    expect(container.find('header.app-header').exists()).toBe(true);
    expect(container.find('main.content-area').exists()).toBe(true);
  });
});