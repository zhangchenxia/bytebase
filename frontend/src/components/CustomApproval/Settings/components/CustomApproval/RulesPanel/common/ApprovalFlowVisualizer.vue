<template>
  <div class="w-full h-64 bg-gray-50 rounded-lg p-4 flex items-center justify-center">
    <svg ref="svgRef" width="100%" height="100%" viewBox="0 0 800 200">
      <!-- Render approval flow nodes and connections -->
      <g v-for="(node, index) in nodes" :key="index">
        <!-- Node -->
        <rect
          :x="node.x"
          :y="node.y"
          :width="node.width"
          :height="node.height"
          fill="#ffffff"
          stroke="#3b82f6"
          stroke-width="2"
          rx="8"
        />
        <!-- Node title -->
        <text
          :x="node.x + node.width / 2"
          :y="node.y + 30"
          text-anchor="middle"
          font-size="14"
          font-weight="600"
          fill="#1f2937"
        >
          {{ node.title }}
        </text>
        <!-- Node description -->
        <text
          :x="node.x + node.width / 2"
          :y="node.y + 50"
          text-anchor="middle"
          font-size="12"
          fill="#6b7280"
        >
          {{ node.description }}
        </text>
      </g>
      
      <!-- Render connections between nodes -->
      <g v-for="(connection, index) in connections" :key="index">
        <path
          :d="connection.path"
          fill="none"
          stroke="#3b82f6"
          stroke-width="2"
        />
        <!-- Arrow head -->
        <polygon
          :points="connection.arrowPoints"
          fill="#3b82f6"
        />
      </g>
    </svg>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from "vue";
import type { ApprovalFlow } from "@/types/proto-es/v1/issue_service_pb";

interface Node {
  x: number;
  y: number;
  width: number;
  height: number;
  title: string;
  description: string;
}

interface Connection {
  path: string;
  arrowPoints: string;
}

const props = defineProps<{
  flow: ApprovalFlow;
}>();

const svgRef = ref<SVGSVGElement | null>(null);

const nodes = computed<Node[]>(() => {
  if (!props.flow || !props.flow.roles) return [];
  
  const nodeWidth = 150;
  const nodeHeight = 80;
  const spacing = 120;
  const startX = 100;
  const startY = 60;
  
  return props.flow.roles.map((role, index) => ({
    x: startX + index * (nodeWidth + spacing),
    y: startY,
    width: nodeWidth,
    height: nodeHeight,
    title: `Step ${index + 1}`,
    description: role,
  }));
});

const connections = computed<Connection[]>(() => {
  if (nodes.value.length < 2) return [];
  
  return nodes.value.slice(0, -1).map((node, index) => {
    const nextNode = nodes.value[index + 1];
    const startX = node.x + node.width;
    const startY = node.y + node.height / 2;
    const endX = nextNode.x;
    const endY = nextNode.y + nextNode.height / 2;
    
    // Create a curved path
    const controlPointX = (startX + endX) / 2;
    const path = `M ${startX} ${startY} Q ${controlPointX} ${startY} ${endX - 10} ${endY}`;
    
    // Calculate arrow points
    const arrowSize = 8;
    const angle = Math.atan2(endY - startY, endX - startX);
    const arrowPoint1X = endX - 10 - arrowSize * Math.cos(angle - Math.PI / 6);
    const arrowPoint1Y = endY - arrowSize * Math.sin(angle - Math.PI / 6);
    const arrowPoint2X = endX - 10 - arrowSize * Math.cos(angle + Math.PI / 6);
    const arrowPoint2Y = endY - arrowSize * Math.sin(angle + Math.PI / 6);
    
    return {
      path,
      arrowPoints: `${endX - 10},${endY} ${arrowPoint1X},${arrowPoint1Y} ${arrowPoint2X},${arrowPoint2Y}`,
    };
  });
});
</script>

<style scoped>
svg {
  overflow: visible;
}
</style>