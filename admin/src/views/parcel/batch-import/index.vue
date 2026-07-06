<script setup lang="ts">
import { ref } from "vue";
import { message } from "@/utils/message";
import { batchImport } from "@/api/parcel";

defineOptions({
  name: "ParcelBatchImport"
});

const fileList = ref<File[]>([]);
const loading = ref(false);
const uploadResult = ref<{
  batch_id: string;
  total: number;
  status: string;
} | null>(null);

const handleFileChange = (file: File) => {
  fileList.value = [file];
};

const handleUpload = async () => {
  if (fileList.value.length === 0) {
    message("请选择文件", { type: "warning" });
    return;
  }

  const file = fileList.value[0];
  if (!file.name.endsWith(".xlsx")) {
    message("仅支持 .xlsx 格式文件", { type: "warning" });
    return;
  }

  if (file.size > 5 * 1024 * 1024) {
    message("文件大小不能超过 5MB", { type: "warning" });
    return;
  }

  loading.value = true;
  try {
    const res = await batchImport({ file, station_id: 1 });
    uploadResult.value = res;
    message("文件上传成功，正在处理中", { type: "success" });
  } catch (err: any) {
    message(err?.message || "上传失败", { type: "error" });
  } finally {
    loading.value = false;
  }
};

const handleDownloadTemplate = () => {
  message("模板下载功能待实现", { type: "info" });
};
</script>

<template>
  <div class="p-4">
    <el-card shadow="never">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="font-medium text-lg">批量导入包裹</span>
          <el-button type="primary" link @click="handleDownloadTemplate">
            <IconifyIconOffline icon="ep/download" class="mr-1" />
            下载导入模板
          </el-button>
        </div>
      </template>

      <el-upload
        class="upload-demo"
        drag
        :auto-upload="false"
        accept=".xlsx"
        :limit="1"
        :on-change="(file: any) => handleFileChange(file.raw)"
      >
        <IconifyIconOffline icon="ep/upload-filled" width="48" height="48" class="text-gray-400" />
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            仅支持 .xlsx 格式文件，单文件不超过 5MB
          </div>
        </template>
      </el-upload>

      <div class="mt-4">
        <el-button
          type="primary"
          :loading="loading"
          :disabled="fileList.length === 0"
          @click="handleUpload"
        >
          开始导入
        </el-button>
      </div>

      <el-card v-if="uploadResult" shadow="never" class="mt-6">
        <template #header>
          <span class="font-medium">导入结果</span>
        </template>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="批次ID">
            {{ uploadResult.batch_id }}
          </el-descriptions-item>
          <el-descriptions-item label="总行数">
            {{ uploadResult.total }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag type="warning">{{ uploadResult.status }}</el-tag>
          </el-descriptions-item>
        </el-descriptions>
        <p class="mt-4 text-sm text-gray-500">
          文件已提交处理，请稍后查看导入结果。
        </p>
      </el-card>
    </el-card>
  </div>
</template>
