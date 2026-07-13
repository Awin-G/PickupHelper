<template>
  <el-select
    v-model="selectedValue"
    :placeholder="placeholder"
    :disabled="disabled"
    filterable
    clearable
    @change="handleChange"
  >
    <el-option
      v-for="company in companies"
      :key="company"
      :label="company"
      :value="company"
    />
  </el-select>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  modelValue?: string;
  placeholder?: string;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: "选择快递公司",
  disabled: false
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
  change: [value: string];
}>();

const selectedValue = computed({
  get: () => props.modelValue,
  set: val => emit("update:modelValue", val)
});

const companies = [
  "顺丰速运",
  "中通快递",
  "圆通速递",
  "韵达快递",
  "申通快递",
  "百世快递",
  "极兔速递",
  "邮政快递",
  "京东物流",
  "EMS",
  "其他"
];

const handleChange = (val: string) => {
  emit("change", val);
};
</script>
