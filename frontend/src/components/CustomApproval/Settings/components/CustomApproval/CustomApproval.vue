<template>
  <div class="w-full">
    <NTabs type="card" size="large" class="custom-approval-tabs">
      <NTabPane name="rules" tab="审批规则">
        <RulesPanel /> 
      </NTabPane>
      <NTabPane name="custom-nodes" tab="自定义节点">
        <CustomNodeManager /> 
      </NTabPane>
    </NTabs>
  </div>
</template>

<script lang="ts" setup>
import RulesPanel from "./RulesPanel";
import CustomNodeManager from "@/components/CustomApproval/common/CustomNodeManager.vue";
import { NTabs, NTabPane } from "naive-ui";
</script>

<style scoped>
.custom-approval-tabs {
  margin-bottom: 24px;
}
</style>
