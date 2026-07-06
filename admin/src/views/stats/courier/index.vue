<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { getCourierCheck } from "@/api/stats";
import { formatMoney } from "@/utils/format/money";

defineOptions({
  name: "StatsCourier"
});

const loading = ref(true);
const tableData = ref<any[]>([]);

const queryParams = reactive({
  station_id: undefined as number | undefined,
  courier_company: "",
  start: "",
  end: ""
});

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getCourierCheck(queryParams);
    tableData.value = res || [];
  } catch {
    message("加载失败", { type: "error" });
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  loadData();
};

const handleReset = () => {
  queryParams.courier_company = "";
  queryParams.start = "";
  queryParams.end = "";
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
        <span class="font-medium text-lg">快递公司对账</span>
      </template>

      <el-form :model="queryParams" inline class="mb-4">
        <el-form-item label="快递公司">
          <el-input
            v-model="queryParams.courier_company"
            placeholder="请输入快递公司"
            clearable
            style="width: 180px"
          />
        </el-form-item>
        <el-form-item label="开始日期">
          <el-date-picker
            v-model="queryParams.start"
            type="date"
            placeholder="选择日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="结束日期">
          <el-date-picker
            v-model="queryParams.end"
            type="date"
            placeholder="选择日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" stripe border>
        <el-table-column prop="courier_company" label="快递公司" width="140" />
        <el-table-column prop="inbound_count" label="入库量" width="100" />
        <el-table-column prop="outbound_count" label="出库量" width="100" />
        <el-table-column prop="delayed_count" label="滞留量" width="100" />
        <el-table-column prop="returned_count" label="退件量" width="100" />
        <el-table-column prop="avg_storage_hours" label="平均滞库时长(h)" width="150">
          <template #default="{ row }">
            {{ row.avg_storage_hours?.toFixed(1) || "-" }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
