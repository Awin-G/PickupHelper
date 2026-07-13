<script setup lang="ts">
import { ref } from "vue";
import { message } from "@/utils/message";
import { useUserStoreHook } from "@/store/modules/user";

defineOptions({ name: "AccountSettings" });

const userStore = useUserStoreHook();
const nickname = ref(userStore.nickname);
const loading = ref(false);

const handleSave = async () => {
  loading.value = true;
  try {
    message("保存成功", { type: "success" });
  } catch {
    message("保存失败", { type: "error" });
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="p-4">
    <el-card shadow="never">
      <template #header>
        <span class="font-medium text-lg">账户设置</span>
      </template>
      <el-form label-width="80px" class="max-w-md">
        <el-form-item label="用户名">
          <el-input :model-value="userStore.username" disabled />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSave">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>
