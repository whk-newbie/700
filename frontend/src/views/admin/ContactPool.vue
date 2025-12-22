<template>
  <div class="contact-pool">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>底库管理</span>
          <el-button type="primary" @click="handleImport">导入联系人</el-button>
        </div>
      </template>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="统计汇总" name="summary">
          <div class="summary-content">统计汇总内容</div>
        </el-tab-pane>
        <el-tab-pane label="详细列表" name="detail">
          <el-table :data="tableData" style="width: 100%">
            <el-table-column prop="line_id" label="Line ID" />
            <el-table-column prop="activation_code" label="激活码" />
            <el-table-column prop="platform" label="平台" />
            <el-table-column prop="created_at" label="创建时间" />
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const activeTab = ref('summary')
const tableData = ref([])

const handleImport = () => {
  // TODO: 实现导入功能
}
</script>

<style scoped lang="less">
.contact-pool {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

