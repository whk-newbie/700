<template>
  <div class="skeleton-container">
    <el-skeleton :rows="rows" :animated="true" />
  </div>
</template>

<script setup>
defineProps({
  rows: {
    type: Number,
    default: 5
  }
})
</script>

<style scoped lang="less">
.skeleton-container {
  padding: 20px;
}
</style>

