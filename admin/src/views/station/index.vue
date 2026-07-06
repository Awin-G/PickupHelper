<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { getStationList, createStation, updateStation } from "@/api/station";
import type { StationItem, StationFormRequest } from "@/api/types/parcel";
import { formatDateTime } from "@/utils/format/datetime";
import type { FormInstance } from "element-plus";

defineOptions({
  name: "StationList"
});

const loading = ref(true);
const tableData = ref<StationItem[]>([]);
const total = ref(0);
const dialogVisible = ref(false);
const dialogTitle = ref("新增驿站");
const formLoading = ref(false);
const editingId = ref<number | null>(null);

const queryParams = reactive({
  keyword: "",
  status: undefined as number | undefined,
  page: 1,
  page_size: 20
});

const formRef = ref<FormInstance>();
const form = reactive<StationFormRequest>({
  name: "",
  address: "",
  latitude: 0,
  longitude: 0,
  business_hours: "09:00-20:00",
  status: 1
});

const rules = {
  name: [{ required: true, message: "请输入驿站名称", trigger: "blur" }],
  address: [{ required: true, message: "请输入地址", trigger: "blur" }],
  latitude: [{ required: true, message: "请输入纬度", trigger: "blur" }],
  longitude: [{ required: true, message: "请输入经度", trigger: "blur" }]
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getStationList(queryParams);
    tableData.value = res.list;
    total.value = res.total;
  } catch {
    message("加载失败", { type: "error" });
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  queryParams.page = 1;
  loadData();
};

const handleReset = () => {
  queryParams.keyword = "";
  queryParams.status = undefined;
  queryParams.page = 1;
  loadData();
};

const handleAdd = () => {
  editingId.value = null;
  dialogTitle.value = "新增驿站";
  form.name = "";
  form.address = "";
  form.latitude = 0;
  form.longitude = 0;
  form.business_hours = "09:00-20:00";
  form.status = 1;
  dialogVisible.value = true;
};

const handleEdit = (row: StationItem) => {
  editingId.value = row.id;
  dialogTitle.value = "编辑驿站";
  form.name = row.name;
  form.address = row.address;
  form.latitude = row.latitude;
  form.longitude = row.longitude;
  form.business_hours = row.business_hours || "09:00-20:00";
  form.status = row.status;
  dialogVisible.value = true;
};

const handleSubmit = async (formEl: FormInstance | undefined) => {
  if (!formEl) return;
  await formEl.validate(async valid => {
    if (!valid) return;
    formLoading.value = true;
    try {
      if (editingId.value) {
        await updateStation(editingId.value, form);
        message("更新成功", { type: "success" });
      } else {
        await createStation(form);
        message("新增成功", { type: "success" });
      }
      dialogVisible.value = false;
      loadData();
    } catch (err: any) {
      message(err?.message || "操作失败", { type: "error" });
    } finally {
      formLoading.value = false;
    }
  });
};

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="p-4">
    <el-card shadow="never">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="font-medium text-lg">驿站管理</span>
          <el-button type="primary" @click="handleAdd">
            <IconifyIconOffline icon="ep/plus" class="mr-1" />
            新增驿站
          </el-button>
        </div>
      </template>

      <el-form :model="queryParams" inline class="mb-4">
        <el-form-item label="搜索">
          <el-input
            v-model="queryParams.keyword"
            placeholder="名称/地址"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="queryParams.status"
            placeholder="全部状态"
            clearable
            style="width: 120px"
          >
            <el-option label="营业中" :value="1" />
            <el-option label="休息中" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" stripe border>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="驿站名称" min-width="150" />
        <el-table-column prop="address" label="地址" min-width="200" show-overflow-tooltip />
        <el-table-column prop="business_hours" label="营业时间" width="120" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? "营业中" : "休息中" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row as StationItem)">
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="550px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="驿站名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入驿站名称" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="form.address" placeholder="请输入地址" />
        </el-form-item>
        <el-form-item label="纬度" prop="latitude">
          <el-input-number v-model="form.latitude" :precision="6" :step="0.001" />
        </el-form-item>
        <el-form-item label="经度" prop="longitude">
          <el-input-number v-model="form.longitude" :precision="6" :step="0.001" />
        </el-form-item>
        <el-form-item label="营业时间">
          <el-input v-model="form.business_hours" placeholder="如 09:00-20:00" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">营业中</el-radio>
            <el-radio :value="0">休息中</el-radio>
          </el-radio-group>
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
