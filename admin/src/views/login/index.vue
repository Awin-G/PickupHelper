<script setup lang="ts">
import { useRouter } from "vue-router";
import { message } from "@/utils/message";
import { debounce } from "@pureadmin/utils";
import { useNav } from "@/layout/hooks/useNav";
import type { FormInstance } from "element-plus";
import { useUserStoreHook } from "@/store/modules/user";
import { initRouter, getTopMenu } from "@/router/utils";
import { ref, reactive } from "vue";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

import Lock from "~icons/ri/lock-fill";
import User from "~icons/ri/user-3-fill";

defineOptions({
  name: "Login"
});

const router = useRouter();
const loading = ref(false);
const disabled = ref(false);
const ruleFormRef = ref<FormInstance>();

const { title } = useNav();

const ruleForm = reactive({
  username: "",
  password: ""
});

const loginRules = {
  username: [
    { required: true, message: "请输入用户名", trigger: "blur" }
  ],
  password: [
    { required: true, message: "请输入密码", trigger: "blur" },
    { min: 6, message: "密码长度不能少于6位", trigger: "blur" }
  ]
};

const onLogin = async (formEl: FormInstance | undefined) => {
  if (!formEl) return;
  await formEl.validate(valid => {
    if (valid) {
      loading.value = true;
      useUserStoreHook()
        .loginByUsername({
          username: ruleForm.username,
          password: ruleForm.password
        })
        .then(async () => {
          await initRouter();
          disabled.value = true;
          router.push(getTopMenu(true).path).then(() => {
            message("登录成功", { type: "success" });
          });
        })
        .catch(_err => {
          message("登录失败", { type: "error" });
        })
        .finally(() => {
          disabled.value = false;
          loading.value = false;
        });
    }
  });
};

const immediateDebounce: any = debounce(
  formRef => onLogin(formRef),
  1000,
  true
);
</script>

<template>
  <div class="select-none">
    <img :src="`https://unpkg.com/@pureadmin/theme@2.0.4/public/previews/dark/assets/avatar.svg`" class="bg" />
    <div class="login-box">
      <div class="login-form">
        <el-logo class="flex items-center justify-center mb-6">
          <IconifyIconOffline icon="ep/box" width="32" height="32" />
          <span class="ml-2 text-xl font-bold">{{ title || "快递代取管理系统" }}</span>
        </el-logo>
        <el-form ref="ruleFormRef" :model="ruleForm" :rules="loginRules" size="large">
          <el-form-item prop="username">
            <el-input
              v-model="ruleForm.username"
              placeholder="请输入用户名"
              :prefix-icon="useRenderIcon(User)"
              @keyup.enter="immediateDebounce(ruleFormRef)"
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="ruleForm.password"
              type="password"
              placeholder="请输入密码"
              show-password
              :prefix-icon="useRenderIcon(Lock)"
              @keyup.enter="immediateDebounce(ruleFormRef)"
            />
          </el-form-item>
          <el-button
            class="w-full"
            type="primary"
            size="large"
            :loading="loading"
            :disabled="disabled"
            @click="onLogin(ruleFormRef)"
          >
            登 录
          </el-button>
        </el-form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bg {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: -1;
  filter: blur(2px);
}

.login-box {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100vh;
}

.login-form {
  width: 400px;
  padding: 40px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

:deep(.el-logo) {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
