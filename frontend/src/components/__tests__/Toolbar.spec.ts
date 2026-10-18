import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import Toolbar from '../Toolbar.vue';

describe('Toolbar.vue', () => {
  it('emits toggle events from toolbar buttons', async () => {
    const wrapper = mount(Toolbar, {
      props: {
        currentTheme: 'light',
        showFrontmatter: false,
      },
    });

    await wrapper.get('.refresh-btn').trigger('click');
    await wrapper.get('.frontmatter-toggle-btn').trigger('click');
    await wrapper.get('.theme-toggle-btn').trigger('click');

    expect(wrapper.emitted('refresh')).toHaveLength(1);
    expect(wrapper.emitted('toggleFrontmatter')).toHaveLength(1);
    expect(wrapper.emitted('toggleTheme')).toHaveLength(1);
  });

  it('updates the theme button title from the current theme', () => {
    const wrapper = mount(Toolbar, {
      props: {
        currentTheme: 'dark',
        showFrontmatter: true,
      },
    });

    expect(wrapper.get('.refresh-btn').attributes('title')).toBe('Refresh');
    expect(wrapper.get('.theme-toggle-btn').attributes('title')).toBe('Switch to light theme');
    expect(wrapper.get('.frontmatter-toggle-btn').attributes('title')).toBe('Hide Frontmatter');
  });

  it('renders refresh before frontmatter and theme using the shared icon component', () => {
    const wrapper = mount(Toolbar, {
      props: {
        currentTheme: 'light',
        showFrontmatter: false,
      },
    });

    const buttons = wrapper.findAll('.toolbar-right > .toolbar-btn');
    expect(buttons[0].classes()).toContain('refresh-btn');
    expect(buttons[1].classes()).toContain('frontmatter-toggle-btn');
    expect(buttons[2].classes()).toContain('theme-toggle-btn');

    expect(wrapper.find('.refresh-btn .lucide-refresh').exists()).toBe(true);
    expect(wrapper.find('.back-btn').exists()).toBe(false);
  });

  it('renders back button to the left of refresh button when canGoBack is true and emits back event', async () => {
    const wrapper = mount(Toolbar, {
      props: {
        currentTheme: 'light',
        showFrontmatter: false,
        canGoBack: true,
      },
    });

    const buttons = wrapper.findAll('.toolbar-right > .toolbar-btn');
    expect(buttons[0].classes()).toContain('back-btn');
    expect(buttons[1].classes()).toContain('refresh-btn');
    expect(buttons[2].classes()).toContain('frontmatter-toggle-btn');
    expect(buttons[3].classes()).toContain('theme-toggle-btn');

    expect(wrapper.find('.back-btn .lucide-arrow-left').exists()).toBe(true);

    await wrapper.get('.back-btn').trigger('click');
    expect(wrapper.emitted('back')).toHaveLength(1);
  });
});