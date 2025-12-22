<template>
  <div class="account-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>账号列表</span>
        </div>
      </template>
      <el-table :data="tableData" style="width: 100%">
        <el-table-column prop="line_id" label="Line ID" />
        <el-table-column prop="platform" label="平台" />
        <el-table-column prop="status" label="状态" />
        <el-table-column prop="group_name" label="所属分组" />
        <el-table-column label="操作" width="200">
          <template #default>
            <el-button type="primary" link>查看</el-button>
            <el-button type="danger" link>删除</el-button>
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
.account-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

