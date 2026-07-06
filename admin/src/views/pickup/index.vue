<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { verifyPickup, getPickupLogs } from "@/api/pickup";
import type { PickupLogItem } from "@/api/types/parcel";
import { formatDateTime } from "@/utils/format/datetime";
import type { FormInstance } from "element-plus";

defineOptions({
  name: "PickupVerify"
});

const formRef = ref<FormInstance>();
const loading = ref(false);
const mode = ref<"scan" | "manual">("manual");
const verifyResult = ref<any>(null);

const form = reactive({
  pickup_code: "",
  verification_method: 2 as 1 | 2,
  station_id: 1
});

const rules = {
  pickup_code: [
    { required: true, message: "请输入取件码", trigger: "blur" },
    { pattern: /^\d{6}$/, message: "取件码为6位数字", trigger: "blur" }
  ]
};

const logs = ref<PickupLogItem[]>([]);
const logsLoading = ref(false);

const handleVerify = async (formEl: FormInstance | undefined) => {
  if (!formEl) return;
  await formEl.validate(async valid => {
    if (!valid) return;
    loading.value = true;
    try {
      form.verification_method = mode.value === "scan" ? 1 : 2;
      const res = await verifyPickup(form);
      verifyResult.value = res;
      message("核销成功", { type: "success" });
      form.pickup_code = "";
      formRef.value?.resetFields();
      loadLogs();
    } catch (err: any) {
      message(err?.message || "核销失败", { type: "error" });
    } finally {
      loading.value = false;
    }
  });
};

const loadLogs = async () => {
  logsLoading.value = true;
  try {
    const res = await getPickupLogs({ page: 1, page_size: 10 });
    logs.value = res.list;
  } catch {
    // 忽略错误
  } finally {
    logsLoading.value = false;
  }
};

onMounted(() => {
  loadLogs();
});
</script>

<template>
  <div class="p-4">
    <el-row :gutter="24">
      <el-col :xs="24" :md="10">
        <el-card shadow="never">
          <template #header>
            <div class="flex justify-between items-center">
              <span class="font-medium text-lg">出库核销</span>
              <el-radio-group v-model="mode" size="default">
                <el-radio-button value="scan">扫码核销</el-radio-button>
                <el-radio-button value="manual">手动输入</el-radio-button>
              </el-radio-group>
            </div>
          </template>

          <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
            <el-form-item label="取件码" prop="pickup_code">
              <el-input
                v-model="form.pickup_code"
                placeholder="请输入6位取件码"
                maxlength="6"
                size="large"
                class="text-center"
              />
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                size="large"
                class="w-full"
                :loading="loading"
                @click="handleVerify(formRef)"
              >
                确认核销
              </el-button>
            </el-form-item>
          </el-form>

          <el-card v-if="verifyResult" shadow="never" class="mt-4 bg-green-50">
            <div class="text-center">
              <IconifyIconOffline icon="ep/check-circle" width="48" height="48" color="#67c23a" />
              <p class="mt-2 text-lg font-medium text-green-600">核销成功</p>
              <el-descriptions :column="1" class="mt-4" border size="small">
                <el-descriptions-item label="包裹ID">
                  {{ verifyResult.parcel_id }}
                </el-descriptions-item>
                <el-descriptions-item label="快递单号">
                  {{ verifyResult.tracking_no }}
                </el-descriptions-item>
                <el-descriptions-item label="取件时间">
                  {{ formatDateTime(verifyResult.pickup_time) }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </el-card>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="14">
        <el-card shadow="never">
          <template #header>
            <span class="font-medium text-lg">最近核销记录</span>
          </template>

          <el-table v-loading="logsLoading" :data="logs" stripe size="small">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="parcel_id" label="包裹ID" width="80" />
            <el-table-column prop="operator_type" label="操作人类型" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.operator_type === 1" type="primary" size="small">管理员</el-tag>
                <el-tag v-else-if="row.operator_type === 3" type="warning" size="small">跑腿员</el-tag>
                <el-tag v-else size="small">本人</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="verification_method" label="核验方式" width="100">
              <template #default="{ row }">
                {{ row.verification_method === 1 ? "扫码" : "手动" }}
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" width="170">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
