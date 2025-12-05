import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import CustomNodeForm from '../common/CustomNodeForm.vue';
import { NForm, NFormItem, NInput, NSelect, NTextarea, NButton } from 'naive-ui';

// Mock the translate function
vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}));

describe('CustomNodeForm', () => {
  let wrapper: any;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();

    // Mount the component
    wrapper = mount(CustomNodeForm, {
      global: {
        components: {
          NForm,
          NFormItem,
          NInput,
          NSelect,
          NTextarea,
          NButton,
        },
      },
    });
  });

  it('should render correctly', () => {
    expect(wrapper.find('.custom-node-form').exists()).toBe(true);
    expect(wrapper.findComponent(NForm).exists()).toBe(true);
    expect(wrapper.findComponent(NFormItem).exists()).toBe(true);
    expect(wrapper.findComponent(NInput).exists()).toBe(true);
    expect(wrapper.findComponent(NSelect).exists()).toBe(true);
    expect(wrapper.findComponent(NTextarea).exists()).toBe(true);
    expect(wrapper.findComponent(NButton).exists()).toBe(true);
  });

  it('should render edit mode correctly when node prop is provided', () => {
    // Test node data
    const testNode = {
      id: '1',
      name: 'Test Node',
      description: 'Test Description',
      type: 'APPROVAL',
      config: {
        approvers: ['user1', 'user2'],
      },
    };

    // Re-mount the component with the node prop
    wrapper = mount(CustomNodeForm, {
      props: {
        node: testNode,
      },
      global: {
        components: {
          NForm,
          NFormItem,
          NInput,
          NSelect,
          NTextarea,
          NButton,
        },
      },
    });

    // Check if the form is populated with the correct data
    const nameInput = wrapper.findComponent(NInput).element as HTMLInputElement;
    expect(nameInput.value).toBe('Test Node');

    const typeSelect = wrapper.findComponent(NSelect).element as HTMLSelectElement;
    expect(typeSelect.value).toBe('APPROVAL');

    const descriptionTextarea = wrapper.findComponent(NTextarea).element as HTMLTextAreaElement;
    expect(descriptionTextarea.value).toBe('Test Description');
  });

  it('should validate required fields', async () => {
    // Submit the form without filling any fields
    const submitButton = wrapper.findComponent(NButton);
    await submitButton.trigger('click');

    // Check if validation errors are displayed
    expect(wrapper.find('.n-form-item__feedback--error').exists()).toBe(true);
  });

  it('should emit submit event with correct data when form is valid', async () => {
    // Fill in the form fields
    const nameInput = wrapper.findComponent(NInput);
    await nameInput.setValue('Test Node');

    const typeSelect = wrapper.findComponent(NSelect);
    await typeSelect.setValue('APPROVAL');

    const descriptionTextarea = wrapper.findComponent(NTextarea);
    await descriptionTextarea.setValue('Test Description');

    // Submit the form
    const submitButton = wrapper.findComponent(NButton);
    await submitButton.trigger('click');

    // Check if submit event was emitted with the correct data
    expect(wrapper.emitted('submit')).toBeDefined();
    expect(wrapper.emitted('submit')?.[0][0]).toEqual({
      name: 'Test Node',
      description: 'Test Description',
      type: 'APPROVAL',
      config: {},
    });
  });

  it('should emit cancel event when cancel button is clicked', async () => {
    // Click the cancel button
    const cancelButton = wrapper.findAllComponents(NButton)[1];
    await cancelButton.trigger('click');

    // Check if cancel event was emitted
    expect(wrapper.emitted('cancel')).toBeDefined();
  });

  it('should render different config fields based on node type', async () => {
    // Select CONDITION type
    const typeSelect = wrapper.findComponent(NSelect);
    await typeSelect.setValue('CONDITION');

    // Check if condition field is rendered
    expect(wrapper.find('.condition-field').exists()).toBe(true);

    // Select NOTIFICATION type
    await typeSelect.setValue('NOTIFICATION');

    // Check if notification fields are rendered
    expect(wrapper.find('.notification-field').exists()).toBe(true);

    // Select ACTION type
    await typeSelect.setValue('ACTION');

    // Check if action field is rendered
    expect(wrapper.find('.action-field').exists()).toBe(true);
  });
});