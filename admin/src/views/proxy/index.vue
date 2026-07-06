<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { getProxyOrders } from "@/api/proxy";
import type { ProxyOrderItem } from "@/api/types/parcel";
import { formatDateTime } from "@/utils/format/datetime";
import { formatMoneyWithSymbol } from "@/utils/format/money";

defineOptions({
  name: "ProxyOrders"
});

const loading = ref(true);
const tableData = ref<ProxyOrderItem[]>([]);
const total = ref(0);

const queryParams = reactive({
  role: "",
  status: undefined as number | undefined,
  page: 1,
  page_size: 20
});

const statusOptions = [
  { label: "待接单", value: 1 },
  { label: "配送中", value: 2 },
  { label: "待确认", value: 3 },
  { label: "已完成", value: 4 },
  { label: "已取消", value: 5 },
  { label: "取件失败", value: 6 }
];

const statusMap: Record<number, { text: string; type: string }> = {
  1: { text: "待接单", type: "primary" },
  2: { text: "配送中", type: "warning" },
  3: { text: "待确认", type: "info" },
  4: { text: "已完成", type: "success" },
  5: { text: "已取消", type: "info" },
  6: { text: "取件失败", type: "danger" }
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getProxyOrders(queryParams);
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
  queryParams.role = "";
  queryParams.status = undefined;
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
        <span class="font-medium text-lg">代取订单</span>
      </template>

      <el-form :model="queryParams" inline class="mb-4">
        <el-form-item label="订单状态">
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
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" stripe border>
        <el-table-column prop="id" label="订单ID" width="80" />
        <el-table-column prop="parcel_id" label="包裹ID" width="80" />
        <el-table-column prop="station_name" label="驿站" width="120" />
        <el-table-column prop="publisher_nickname" label="发布者" width="100" />
        <el-table-column prop="taker_nickname" label="跑腿员" width="100">
          <template #default="{ row }">
            {{ row.taker_nickname || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="reward_amount" label="悬赏金额" width="100">
          <template #default="{ row }">
            <span class="text-red-500 font-medium">
              {{ formatMoneyWithSymbol(row.reward_amount) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="(statusMap[row.status]?.type as any) || 'info'" size="small">
              {{ statusMap[row.status]?.text || "未知" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="deadline" label="截止时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.deadline) }}
          </template>
        </el-table-column>
        <el-table-column prop="delivery_time" label="送达时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.delivery_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="mt-4 flex justify-end">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadData"
          @size-change="loadData"
        />
      </div>
    </el-card>
  </div>
</template>
