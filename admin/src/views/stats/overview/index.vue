<script setup lang="ts">
import { ref, onMounted } from "vue";
import { message } from "@/utils/message";
import { getDashboard, getStatsTrend } from "@/api/stats";
import { ReNormalCountTo } from "@/components/ReCountTo";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

defineOptions({
  name: "StatsOverview"
});

const loading = ref(true);
const dashboardData = ref({
  date: "",
  today_inbound: 0,
  today_outbound: 0,
  pending_count: 0,
  delayed_count: 0,
  abnormal_count: 0,
  proxy_active: 0,
  shelf_usage_rate: 0
});

const trendData = ref<any[]>([]);
const granularity = ref<"day" | "week" | "month" | "year">("day");

const loadData = async () => {
  loading.value = true;
  try {
    const [dashRes, trendRes] = await Promise.all([
      getDashboard(),
      getStatsTrend({ granularity: granularity.value })
    ]);
    dashboardData.value = dashRes;
    trendData.value = trendRes.points || [];
  } catch {
    message("加载失败", { type: "error" });
  } finally {
    loading.value = false;
  }
};

const cards = [
  { title: "今日入库", key: "today_inbound", icon: "ep/upload", color: "#409eff" },
  { title: "今日出库", key: "today_outbound", icon: "ep/download", color: "#67c23a" },
  { title: "当前待取", key: "pending_count", icon: "ep/box", color: "#e6a23c" },
  { title: "滞留包裹", key: "delayed_count", icon: "ep/warning", color: "#f56c6c" },
  { title: "异常包裹", key: "abnormal_count", icon: "ep/error", color: "#f56c6c" },
  { title: "进行中代取", key: "proxy_active", icon: "ep/van", color: "#909399" }
];

const handleGranularityChange = () => {
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
          <span class="font-medium text-lg">数据概览</span>
          <el-radio-group v-model="granularity" size="default" @change="handleGranularityChange">
            <el-radio-button value="day">日</el-radio-button>
            <el-radio-button value="week">周</el-radio-button>
            <el-radio-button value="month">月</el-radio-button>
            <el-radio-button value="year">年</el-radio-button>
          </el-radio-group>
        </div>
      </template>

      <el-row :gutter="24">
        <el-col
          v-for="(card, index) in cards"
          :key="index"
          :xs="24"
          :sm="12"
          :md="8"
          class="mb-4"
        >
          <el-card shadow="never" class="h-full">
            <div class="flex justify-between items-center">
              <div>
                <p class="text-sm text-gray-500">{{ card.title }}</p>
                <ReNormalCountTo
                  class="text-2xl font-bold mt-2"
                  :duration="1000"
                  :startVal="0"
                  :endVal="dashboardData[card.key]"
                />
              </div>
              <div
                class="w-12 h-12 rounded-lg flex items-center justify-center"
                :style="{ backgroundColor: card.color + '15' }"
              >
                <IconifyIconOffline
                  :icon="card.icon"
                  :color="card.color"
                  width="24"
                  height="24"
                />
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <span class="font-medium">趋势数据</span>
      </template>
      <el-table :data="trendData" stripe border v-loading="loading">
        <el-table-column prop="date" label="日期/周期" width="150" />
        <el-table-column prop="inbound" label="入库量" width="120" />
        <el-table-column prop="outbound" label="出库量" width="120" />
        <el-table-column prop="delayed" label="新增滞留" width="120" />
      </el-table>
    </el-card>
  </div>
</template>
