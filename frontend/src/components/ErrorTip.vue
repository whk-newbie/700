<template>
  <div class="error-tip">
    <el-empty :description="message || '加载失败，请稍后重试'">
      <template #image>
        <el-icon :size="80" color="#c0c4cc">
          <Warning />
        </el-icon>
      </template>
      <el-button type="primary" @click="$emit('retry')">重试</el-button>
    </el-empty>
  </div>
</template>

<script setup>
import { Warning } from '@element-plus/icons-vue'

defineProps({
  message: {
    type: String,
    default: ''
  }
})

defineEmits(['retry'])
</script>

<style scoped lang="less">
.error-tip {
  padding: 40px 20px;
  text-align: center;
}
</style>

