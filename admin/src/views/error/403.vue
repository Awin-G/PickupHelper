<script setup lang="ts">
defineOptions({ name: "403" });
</script>

<template>
  <div class="flex justify-center items-center h-screen">
    <div class="text-center">
      <h1 class="text-6xl font-bold text-red-500">403</h1>
      <p class="text-xl text-gray-500 mt-4">无权访问</p>
      <el-button type="primary" class="mt-8" @click="$router.push('/')">返回首页</el-button>
    </div>
  </div>
</template>
