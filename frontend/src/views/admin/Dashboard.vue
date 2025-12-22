<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6" v-for="stat in stats" :key="stat.title">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stat.value }}</div>
            <div class="stat-title">{{ stat.title }}</div>
          </div>
          <div class="stat-icon" :style="{ color: stat.color }">
            <el-icon :size="40">
              <component :is="stat.icon" />
            </el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { 
  User, 
  Folder, 
  List, 
  TrendCharts 
} from '@element-plus/icons-vue'

const iconMap = {
  Folder,
  User,
  List,
  TrendCharts
}

const stats = ref([
  {
    title: '总分组数',
    value: 0,
    icon: Folder,
    color: '#409eff'
  },
  {
    title: '总账号数',
    value: 0,
    icon: User,
    color: '#67c23a'
  },
  {
    title: '今日进线',
    value: 0,
    icon: List,
    color: '#e6a23c'
  },
  {
    title: '总进线数',
    value: 0,
    icon: TrendCharts,
    color: '#f56c6c'
  }
])
</script>

<style scoped lang="less">
.dashboard {
  .stat-card {
    position: relative;
    overflow: hidden;
    
    .stat-content {
      .stat-value {
        font-size: 32px;
        font-weight: bold;
        color: #303133;
        margin-bottom: 8px;
      }
      
      .stat-title {
        font-size: 14px;
        color: #909399;
      }
    }
    
    .stat-icon {
      position: absolute;
      right: 20px;
      top: 50%;
      transform: translateY(-50%);
      opacity: 0.3;
    }
  }
}
</style>

