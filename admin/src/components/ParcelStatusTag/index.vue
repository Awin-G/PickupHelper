<template>
  <el-tag :type="tagType" :effect="effect">{{ statusText }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  status: number;
  effect?: "dark" | "light" | "plain";
}

const props = withDefaults(defineProps<Props>(), {
  effect: "light"
});

const statusMap: Record<number, { text: string; type: "primary" | "success" | "warning" | "info" | "danger" }> = {
  1: { text: "待取件", type: "primary" },
  2: { text: "已取件", type: "success" },
  3: { text: "滞留", type: "warning" },
  4: { text: "已退件", type: "info" },
  5: { text: "异常", type: "danger" }
};

const tagType = computed(() => statusMap[props.status]?.type ?? "info");
const statusText = computed(() => statusMap[props.status]?.text ?? "未知");
</script>
