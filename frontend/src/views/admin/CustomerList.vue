<template>
  <div class="customer-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>客户列表</span>
        </div>
      </template>
      <el-table :data="tableData" style="width: 100%">
        <el-table-column prop="line_id" label="Line ID" />
        <el-table-column prop="platform" label="平台" />
        <el-table-column prop="customer_type" label="客户类型" />
        <el-table-column prop="created_at" label="创建时间" />
        <el-table-column label="操作" width="200">
          <template #default>
            <el-button type="primary" link>查看详情</el-button>
            <el-button type="primary" link>跟进记录</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const tableData = ref([])
</script>

<style scoped lang="less">
.customer-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

