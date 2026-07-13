<script setup lang="ts">
import { ref, onMounted } from "vue";
import { getDashboard } from "@/api/stats";
import { ReNormalCountTo } from "@/components/ReCountTo";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

defineOptions({
  name: "Welcome"
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

const cards = ref([
  { title: "今日入库", key: "today_inbound", icon: "ep/upload", color: "#409eff" },
  { title: "今日出库", key: "today_outbound", icon: "ep/download", color: "#67c23a" },
  { title: "当前待取", key: "pending_count", icon: "ep box", color: "#e6a23c" },
  { title: "滞留包裹", key: "delayed_count", icon: "ep/warning", color: "#f56c6c" },
  { title: "异常包裹", key: "abnormal_count", icon: "ep/error", color: "#f56c6c" },
  { title: "进行中代取", key: "proxy_active", icon: "ep/van", color: "#909399" }
]);

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getDashboard();
    dashboardData.value = res.data;
  } catch {
    // 使用默认数据
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="p-4">
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

    <el-row :gutter="24">
      <el-col :xs="24" :md="12" class="mb-4">
        <el-card shadow="never">
          <template #header>
            <span class="font-medium">货架占用率</span>
          </template>
          <div class="text-center py-8">
            <el-progress
              type="dashboard"
              :percentage="dashboardData.shelf_usage_rate * 100"
              :width="180"
              :stroke-width="12"
              :color="dashboardData.shelf_usage_rate > 0.9 ? '#f56c6c' : dashboardData.shelf_usage_rate > 0.7 ? '#e6a23c' : '#409eff'"
            >
              <template #default>
                <span class="text-2xl font-bold">
                  {{ (dashboardData.shelf_usage_rate * 100).toFixed(1) }}%
                </span>
                <br />
                <span class="text-xs text-gray-500">货架占用率</span>
              </template>
            </el-progress>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12" class="mb-4">
        <el-card shadow="never">
          <template #header>
            <span class="font-medium">快捷操作</span>
          </template>
          <div class="grid grid-cols-2 gap-4">
            <el-button type="primary" @click="$router.push('/parcel/inbound')">
              <IconifyIconOffline icon="ep/upload" class="mr-1" />
              包裹入库
            </el-button>
            <el-button type="success" @click="$router.push('/pickup/verify')">
              <IconifyIconOffline icon="ep/check" class="mr-1" />
              出库核销
            </el-button>
            <el-button type="warning" @click="$router.push('/parcel/list')">
              <IconifyIconOffline icon="ep/list" class="mr-1" />
              包裹列表
            </el-button>
            <el-button @click="$router.push('/shelf/list')">
              <IconifyIconOffline icon="ep/grid" class="mr-1" />
              货架管理
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
