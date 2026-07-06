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
      v-for="shelf in shelves"
      :key="shelf.id"
      :label="`${shelf.shelf_code} (${shelf.current_capacity}/${shelf.max_capacity})`"
      :value="shelf.shelf_code"
      :disabled="shelf.current_capacity >= shelf.max_capacity"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { getShelfList } from "@/api/shelf";
import type { ShelfItem } from "@/api/types/parcel";

interface Props {
  modelValue?: string;
  stationId: number;
  placeholder?: string;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: "选择货架",
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

const shelves = ref<ShelfItem[]>([]);

const loadShelves = async () => {
  if (!props.stationId) return;
  try {
    const res = await getShelfList({ station_id: props.stationId });
    shelves.value = res.list;
  } catch {
    shelves.value = [];
  }
};

const handleChange = (val: string) => {
  emit("change", val);
};

watch(
  () => props.stationId,
  () => {
    loadShelves();
  }
);

onMounted(() => {
  loadShelves();
});
</script>
