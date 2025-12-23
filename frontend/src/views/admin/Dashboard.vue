<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="24" :sm="12" :md="6" v-for="stat in stats" :key="stat.title">
        <el-card class="stat-card" :class="{ 'updating': stat.updating }">
          <div class="stat-content">
            <div class="stat-value" :class="stat.animationClass">
              {{ stat.value }}
            </div>
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

    <!-- 进线趋势图 -->
    <el-card class="chart-card" style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>进线趋势</span>
          <el-radio-group v-model="trendDays" size="small" @change="fetchTrendData">
            <el-radio-button :label="7">近7天</el-radio-button>
            <el-radio-button :label="15">近15天</el-radio-button>
            <el-radio-button :label="30">近30天</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <div ref="chartContainer" class="chart-container"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { 
  User, 
  Folder, 
  List, 
  TrendCharts 
} from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { getOverviewStats } from '@/api/stats'
import { getGroups } from '@/api/group'
import { useWebSocketStore } from '@/store/websocket'
import { ElMessage } from 'element-plus'

const chartContainer = ref(null)
let chartInstance = null
const trendDays = ref(7)

const stats = ref([
  {
    title: '总分组数',
    value: 0,
    icon: Folder,
    color: '#409eff',
    updating: false,
    animationClass: ''
  },
  {
    title: '总账号数',
    value: 0,
    icon: User,
    color: '#67c23a',
    updating: false,
    animationClass: ''
  },
  {
    title: '今日进线',
    value: 0,
    icon: List,
    color: '#e6a23c',
    updating: false,
    animationClass: ''
  },
  {
    title: '总进线数',
    value: 0,
    icon: TrendCharts,
    color: '#f56c6c',
    updating: false,
    animationClass: ''
  }
])

// WebSocket Store
const wsStore = useWebSocketStore()

// 获取统计数据
const fetchStats = async () => {
  try {
    const res = await getOverviewStats()
    if (res.code === 1000 && res.data) {
      const data = res.data
      
      // 更新统计卡片，添加动画效果
      updateStatValue(0, data.total_groups || 0)
      updateStatValue(1, data.total_accounts || 0)
      updateStatValue(2, data.today_incoming || 0)
      updateStatValue(3, data.total_incoming || 0)
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

// 更新统计值（带动画）
const updateStatValue = (index, newValue) => {
  const stat = stats.value[index]
  if (stat.value === newValue) return
  
  stat.updating = true
  stat.animationClass = 'value-update'
  
  setTimeout(() => {
    stat.value = newValue
    stat.animationClass = ''
    setTimeout(() => {
      stat.updating = false
    }, 300)
  }, 150)
}

// 获取趋势数据
const fetchTrendData = async () => {
  if (!chartContainer.value) return
  
  try {
    // 获取所有分组
    const groupsRes = await getGroups({ page: 1, page_size: 100 })
    if (groupsRes.code !== 1000 || !groupsRes.data) {
      return
    }
    
    const groups = groupsRes.data.list || groupsRes.data.data || []
    if (groups.length === 0) {
      return
    }
    
    // 获取第一个分组的趋势数据（这里可以根据需要调整）
    const groupId = groups[0].id
    const { getGroupIncomingTrend } = await import('@/api/stats')
    const trendRes = await getGroupIncomingTrend(groupId, trendDays.value)
    
    if (trendRes.code === 1000 && trendRes.data) {
      const trendData = trendRes.data
      updateChart(trendData)
    }
  } catch (error) {
    console.error('获取趋势数据失败:', error)
  }
}

// 更新图表
const updateChart = (data) => {
  if (!chartInstance) {
    chartInstance = echarts.init(chartContainer.value)
  }
  
  const dates = data.map(item => item.date)
  const incoming = data.map(item => item.incoming || 0)
  const duplicate = data.map(item => item.duplicate || 0)
  
  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      }
    },
    legend: {
      data: ['进线数', '重复数']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '进线数',
        type: 'line',
        smooth: true,
        data: incoming,
        itemStyle: {
          color: '#409eff'
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
              { offset: 1, color: 'rgba(64, 158, 255, 0.1)' }
            ]
          }
        }
      },
      {
        name: '重复数',
        type: 'line',
        smooth: true,
        data: duplicate,
        itemStyle: {
          color: '#e6a23c'
        }
      }
    ]
  }
  
  chartInstance.setOption(option)
}

// 初始化WebSocket消息处理器
const initWebSocket = () => {
  wsStore.registerMessageHandler('dashboard', (message) => {
    if (message.type === 'stats_update' || message.type === 'incoming_update') {
      fetchStats()
      fetchTrendData()
    }
  })
}

// 窗口大小变化时调整图表
const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

onMounted(async () => {
  await fetchStats()
  await nextTick()
  if (chartContainer.value) {
    await fetchTrendData()
  }
  initWebSocket()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  wsStore.unregisterMessageHandler('dashboard')
  window.removeEventListener('resize', handleResize)
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
})
</script>

<style scoped lang="less">
.dashboard {
  .stats-row {
    margin-bottom: 20px;
  }

  .stat-card {
    position: relative;
    overflow: hidden;
    transition: all 0.3s;
    cursor: pointer;
    
    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }
    
    &.updating {
      background: linear-gradient(90deg, #f0f2f5 25%, #e4e7ed 50%, #f0f2f5 75%);
      background-size: 200% 100%;
      animation: shimmer 1.5s infinite;
    }
    
    .stat-content {
      .stat-value {
        font-size: 32px;
        font-weight: bold;
        color: #303133;
        margin-bottom: 8px;
        transition: all 0.3s;
        
        &.value-update {
          animation: pulse 0.3s ease-in-out;
        }
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
      transition: opacity 0.3s;
    }
    
    &:hover .stat-icon {
      opacity: 0.5;
    }
  }

  .chart-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    
    .chart-container {
      width: 100%;
      height: 400px;
    }
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

@keyframes shimmer {
  0% {
    background-position: -200% 0;
  }
  100% {
    background-position: 200% 0;
  }
}

// 响应式布局
@media (max-width: 768px) {
  .dashboard {
    .stat-card {
      .stat-content {
        .stat-value {
          font-size: 24px;
        }
      }
      
      .stat-icon {
        right: 10px;
        
        :deep(.el-icon) {
          font-size: 30px !important;
        }
      }
    }
    
    .chart-card {
      .chart-container {
        height: 300px;
      }
    }
  }
}
</style>

