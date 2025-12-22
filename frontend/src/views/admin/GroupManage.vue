<template>
  <div class="group-manage">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>分组管理</span>
          <el-button type="primary" @click="handleAdd">新增分组</el-button>
        </div>
      </template>
      <el-table :data="tableData" style="width: 100%">
        <el-table-column prop="name" label="分组名称" />
        <el-table-column prop="activation_code" label="激活码" />
        <el-table-column prop="status" label="状态" />
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
  // TODO: 实现新增分组
}
</script>

<style scoped lang="less">
.group-manage {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

