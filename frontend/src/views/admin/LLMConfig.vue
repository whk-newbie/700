<template>
  <div class="llm-config">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>大模型配置</span>
          <el-button type="primary" @click="handleAdd">新增配置</el-button>
        </div>
      </template>
      <el-table :data="tableData" style="width: 100%">
        <el-table-column prop="name" label="配置名称" />
        <el-table-column prop="provider" label="提供商" />
        <el-table-column prop="model" label="模型" />
        <el-table-column prop="status" label="状态" />
        <el-table-column label="操作" width="250">
          <template #default>
            <el-button type="primary" link>编辑</el-button>
            <el-button type="success" link>测试</el-button>
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
  // TODO: 实现新增配置
}
</script>

<style scoped lang="less">
.llm-config {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

