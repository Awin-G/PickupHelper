<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { getParcelList } from "@/api/parcel";
import type { ParcelItem, ParcelListParams } from "@/api/types/parcel";
import ParcelStatusTag from "@/components/ParcelStatusTag/index.vue";
import CourierCompanySelect from "@/components/CourierCompanySelect/index.vue";
import { maskPhone } from "@/utils/format/phone";
import { formatDateTime } from "@/utils/format/datetime";

defineOptions({
  name: "ParcelList"
});

const loading = ref(true);
const tableData = ref<ParcelItem[]>([]);
const total = ref(0);

const queryParams = reactive<ParcelListParams>({
  tracking_no: "",
  receiver_phone: "",
  status: undefined,
  courier_company: "",
  shelf_code: "",
  storage_start: "",
  storage_end: "",
  page: 1,
  page_size: 20
});

const statusOptions = [
  { label: "待取件", value: 1 },
  { label: "已取件", value: 2 },
  { label: "滞留", value: 3 },
  { label: "已退件", value: 4 },
  { label: "异常", value: 5 }
];

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getParcelList(queryParams);
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
  queryParams.tracking_no = "";
  queryParams.receiver_phone = "";
  queryParams.status = undefined;
  queryParams.courier_company = "";
  queryParams.shelf_code = "";
  queryParams.storage_start = "";
  queryParams.storage_end = "";
  queryParams.page = 1;
  loadData();
};

const handlePageChange = (page: number) => {
  queryParams.page = page;
  loadData();
};

const handleSizeChange = (size: number) => {
  queryParams.page_size = size;
  queryParams.page = 1;
  loadData();
};

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="p-4">
    <el-card shadow="never">
      <template #header>
        <span class="font-medium text-lg">包裹列表</span>
      </template>

      <el-form :model="queryParams" inline class="mb-4">
        <el-form-item label="快递单号">
          <el-input
            v-model="queryParams.tracking_no"
            placeholder="请输入快递单号"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="收件人手机">
          <el-input
            v-model="queryParams.receiver_phone"
            placeholder="请输入收件人手机号"
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
            style="width: 140px"
          >
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="快递公司">
          <CourierCompanySelect
            v-model="queryParams.courier_company"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="货架编号">
          <el-input
            v-model="queryParams.shelf_code"
            placeholder="请输入货架编号"
            clearable
            style="width: 140px"
          />
        </el-form-item>
        <el-form-item label="入库时间">
          <el-date-picker
            v-model="queryParams.storage_start"
            type="datetime"
            placeholder="开始时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ss+08:00"
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item>
          <el-date-picker
            v-model="queryParams.storage_end"
            type="datetime"
            placeholder="结束时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ss+08:00"
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="tracking_no" label="快递单号" min-width="180" show-overflow-tooltip />
        <el-table-column prop="courier_company" label="快递公司" width="120" />
        <el-table-column prop="receiver_phone" label="收件人手机" width="130">
          <template #default="{ row }">
            {{ maskPhone(row.receiver_phone) }}
          </template>
        </el-table-column>
        <el-table-column prop="receiver_name" label="收件人" width="100" />
        <el-table-column prop="shelf_code" label="货架" width="80" />
        <el-table-column prop="pickup_code" label="取件码" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.pickup_code" effect="dark" type="success">
              {{ row.pickup_code }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <ParcelStatusTag :status="row.status" />
          </template>
        </el-table-column>
        <el-table-column prop="storage_time" label="入库时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.storage_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="pickup_time" label="取件时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.pickup_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="notify_count" label="催取次数" width="90" />
      </el-table>

      <div class="mt-4 flex justify-end">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>
