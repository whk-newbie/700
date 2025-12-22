<template>
  <div class="leads-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>线索列表</span>
        </div>
      </template>
      <el-table :data="tableData" style="width: 100%">
        <el-table-column prop="incoming_line_id" label="进线Line ID" />
        <el-table-column prop="line_account_id" label="账号ID" />
        <el-table-column prop="is_duplicate" label="是否重复" />
        <el-table-column prop="incoming_time" label="进线时间" />
        <el-table-column label="操作" width="200">
          <template #default>
            <el-button type="primary" link>查看详情</el-button>
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
.leads-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

