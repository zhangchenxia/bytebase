<template>
  <div class="custom-node-manager">
    <NCard title="自定义审批节点" :bordered="false" class="node-manager-card">
      <template #header-extra>
        <NButton type="primary" @click="showCreateModal = true">
          <template #icon>
            <PlusOutlined />
          </template>
          添加节点
        </NButton>
      </template>

      <NTree
        :data="customNodesTree"
        :default-expanded-keys="defaultExpandedKeys"
        :render="renderTree"
        class="node-tree"
      />

      <NDrawer
        v-model:show="showCreateModal"
        title="创建自定义节点"
        placement="right"
        size="large"
      >
        <CustomNodeForm
          @submit="handleCreateNode"
          @cancel="showCreateModal = false"
        />
      </NDrawer>

      <NDrawer
        v-model:show="showEditModal"
        title="编辑自定义节点"
        placement="right"
        size="large"
      >
        <CustomNodeForm
          :node="selectedNode"
          @submit="handleEditNode"
          @cancel="showEditModal = false"
        />
      </NDrawer>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { NCard, NTree, NButton, NDrawer, message } from "naive-ui";
import { PlusOutlined, EditOutlined, DeleteOutlined } from "@vicons/antd";
import CustomNodeForm from "./CustomNodeForm.vue";
import { useWorkspaceApprovalSettingStore } from "@/store";
import type { CustomApprovalNode } from "./types";

const store = useWorkspaceApprovalSettingStore();

// State
const showCreateModal = ref(false);
const showEditModal = ref(false);
const selectedNode = ref<CustomApprovalNode | null>(null);

// Computed
const customNodesTree = computed(() => {
  // Convert flat list to tree structure
  return [
    {
      key: "custom-nodes",
      title: "自定义审批节点",
      children: store.customApprovalNodes?.map(node => ({
        key: node.id,
        title: node.name,
        description: node.description,
        type: node.type,
        nodeData: node
      })) || []
    }
  ];
});

const defaultExpandedKeys = computed(() => ["custom-nodes"]);

// Methods
const renderTree = (option: any) => {
  if (option.nodeData) {
    const node = option.nodeData as CustomApprovalNode;
    return (
      <div class="tree-node">
        <div class="node-info">
          <div class="node-title">
            {node.name}
            <NTag size="small" :type="getTypeTagType(node.type)">
              {getTypeDisplayName(node.type)}
            </NTag>
          </div>
          {node.description && (
            <div class="node-description">{node.description}</div>
          )}
        </div>
        <div class="node-actions">
          <NButton
            size="small"
            type="primary"
            ghost
            @click="handleEdit(node)"
          >
            <template #icon>
              <EditOutlined />
            </template>
          </NButton>
          <NButton
            size="small"
            type="error"
            ghost
            @click="handleDelete(node)"
          >
            <template #icon>
              <DeleteOutlined />
            </template>
          </NButton>
        </div>
      </div>
    );
  }
  return option.title;
};

const getTypeTagType = (type: string) => {
  const typeMap = {
    "APPROVAL": "primary",
    "CONDITION": "warning",
    "NOTIFICATION": "info",
    "ACTION": "success"
  };
  return typeMap[type as keyof typeof typeMap] || "default";
};

const getTypeDisplayName = (type: string) => {
  const typeMap = {
    "APPROVAL": "审批节点",
    "CONDITION": "条件节点",
    "NOTIFICATION": "通知节点",
    "ACTION": "操作节点"
  };
  return typeMap[type as keyof typeof typeMap] || type;
};

const handleCreateNode = (nodeData: CustomApprovalNode) => {
  store.createCustomApprovalNode(nodeData).then(() => {
    message.success("自定义节点创建成功");
    showCreateModal.value = false;
  }).catch((error) => {
    message.error(`创建失败: ${error.message}`);
  });
};

const handleEdit = (node: CustomApprovalNode) => {
  selectedNode.value = { ...node };
  showEditModal.value = true;
};

const handleEditNode = (nodeData: CustomApprovalNode) => {
  if (!selectedNode.value) return;
  
  store.updateCustomApprovalNode(selectedNode.value.id, nodeData).then(() => {
    message.success("自定义节点更新成功");
    showEditModal.value = false;
    selectedNode.value = null;
  }).catch((error) => {
    message.error(`更新失败: ${error.message}`);
  });
};

const handleDelete = (node: CustomApprovalNode) => {
  store.deleteCustomApprovalNode(node.id).then(() => {
    message.success("自定义节点删除成功");
  }).catch((error) => {
    message.error(`删除失败: ${error.message}`);
  });
};

// Lifecycle
onMounted(() => {
  store.loadCustomApprovalNodes();
});
</script>

<style scoped>
.custom-node-manager {
  width: 100%;
}

.node-manager-card {
  margin-bottom: 24px;
}

.node-tree {
  margin-top: 16px;
}

.tree-node {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.node-info {
  flex: 1;
  margin-right: 16px;
}

.node-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.node-description {
  font-size: 14px;
  color: var(--text-2);
  margin-top: 4px;
}

.node-actions {
  display: flex;
  gap: 8px;
}
</style>