import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import CustomNodeManager from '../common/CustomNodeManager.vue';
import CustomNodeForm from '../common/CustomNodeForm.vue';
import { NButton, NModal, NTable, NPopconfirm } from 'naive-ui';
import { useCustomApprovalContext } from '@/components/CustomApproval/Settings/components/CustomApproval/context';

// Mock the context
vi.mock('@/components/CustomApproval/Settings/components/CustomApproval/context', () => ({
  useCustomApprovalContext: vi.fn(() => ({
    nodes: [],
    addNode: vi.fn(),
    updateNode: vi.fn(),
    deleteNode: vi.fn(),
  })),
}));

describe('CustomNodeManager', () => {
  let wrapper: any;
  let context: any;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();

    // Create a fresh context mock before each test
    context = {
      nodes: [],
      addNode: vi.fn(),
      updateNode: vi.fn(),
      deleteNode: vi.fn(),
    };
    (useCustomApprovalContext as vi.Mock).mockReturnValue(context);

    // Mount the component
    wrapper = mount(CustomNodeManager, {
      global: {
        components: {
          NButton,
          NModal,
          NTable,
          NPopconfirm,
          CustomNodeForm,
        },
      },
    });
  });

  it('should render correctly', () => {
    expect(wrapper.find('.custom-node-manager').exists()).toBe(true);
    expect(wrapper.find('h2').text()).toContain('自定义节点');
    expect(wrapper.findComponent(NButton).exists()).toBe(true);
    expect(wrapper.findComponent(NTable).exists()).toBe(true);
  });

  it('should open create modal when click add button', async () => {
    // Click the add button
    await wrapper.findComponent(NButton).trigger('click');

    // Check if the modal is visible
    expect(wrapper.findComponent(NModal).props('show')).toBe(true);
    expect(wrapper.findComponent(CustomNodeForm).exists()).toBe(true);
  });

  it('should call addNode when create form is submitted', async () => {
    // Open the create modal
    await wrapper.findComponent(NButton).trigger('click');

    // Get the form component
    const form = wrapper.findComponent(CustomNodeForm);

    // Submit the form with test data
    const testNode = {
      id: '1',
      name: 'Test Node',
      description: 'Test Description',
      type: 'APPROVAL',
      config: {},
    };
    await form.vm.$emit('submit', testNode);

    // Check if addNode was called with the correct data
    expect(context.addNode).toHaveBeenCalledWith(testNode);
    expect(wrapper.findComponent(NModal).props('show')).toBe(false);
  });

  it('should render nodes in table', () => {
    // Add some test nodes to the context
    context.nodes = [
      {
        id: '1',
        name: 'Test Node 1',
        description: 'Test Description 1',
        type: 'APPROVAL',
        config: {},
      },
      {
        id: '2',
        name: 'Test Node 2',
        description: 'Test Description 2',
        type: 'CONDITION',
        config: {},
      },
    ];

    // Re-mount the component with the new nodes
    wrapper = mount(CustomNodeManager, {
      global: {
        components: {
          NButton,
          NModal,
          NTable,
          NPopconfirm,
          CustomNodeForm,
        },
      },
    });

    // Check if the nodes are rendered in the table
    const rows = wrapper.findComponent(NTable).findAll('tbody tr');
    expect(rows.length).toBe(2);
    expect(rows[0].find('td').text()).toContain('Test Node 1');
    expect(rows[1].find('td').text()).toContain('Test Node 2');
  });

  it('should open edit modal when click edit button', async () => {
    // Add a test node to the context
    context.nodes = [
      {
        id: '1',
        name: 'Test Node',
        description: 'Test Description',
        type: 'APPROVAL',
        config: {},
      },
    ];

    // Re-mount the component with the test node
    wrapper = mount(CustomNodeManager, {
      global: {
        components: {
          NButton,
          NModal,
          NTable,
          NPopconfirm,
          CustomNodeForm,
        },
      },
    });

    // Click the edit button
    const editButton = wrapper.findComponent(NTable).findAll('tbody tr')[0].findAll('button')[0];
    await editButton.trigger('click');

    // Check if the modal is visible and contains the correct data
    expect(wrapper.findComponent(NModal).props('show')).toBe(true);
    const form = wrapper.findComponent(CustomNodeForm);
    expect(form.props('node')).toEqual(context.nodes[0]);
  });

  it('should call updateNode when edit form is submitted', async () => {
    // Add a test node to the context
    context.nodes = [
      {
        id: '1',
        name: 'Test Node',
        description: 'Test Description',
        type: 'APPROVAL',
        config: {},
      },
    ];

    // Re-mount the component with the test node
    wrapper = mount(CustomNodeManager, {
      global: {
        components: {
          NButton,
          NModal,
          NTable,
          NPopconfirm,
          CustomNodeForm,
        },
      },
    });

    // Open the edit modal
    const editButton = wrapper.findComponent(NTable).findAll('tbody tr')[0].findAll('button')[0];
    await editButton.trigger('click');

    // Get the form component
    const form = wrapper.findComponent(CustomNodeForm);

    // Submit the form with updated data
    const updatedNode = {
      id: '1',
      name: 'Updated Test Node',
      description: 'Updated Test Description',
      type: 'APPROVAL',
      config: {},
    };
    await form.vm.$emit('submit', updatedNode);

    // Check if updateNode was called with the correct data
    expect(context.updateNode).toHaveBeenCalledWith(updatedNode);
    expect(wrapper.findComponent(NModal).props('show')).toBe(false);
  });

  it('should call deleteNode when delete button is confirmed', async () => {
    // Add a test node to the context
    context.nodes = [
      {
        id: '1',
        name: 'Test Node',
        description: 'Test Description',
        type: 'APPROVAL',
        config: {},
      },
    ];

    // Re-mount the component with the test node
    wrapper = mount(CustomNodeManager, {
      global: {
        components: {
          NButton,
          NModal,
          NTable,
          NPopconfirm,
          CustomNodeForm,
        },
      },
    });

    // Click the delete button and confirm
    const deleteButton = wrapper.findComponent(NTable).findAll('tbody tr')[0].findAll('button')[1];
    await deleteButton.trigger('click');

    // Find the confirm button in the popconfirm
    const popconfirm = wrapper.findComponent(NPopconfirm);
    const confirmButton = popconfirm.find('button');
    await confirmButton.trigger('click');

    // Check if deleteNode was called with the correct id
    expect(context.deleteNode).toHaveBeenCalledWith('1');
  });
});