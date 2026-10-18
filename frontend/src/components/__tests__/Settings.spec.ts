import { mount } from '@vue/test-utils';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import Settings from '../Settings.vue';
import { createConfigFixture } from '../../test-utils/config-fixtures';

const appApiMocks = vi.hoisted(() => ({
  GetSettings: vi.fn(),
  GetAlertCalloutStyles: vi.fn(),
  SaveSettings: vi.fn(),
  SaveSettingsSessionOnly: vi.fn(),
  GetAvailableFonts: vi.fn(),
  GetAvailableMonospaceFonts: vi.fn(),
  GetAdvancedFontDetectionStatus: vi.fn(),
  SetAdvancedFontDetection: vi.fn(),
}));

vi.mock('../../../wailsjs/go/main/App', () => appApiMocks);

describe('Settings.vue', () => {
  beforeEach(() => {
    const config = createConfigFixture();
    appApiMocks.GetSettings.mockResolvedValue(config);
    appApiMocks.GetAlertCalloutStyles.mockResolvedValue({ GFMPlus: 'GitHub Flavored Markdown Plus' });
    appApiMocks.GetAvailableFonts.mockResolvedValue(['Verdana', 'Tahoma']);
    appApiMocks.GetAvailableMonospaceFonts.mockResolvedValue(['Consolas', 'Courier New']);
    appApiMocks.GetAdvancedFontDetectionStatus.mockResolvedValue(false);
    appApiMocks.SaveSettings.mockResolvedValue(undefined);
    appApiMocks.SaveSettingsSessionOnly.mockResolvedValue(undefined);
    appApiMocks.SetAdvancedFontDetection.mockResolvedValue(undefined);
  });

  it('loads settings data when shown', async () => {
    const wrapper = mount(Settings, {
      props: { show: true },
      global: {
        stubs: {
          teleport: true,
        },
      },
    });

    await vi.waitFor(() => {
      expect(appApiMocks.GetSettings).toHaveBeenCalled();
      expect(appApiMocks.GetAlertCalloutStyles).toHaveBeenCalled();
      expect(appApiMocks.GetAvailableFonts).toHaveBeenCalled();
      expect(appApiMocks.GetAvailableMonospaceFonts).toHaveBeenCalled();
    });

    expect(wrapper.get('#settings-overlay').isVisible()).toBe(true);
    expect((wrapper.get('select.setting-select').element as HTMLSelectElement).value).toBe('Verdana');
  });

  it('saves to disk with the loaded config and emits close/saved', async () => {
    const wrapper = mount(Settings, {
      props: { show: true },
      global: {
        stubs: {
          teleport: true,
        },
      },
    });

    await vi.waitFor(() => expect(appApiMocks.GetSettings).toHaveBeenCalled());
    await wrapper.get('#application-auto-refresh').setValue(false);
    await wrapper.get('form.settings-form').trigger('submit');

    await vi.waitFor(() => expect(appApiMocks.SaveSettings).toHaveBeenCalledTimes(1));

    expect(appApiMocks.SaveSettings.mock.calls[0][0].application.use_auto_refresh).toBe(false);
    expect(appApiMocks.SaveSettings.mock.calls[0][0].application.font_family).toBe('Verdana');
    expect(wrapper.emitted('saved')).toHaveLength(1);
    expect(wrapper.emitted('close')).toHaveLength(1);
  });

  it('applies settings for the current session only', async () => {
    const wrapper = mount(Settings, {
      props: { show: true },
      global: {
        stubs: {
          teleport: true,
        },
      },
    });

    await vi.waitFor(() => expect(appApiMocks.GetSettings).toHaveBeenCalled());
    await wrapper.findAll('button').find((button) => button.text() === 'Apply for Session')!.trigger('click');

    await vi.waitFor(() => expect(appApiMocks.SaveSettingsSessionOnly).toHaveBeenCalledTimes(1));
    expect(wrapper.emitted('saved')).toHaveLength(1);
    expect(wrapper.emitted('close')).toHaveLength(1);
  });

  it('resets font selections back to defaults', async () => {
    const wrapper = mount(Settings, {
      props: { show: true },
      global: {
        stubs: {
          teleport: true,
        },
      },
    });

    await vi.waitFor(() => expect(appApiMocks.GetSettings).toHaveBeenCalled());

    const inlineHtmlCheckbox = wrapper.get('input[type="checkbox"]');
    await inlineHtmlCheckbox.setValue(false);
    expect((inlineHtmlCheckbox.element as HTMLInputElement).checked).toBe(false);

    await wrapper.findAll('button').find((button) => button.text() === 'Reset to Defaults')!.trigger('click');

    expect((wrapper.get('input[type="checkbox"]').element as HTMLInputElement).checked).toBe(true);
  });

  it('removes Sanitize HTML checkbox and syncs use_sanitize_html with use_inline_html on save', async () => {
    const wrapper = mount(Settings, {
      props: { show: true },
      global: {
        stubs: {
          teleport: true,
        },
      },
    });

    await vi.waitFor(() => expect(appApiMocks.GetSettings).toHaveBeenCalled());

    // Sanitize HTML checkbox should not exist anywhere in the modal
    expect(wrapper.text()).not.toContain('Sanitize HTML');
    expect(wrapper.text()).toContain('Allow Inline HTML');
    expect(wrapper.text()).toContain('automatically sanitized for safety');

    const inlineHtmlCheckbox = wrapper.get('input[type="checkbox"]');
    await inlineHtmlCheckbox.setValue(false);

    await wrapper.get('form.settings-form').trigger('submit');
    await vi.waitFor(() => expect(appApiMocks.SaveSettings).toHaveBeenCalled());

    const savedConfig = appApiMocks.SaveSettings.mock.calls[0][0];
    expect(savedConfig.application.use_inline_html).toBe(false);
    expect(savedConfig.application.use_sanitize_html).toBe(false);
  });
});