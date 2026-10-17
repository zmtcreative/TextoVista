import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import Help from '../Help.vue';

describe('Help.vue', () => {
  it('renders help content when shown via exposed method', async () => {
    const wrapper = mount(Help, {
      props: { show: true },
      global: {
        stubs: {
          teleport: true,
        },
      },
    });

    wrapper.vm.showHelpModalDialog('About TextoVista', '<p>Rendered help text</p>');
    await wrapper.vm.$nextTick();

    expect(wrapper.get('#help-modal-overlay').isVisible()).toBe(true);
    expect(wrapper.get('#help-modal-content').text()).toContain('About TextoVista');
    expect(wrapper.get('#help-modal-text').html()).toContain('Rendered help text');
  });

  it('emits close when the overlay background is clicked', async () => {
    const wrapper = mount(Help, {
      props: { show: true },
      global: {
        stubs: {
          teleport: true,
        },
      },
    });

    await wrapper.get('#help-modal-overlay').trigger('click');

    expect(wrapper.emitted('close')).toHaveLength(1);
  });
});