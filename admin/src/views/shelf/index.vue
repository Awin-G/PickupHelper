<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { getShelfList, createShelf, updateShelf, getShelfOccupancy } from "@/api/shelf";
import type { ShelfItem, ShelfFormRequest } from "@/api/types/parcel";
import type { FormInstance } from "element-plus";

defineOptions({
  name: "ShelfList"
});

const loading = ref(true);
const tableData = ref<ShelfItem[]>([]);
const total = ref(0);
const dialogVisible = ref(false);
const dialogTitle = ref("新增货架");
const formLoading = ref(false);
const editingId = ref<number | null>(null);

const queryParams = reactive({
  station_id: undefined as number | undefined,
  page: 1,
  page_size: 20
});

const formRef = ref<FormInstance>();
const form = reactive<ShelfFormRequest>({
  station_id: 1,
  shelf_code: "",
  row_num: 1,
  col_num: 1,
  max_capacity: 100,
  remark: ""
});

const rules = {
  shelf_code: [{ required: true, message: "请输入货架编号", trigger: "blur" }],
  row_num: [{ required: true, message: "请输入排数", trigger: "blur" }],
  col_num: [{ required: true, message: "请输入列数", trigger: "blur" }],
  max_capacity: [{ required: true, message: "请输入最大容量", trigger: "blur" }]
};

const occupancyData = ref<{
  total_used: number;
  total_max: number;
  shelves: any[];
}>({ total_used: 0, total_max: 0, shelves: [] });

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getShelfList(queryParams);
    tableData.value = res.data.list;
    total.value = res.data.total;
  } catch {
    message("加载失败", { type: "error" });
  } finally {
    loading.value = false;
  }
};

const loadOccupancy = async () => {
  try {
    const res = await getShelfOccupancy({ station_id: queryParams.station_id || 1 });
    occupancyData.value = res.data;
  } catch {
    // 忽略
  }
};

const handleAdd = () => {
  editingId.value = null;
  dialogTitle.value = "新增货架";
  form.shelf_code = "";
  form.row_num = 1;
  form.col_num = 1;
  form.max_capacity = 100;
  form.remark = "";
  dialogVisible.value = true;
};

const handleEdit = (row: ShelfItem) => {
  editingId.value = row.id;
  dialogTitle.value = "编辑货架";
  form.shelf_code = row.shelf_code;
  form.row_num = row.row_num;
  form.col_num = row.col_num;
  form.max_capacity = row.max_capacity;
  form.remark = row.remark || "";
  dialogVisible.value = true;
};

const handleSubmit = async (formEl: FormInstance | undefined) => {
  if (!formEl) return;
  await formEl.validate(async valid => {
    if (!valid) return;
    formLoading.value = true;
    try {
      if (editingId.value) {
        await updateShelf(editingId.value, form);
        message("更新成功", { type: "success" });
      } else {
        await createShelf(form);
        message("新增成功", { type: "success" });
      }
      dialogVisible.value = false;
      loadData();
      loadOccupancy();
    } catch (err: any) {
      message(err?.message || "操作失败", { type: "error" });
    } finally {
      formLoading.value = false;
    }
  });
};

const getOccupancyColor = (rate: number) => {
  if (rate >= 0.9) return "#f56c6c";
  if (rate >= 0.7) return "#e6a23c";
  return "#67c23a";
};

onMounted(() => {
  loadData();
  loadOccupancy();
});
</script>

<template>
  <div class="p-4">
    <el-row :gutter="24">
      <el-col :xs="24" :md="16">
        <el-card shadow="never">
          <template #header>
            <div class="flex justify-between items-center">
              <span class="font-medium text-lg">货架列表</span>
              <el-button type="primary" @click="handleAdd">
                <IconifyIconOffline icon="ep/plus" class="mr-1" />
                新增货架
              </el-button>
            </div>
          </template>

          <el-table v-loading="loading" :data="tableData" stripe border>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="shelf_code" label="货架编号" width="120" />
            <el-table-column prop="row_num" label="排数" width="70" />
            <el-table-column prop="col_num" label="列数" width="70" />
            <el-table-column label="容量" width="150">
              <template #default="{ row }">
                {{ row.current_capacity }} / {{ row.max_capacity }}
              </template>
            </el-table-column>
            <el-table-column prop="occupancy_rate" label="占用率" width="120">
              <template #default="{ row }">
                <el-progress
                  :percentage="Math.round(row.occupancy_rate * 100)"
                  :color="getOccupancyColor(row.occupancy_rate)"
                  :stroke-width="12"
                />
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" show-overflow-tooltip />
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleEdit(row as ShelfItem)">
                  编辑
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="mt-4 flex justify-end">
            <el-pagination
              v-model:current-page="queryParams.page"
              v-model:page-size="queryParams.page_size"
              :page-sizes="[10, 20, 50]"
              :total="total"
              layout="total, sizes, prev, pager, next"
              @current-change="loadData"
              @size-change="loadData"
            />
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="8">
        <el-card shadow="never">
          <template #header>
            <span class="font-medium text-lg">货架占用概览</span>
          </template>
          <div class="text-center mb-4">
            <p class="text-sm text-gray-500">
              总占用: {{ occupancyData.total_used }} / {{ occupancyData.total_max }}
            </p>
            <el-progress
              v-if="occupancyData.total_max > 0"
              :percentage="Math.round((occupancyData.total_used / occupancyData.total_max) * 100)"
              :color="occupancyData.total_used / occupancyData.total_max > 0.9 ? '#f56c6c' : '#409eff'"
              class="mt-2"
            />
          </div>
          <el-table :data="occupancyData.shelves" size="small" stripe>
            <el-table-column prop="shelf_code" label="货架" />
            <el-table-column prop="current_capacity" label="当前" width="60" />
            <el-table-column prop="max_capacity" label="容量" width="60" />
            <el-table-column label="状态" width="70">
              <template #default="{ row }">
                <el-tag
                  :type="row.heat_level >= 4 ? 'danger' : row.heat_level >= 2 ? 'warning' : 'success'"
                  size="small"
                >
                  {{ row.heat_level >= 4 ? "满载" : row.heat_level >= 2 ? "较多" : "空闲" }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="货架编号" prop="shelf_code">
          <el-input v-model="form.shelf_code" placeholder="如 A-01" />
        </el-form-item>
        <el-form-item label="排数" prop="row_num">
          <el-input-number v-model="form.row_num" :min="1" :max="99" />
        </el-form-item>
        <el-form-item label="列数" prop="col_num">
          <el-input-number v-model="form.col_num" :min="1" :max="99" />
        </el-form-item>
        <el-form-item label="最大容量" prop="max_capacity">
          <el-input-number v-model="form.max_capacity" :min="1" :max="9999" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="handleSubmit(formRef)">
          确认
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
