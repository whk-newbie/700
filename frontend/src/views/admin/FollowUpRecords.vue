<template>
  <div class="follow-up-records">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>跟进记录</span>
          <el-button type="primary" @click="handleAdd">新增记录</el-button>
        </div>
      </template>
      <el-table :data="tableData" style="width: 100%">
        <el-table-column prop="customer_line_id" label="客户Line ID" />
        <el-table-column prop="content" label="跟进内容" />
        <el-table-column prop="created_at" label="创建时间" />
        <el-table-column label="操作" width="200">
          <template #default>
            <el-button type="primary" link>编辑</el-button>
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

const handleAdd = () => {
  // TODO: 实现新增记录
}
</script>

<style scoped lang="less">
.follow-up-records {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

