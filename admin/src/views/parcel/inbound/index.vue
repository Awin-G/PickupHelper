<script setup lang="ts">
import { ref, reactive } from "vue";
import { message } from "@/utils/message";
import { scanIn } from "@/api/parcel";
import type { ScanInRequest, ScanInResponse } from "@/api/types/parcel";
import ShelfSelector from "@/components/ShelfSelector/index.vue";
import CourierCompanySelect from "@/components/CourierCompanySelect/index.vue";
import type { FormInstance } from "element-plus";

defineOptions({
  name: "ParcelInbound"
});

const formRef = ref<FormInstance>();
const loading = ref(false);
const mode = ref<"scan" | "manual">("manual");

const form = reactive<ScanInRequest>({
  tracking_no: "",
  courier_company: "",
  receiver_phone: "",
  receiver_name: "",
  shelf_code: "",
  weight: 0,
  is_fragile: false,
  remarks: ""
});

const rules = {
  tracking_no: [{ required: true, message: "请输入快递单号", trigger: "blur" }],
  courier_company: [{ required: true, message: "请选择快递公司", trigger: "change" }],
  receiver_phone: [
    { required: true, message: "请输入收件人手机号", trigger: "blur" },
    { pattern: /^1[3-9]\d{9}$/, message: "手机号格式不正确", trigger: "blur" }
  ]
};

const successRecords = ref<ScanInResponse[]>([]);

const handleSubmit = async (formEl: FormInstance | undefined) => {
  if (!formEl) return;
  await formEl.validate(async valid => {
    if (!valid) return;
    loading.value = true;
    try {
      const res = await scanIn(form);
      message("入库成功", { type: "success" });
      successRecords.value.unshift(res);
      resetForm();
    } catch (err: any) {
      message(err?.message || "入库失败", { type: "error" });
    } finally {
      loading.value = false;
    }
  });
};

const resetForm = () => {
  form.tracking_no = "";
  form.receiver_phone = "";
  form.receiver_name = "";
  form.shelf_code = "";
  form.weight = 0;
  form.is_fragile = false;
  form.remarks = "";
  formRef.value?.resetFields();
};

const clearRecords = () => {
  successRecords.value = [];
};
</script>

<template>
  <div class="p-4">
    <el-card shadow="never">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="font-medium text-lg">包裹入库</span>
          <el-radio-group v-model="mode" size="default">
            <el-radio-button value="scan">扫码入库</el-radio-button>
            <el-radio-button value="manual">手动入库</el-radio-button>
          </el-radio-group>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        class="max-w-xl"
      >
        <el-form-item label="快递单号" prop="tracking_no">
          <el-input
            v-model="form.tracking_no"
            placeholder="请输入快递单号"
            clearable
          />
        </el-form-item>

        <el-form-item label="快递公司" prop="courier_company">
          <CourierCompanySelect v-model="form.courier_company" />
        </el-form-item>

        <el-form-item label="收件人手机" prop="receiver_phone">
          <el-input
            v-model="form.receiver_phone"
            placeholder="请输入收件人手机号"
            maxlength="11"
          />
        </el-form-item>

        <el-form-item label="收件人姓名" prop="receiver_name">
          <el-input
            v-model="form.receiver_name"
            placeholder="请输入收件人姓名（可选）"
          />
        </el-form-item>

        <el-form-item label="分配货架" prop="shelf_code">
          <ShelfSelector
            v-model="form.shelf_code"
            :station-id="1"
            placeholder="留空自动分配"
          />
        </el-form-item>

        <el-form-item label="重量(kg)">
          <el-input-number v-model="form.weight" :min="0" :precision="1" />
        </el-form-item>

        <el-form-item label="特殊标签">
          <el-checkbox v-model="form.is_fragile">易碎</el-checkbox>
        </el-form-item>

        <el-form-item label="备注">
          <el-input
            v-model="form.remarks"
            type="textarea"
            :rows="2"
            placeholder="备注信息（可选）"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSubmit(formRef)">
            确认入库
          </el-button>
          <el-button @click="resetForm">清空表单</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="mt-4" v-if="successRecords.length > 0">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="font-medium">入库成功记录</span>
          <el-button type="danger" text size="small" @click="clearRecords">
            清空记录
          </el-button>
        </div>
      </template>

      <el-table :data="successRecords" stripe>
        <el-table-column prop="parcel_id" label="包裹ID" width="100" />
        <el-table-column prop="pickup_code" label="取件码" width="120">
          <template #default="{ row }">
            <el-tag type="success" effect="dark">{{ row.pickup_code }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="shelf_code" label="货架编号" width="120" />
        <el-table-column prop="storage_time" label="入库时间" />
      </el-table>
    </el-card>
  </div>
</template>
