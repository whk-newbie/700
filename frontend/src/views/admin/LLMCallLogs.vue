<template>
  <div class="llm-call-logs">
    <el-card>
      <template #header>
        <span>大模型调用记录</span>
      </template>
      <el-table :data="tableData" style="width: 100%">
        <el-table-column prop="config_name" label="配置名称" />
        <el-table-column prop="model" label="模型" />
        <el-table-column prop="tokens" label="Token数" />
        <el-table-column prop="cost" label="成本" />
        <el-table-column prop="created_at" label="调用时间" />
        <el-table-column label="操作" width="150">
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
.llm-call-logs {
}
</style>

