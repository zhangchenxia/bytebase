<template>
  <div class="custom-node-form">
    <NForm 
      ref="formRef" 
      :model="formData" 
      :rules="rules"
      layout="vertical"
    >
      <NFormItem label="节点名称" path="name">
        <NInput 
          v-model:value="formData.name" 
          placeholder="请输入节点名称"
          size="large"
        />
      </NFormItem>

      <NFormItem label="节点描述" path="description">
        <NInput 
          v-model:value="formData.description" 
          placeholder="请输入节点描述（可选）"
          type="textarea"
          :rows="3"
          size="large"
        />
      </NFormItem>

      <NFormItem label="节点类型" path="type">
        <NSelect 
          v-model:value="formData.type" 
          placeholder="请选择节点类型"
          size="large"
          @update:value="handleTypeChange"
        >
          <NSelectOption value="APPROVAL">审批节点</NSelectOption>
          <NSelectOption value="CONDITION">条件节点</NSelectOption>
          <NSelectOption value="NOTIFICATION">通知节点</NSelectOption>
          <NSelectOption value="ACTION">操作节点</NSelectOption>
        </NSelect>
      </NFormItem>

      <!-- 审批节点配置 -->
      <NTabs v-if="formData.type === 'APPROVAL'" size="large">
        <NTabPane name="approvers" tab="审批人设置">
          <NFormItem label="审批人" path="config.approvers">
            <NSelect 
              v-model:value="formData.config.approvers" 
              placeholder="请选择审批人" 
              multiple 
              size="large"
            >
              <NSelectOption v-for="user in users" :key="user.id" :value="user.id">
                {{ user.name }}
              </NSelectOption>
            </NSelect>
          </NFormItem>

          <NFormItem label="审批类型" path="config.approvalType">
            <NRadioGroup v-model:value="formData.config.approvalType" size="large">
              <NRadio value="AND">需所有审批人同意</NRadio>
              <NRadio value="OR">只需任一审批人同意</NRadio>
            </NRadioGroup>
          </NFormItem>
        </NTabPane>
      </NTabs>

      <!-- 条件节点配置 -->
      <NTabs v-if="formData.type === 'CONDITION'" size="large">
        <NTabPane name="condition" tab="条件设置">
          <NFormItem label="条件表达式" path="config.condition">
            <NInput 
              v-model:value="formData.config.condition" 
              placeholder="请输入CEL条件表达式"
              type="textarea"
              :rows="5"
              size="large"
            />
            <div class="form-hint">
              示例: request.resourceType == "DATABASE" && request.size > 100
            </div>
          </NFormItem>
        </NTabPane>
      </NTabs>

      <!-- 通知节点配置 -->
      <NTabs v-if="formData.type === 'NOTIFICATION'" size="large">
        <NTabPane name="notification" tab="通知设置">
          <NFormItem label="通知类型" path="config.notificationType">
            <NSelect 
              v-model:value="formData.config.notificationType" 
              placeholder="请选择通知类型"
              size="large"
            >
              <NSelectOption value="EMAIL">邮件通知</NSelectOption>
              <NSelectOption value="IM">IM通知</NSelectOption>
              <NSelectOption value="BOTH">邮件+IM通知</NSelectOption>
            </NSelect>
          </NFormItem>

          <NFormItem label="通知收件人" path="config.recipients">
            <NSelect 
              v-model:value="formData.config.recipients" 
              placeholder="请选择通知收件人" 
              multiple 
              size="large"
            >
              <NSelectOption v-for="user in users" :key="user.id" :value="user.id">
                {{ user.name }}
              </NSelectOption>
            </NSelect>
          </NFormItem>
        </NTabPane>
      </NTabs>

      <!-- 操作节点配置 -->
      <NTabs v-if="formData.type === 'ACTION'" size="large">
        <NTabPane name="action" tab="操作设置">
          <NFormItem label="操作类型" path="config.actionType">
            <NSelect 
              v-model:value="formData.config.actionType" 
              placeholder="请选择操作类型"
              size="large"
            >
              <NSelectOption value="AUTO_APPROVE">自动审批</NSelectOption>
              <NSelectOption value="AUTO_REJECT">自动拒绝</NSelectOption>
              <NSelectOption value="TRIGGER_WEBHOOK">触发Webhook</NSelectOption>
            </NSelect>
          </NFormItem>

          <NFormItem 
            v-if="formData.config.actionType === 'TRIGGER_WEBHOOK'" 
            label="Webhook URL" 
            path="config.actionConfig.webhookUrl"
          >
            <NInput 
              v-model:value="formData.config.actionConfig.webhookUrl" 
              placeholder="请输入Webhook URL"
              size="large"
            />
          </NFormItem>
        </NTabPane>
      </NTabs>

      <div class="form-actions">
        <NButton @click="handleCancel" size="large">取消</NButton>
        <NButton type="primary" @click="handleSubmit" size="large">
          {{ props.node ? "更新" : "创建" }}
        </NButton>
      </div>
    </NForm>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from "vue";
import type { FormInst } from "naive-ui";
import { useUserStore } from "@/store";
import type { CustomApprovalNode, CustomApprovalNodeConfig } from "./types";

const userStore = useUserStore();

const props = defineProps<{
  node?: CustomApprovalNode;
}>();

const emit = defineEmits<{
  submit: [nodeData: CustomApprovalNode];
  cancel: [];
}>();

const formRef = ref<FormInst | null>(null);

// 初始化表单数据
const formData = reactive<{
  name: string;
  description: string;
  type: string;
  config: CustomApprovalNodeConfig;
}>({
  name: "",
  description: "",
  type: "",
  config: {
    approvers: [],
    approvalType: "AND",
    condition: "",
    notificationType: "EMAIL",
    recipients: [],
    actionType: "AUTO_APPROVE",
    actionConfig: {
      webhookUrl: ""
    }
  }
});

// 表单验证规则
const rules = {
  name: [
    { required: true, message: "请输入节点名称", trigger: "blur" }
  ],
  type: [
    { required: true, message: "请选择节点类型", trigger: "change" }
  ]
};

// 加载用户列表
const users = computed(() => userStore.allUsers);

// 方法
const handleTypeChange = (newType: string) => {
  // 切换节点类型时重置配置
  formData.config = {
    approvers: [],
    approvalType: "AND",
    condition: "",
    notificationType: "EMAIL",
    recipients: [],
    actionType: "AUTO_APPROVE",
    actionConfig: {
      webhookUrl: ""
    }
  };
};

const handleSubmit = () => {
  if (!formRef.value) return;
  
  formRef.value.validate().then(() => {
    const nodeData: CustomApprovalNode = {
      id: props.node?.id || `custom-node-${Date.now()}`,
      name: formData.name,
      description: formData.description,
      type: formData.type as any,
      config: formData.config,
      createdAt: props.node?.createdAt || Date.now() / 1000,
      updatedAt: Date.now() / 1000
    };
    
    emit("submit", nodeData);
  }).catch((errors) => {
    console.error("表单验证失败:", errors);
  });
};

const handleCancel = () => {
  emit("cancel");
};

// 生命周期
onMounted(() => {
  // 加载所有用户
  userStore.loadAllUsers();
  
  // 如果有节点数据，初始化表单
  if (props.node) {
    formData.name = props.node.name;
    formData.description = props.node.description || "";
    formData.type = props.node.type;
    formData.config = { ...props.node.config } || {
      approvers: [],
      approvalType: "AND",
      condition: "",
      notificationType: "EMAIL",
      recipients: [],
      actionType: "AUTO_APPROVE",
      actionConfig: {
        webhookUrl: ""
      }
    };
  }
});
</script>

<style scoped>
.custom-node-form {
  padding: 24px;
}

.form-hint {
  font-size: 12px;
  color: var(--text-2);
  margin-top: 4px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 16px;
  margin-top: 32px;
}
</style>