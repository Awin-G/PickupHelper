<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { getRunnerApplications, auditRunnerApplication } from "@/api/user";
import type { RunnerApplicationItem } from "@/api/types/parcel";
import { formatDateTime } from "@/utils/format/datetime";
import { maskPhone } from "@/utils/format/phone";
import type { FormInstance } from "element-plus";

defineOptions({
  name: "RunnerAudit"
});

const loading = ref(true);
const tableData = ref<RunnerApplicationItem[]>([]);
const total = ref(0);
const auditDialogVisible = ref(false);
const auditLoading = ref(false);
const currentApplication = ref<RunnerApplicationItem | null>(null);

const queryParams = reactive({
  status: undefined as number | undefined,
  keyword: "",
  page: 1,
  page_size: 20
});

const auditForm = reactive({
  action: "approve" as "approve" | "reject",
  audit_remark: ""
});

const statusOptions = [
  { label: "审核中", value: 1 },
  { label: "已通过", value: 2 },
  { label: "已拒绝", value: 3 }
];

const statusMap: Record<number, { text: string; type: string }> = {
  1: { text: "审核中", type: "warning" },
  2: { text: "已通过", type: "success" },
  3: { text: "已拒绝", type: "danger" }
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getRunnerApplications(queryParams);
    tableData.value = res.data.list;
    total.value = res.data.total;
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
  queryParams.status = undefined;
  queryParams.keyword = "";
  queryParams.page = 1;
  loadData();
};

const handleAudit = (row: RunnerApplicationItem) => {
  currentApplication.value = row;
  auditForm.action = "approve";
  auditForm.audit_remark = "";
  auditDialogVisible.value = true;
};

const handleConfirmAudit = async () => {
  if (!currentApplication.value) return;
  auditLoading.value = true;
  try {
    await auditRunnerApplication(currentApplication.value.id, auditForm);
    message("审核成功", { type: "success" });
    auditDialogVisible.value = false;
    loadData();
  } catch (err: any) {
    message(err?.message || "审核失败", { type: "error" });
  } finally {
    auditLoading.value = false;
  }
};

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="p-4">
    <el-card shadow="never">
      <template #header>
        <span class="font-medium text-lg">跑腿员审核</span>
      </template>

      <el-form :model="queryParams" inline class="mb-4">
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
        <el-form-item label="搜索">
          <el-input
            v-model="queryParams.keyword"
            placeholder="姓名/手机号"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" stripe border>
        <el-table-column prop="id" label="申请ID" width="80" />
        <el-table-column prop="real_name" label="真实姓名" width="100" />
        <el-table-column prop="phone" label="手机号" width="130">
          <template #default="{ row }">
            {{ maskPhone(row.phone) }}
          </template>
        </el-table-column>
        <el-table-column prop="student_id" label="学号/工号" width="120" />
        <el-table-column prop="id_card_image" label="证件照" width="100">
          <template #default="{ row }">
            <el-image
              v-if="row.id_card_image"
              :src="row.id_card_image"
              :preview-src-list="[row.id_card_image]"
              fit="cover"
              style="width: 40px; height: 40px"
              class="rounded"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="(statusMap[row.status]?.type as any) || 'info'" size="small">
              {{ statusMap[row.status]?.text || "未知" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="audit_remark" label="审核备注" min-width="150" show-overflow-tooltip />
        <el-table-column prop="created_at" label="申请时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 1"
              type="primary"
              link
              size="small"
              @click="handleAudit(row as RunnerApplicationItem)"
            >
              审核
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

    <el-dialog v-model="auditDialogVisible" title="审核跑腿员申请" width="450px">
      <div v-if="currentApplication" class="mb-4">
        <p><strong>申请人：</strong>{{ currentApplication.real_name }}</p>
        <p><strong>手机号：</strong>{{ maskPhone(currentApplication.phone) }}</p>
        <p v-if="currentApplication.student_id">
          <strong>学号：</strong>{{ currentApplication.student_id }}
        </p>
      </div>
      <el-form :model="auditForm" label-width="80px">
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.action">
            <el-radio value="approve">通过</el-radio>
            <el-radio value="reject">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input
            v-model="auditForm.audit_remark"
            type="textarea"
            :rows="3"
            placeholder="审核备注（可选）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditDialogVisible = false">取消</el-button>
        <el-button
          :type="auditForm.action === 'approve' ? 'success' : 'danger'"
          :loading="auditLoading"
          @click="handleConfirmAudit"
        >
          {{ auditForm.action === "approve" ? "确认通过" : "确认拒绝" }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
