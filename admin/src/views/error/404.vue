<script setup lang="ts">
defineOptions({ name: "404" });
</script>

<template>
  <div class="flex justify-center items-center h-screen">
    <div class="text-center">
      <h1 class="text-6xl font-bold text-gray-500">404</h1>
      <p class="text-xl text-gray-500 mt-4">页面不存在</p>
      <el-button type="primary" class="mt-8" @click="$router.push('/')">返回首页</el-button>
    </div>
  </div>
</template>
