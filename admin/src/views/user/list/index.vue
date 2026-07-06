<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "@/utils/message";
import { getUserList } from "@/api/user";
import type { UserItem } from "@/api/types/parcel";
import { maskPhone } from "@/utils/format/phone";
import { formatDateTime } from "@/utils/format/datetime";

defineOptions({
  name: "UserList"
});

const loading = ref(true);
const tableData = ref<UserItem[]>([]);
const total = ref(0);

const queryParams = reactive({
  keyword: "",
  user_type: undefined as number | undefined,
  is_blacklisted: undefined as number | undefined,
  page: 1,
  page_size: 20
});

const userTypeOptions = [
  { label: "普通用户", value: 1 },
  { label: "跑腿员", value: 2 }
];

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getUserList(queryParams);
    tableData.value = res.list;
    total.value = res.total;
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
  queryParams.keyword = "";
  queryParams.user_type = undefined;
  queryParams.is_blacklisted = undefined;
  queryParams.page = 1;
  loadData();
};

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="p-4">
    <el-card shadow="never">
      <template #header>
        <span class="font-medium text-lg">用户列表</span>
      </template>

      <el-form :model="queryParams" inline class="mb-4">
        <el-form-item label="搜索">
          <el-input
            v-model="queryParams.keyword"
            placeholder="姓名/手机号"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="用户类型">
          <el-select
            v-model="queryParams.user_type"
            placeholder="全部类型"
            clearable
            style="width: 140px"
          >
            <el-option
              v-for="item in userTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="黑名单">
          <el-select
            v-model="queryParams.is_blacklisted"
            placeholder="全部"
            clearable
            style="width: 120px"
          >
            <el-option label="是" :value="1" />
            <el-option label="否" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column prop="phone" label="手机号" width="130">
          <template #default="{ row }">
            {{ maskPhone(row.phone) }}
          </template>
        </el-table-column>
        <el-table-column prop="user_type" label="用户类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.user_type === 2 ? 'warning' : 'primary'" size="small">
              {{ row.user_type === 2 ? "跑腿员" : "普通用户" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_blacklisted" label="黑名单" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.is_blacklisted" type="danger" size="small">是</el-tag>
            <el-tag v-else type="success" size="small">否</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="credit_score" label="信用分" width="80" />
        <el-table-column prop="created_at" label="注册时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="mt-4 flex justify-end">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadData"
          @size-change="loadData"
        />
      </div>
    </el-card>
  </div>
</template>
