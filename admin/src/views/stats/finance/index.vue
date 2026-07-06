<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { getProxyFinance } from "@/api/stats";
import { formatMoney, formatMoneyWithSymbol } from "@/utils/format/money";

defineOptions({
  name: "StatsFinance"
});

const loading = ref(true);
const financeData = ref({
  total_orders: 0,
  completed_orders: 0,
  total_reward: 0,
  avg_reward: 0,
  by_taker: [] as any[]
});

const queryParams = reactive({
  station_id: undefined as number | undefined,
  start: "",
  end: ""
});

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getProxyFinance(queryParams);
    financeData.value = res;
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
    <el-card shadow="never" class="mb-4">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="font-medium text-lg">代取财务汇总</span>
          <el-button type="primary" @click="$router.push('/stats/overview')">返回概览</el-button>
        </div>
      </template>

      <el-form :model="queryParams" inline class="mb-4">
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

      <el-row :gutter="24" class="mb-4">
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="never">
            <p class="text-sm text-gray-500">订单总数</p>
            <p class="text-2xl font-bold mt-1">{{ financeData.total_orders }}</p>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="never">
            <p class="text-sm text-gray-500">已完成订单</p>
            <p class="text-2xl font-bold mt-1 text-green-600">{{ financeData.completed_orders }}</p>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="never">
            <p class="text-sm text-gray-500">总悬赏金额</p>
            <p class="text-2xl font-bold mt-1 text-red-500">
              {{ formatMoneyWithSymbol(financeData.total_reward) }}
            </p>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="never">
            <p class="text-sm text-gray-500">平均悬赏</p>
            <p class="text-2xl font-bold mt-1 text-orange-500">
              {{ formatMoneyWithSymbol(financeData.avg_reward) }}
            </p>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <span class="font-medium">跑腿员收益排行</span>
      </template>
      <el-table :data="financeData.by_taker" stripe border v-loading="loading">
        <el-table-column type="index" label="排名" width="60" />
        <el-table-column prop="taker_nickname" label="跑腿员" width="140" />
        <el-table-column prop="order_count" label="完成单数" width="100" />
        <el-table-column prop="total_reward" label="总收益" width="120">
          <template #default="{ row }">
            <span class="text-red-500 font-medium">
              {{ formatMoneyWithSymbol(row.total_reward) }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
