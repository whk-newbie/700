<template>
  <div class="list-page-container leads-list">
    <!-- 统计卡片 -->
    <div class="stats-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-item">
              <div class="stats-label">总进线</div>
              <div class="stats-value">{{ overviewStats.total_incoming || 0 }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-item">
              <div class="stats-label">今日进线</div>
              <div class="stats-value highlight">{{ overviewStats.today_incoming || 0 }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-item">
              <div class="stats-label">重复进线</div>
              <div class="stats-value warning">{{ overviewStats.duplicate_incoming || 0 }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-item">
              <div class="stats-label">今日重复</div>
              <div class="stats-value warning">{{ overviewStats.today_duplicate || 0 }}</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <el-card style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>线索列表</span>
        </div>
      </template>

      <!-- 筛选区域 -->
      <div class="filter-section">
        <el-form :model="filterForm" :inline="true" class="filter-form">
          <el-form-item label="分组">
            <el-select
              v-model="filterForm.group_id"
              placeholder="全部"
              clearable
              filterable
              style="width: 200px"
            >
              <el-option
                v-for="group in groupList"
                :key="group.id"
                :label="`${group.activation_code}${group.remark ? ' - ' + group.remark : ''}`"
                :value="group.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="是否重复">
            <el-select
              v-model="filterForm.is_duplicate"
              placeholder="全部"
              clearable
              style="width: 120px"
            >
              <el-option label="是" :value="true" />
              <el-option label="否" :value="false" />
            </el-select>
          </el-form-item>
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              style="width: 400px"
              @change="handleDateRangeChange"
            />
          </el-form-item>
          <el-form-item label="搜索">
            <el-input
              v-model="filterForm.search"
              placeholder="进线Line ID或显示名称"
              clearable
              style="width: 200px"
              @keyup.enter="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch" :loading="loading">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
            <el-button @click="handleReset">重置</el-button>
            <el-button type="primary" :disabled="loading" @click="handleRefresh">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 数据表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        stripe
        empty-text="暂无数据"
        :default-expand-all="false"
      >
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-content">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="进线Line ID">
                  {{ row.incoming_line_id }}
                </el-descriptions-item>
                <el-descriptions-item label="显示名称">
                  {{ row.display_name || '-' }}
                </el-descriptions-item>
                <el-descriptions-item label="手机号">
                  {{ row.phone_number || '-' }}
                </el-descriptions-item>
                <el-descriptions-item label="头像">
                  <el-image
                    v-if="row.avatar_url"
                    :src="row.avatar_url"
                    style="width: 50px; height: 50px"
                    fit="cover"
                    :preview-src-list="[row.avatar_url]"
                  />
                  <span v-else>-</span>
                </el-descriptions-item>
                <el-descriptions-item label="是否重复">
                  <el-tag :type="row.is_duplicate ? 'warning' : 'success'">
                    {{ row.is_duplicate ? '是' : '否' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="去重范围">
                  <el-tag :type="row.duplicate_scope === 'global' ? 'warning' : 'info'">
                    {{ row.duplicate_scope === 'global' ? '全局' : '当前分组' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="客户类型">
                  {{ row.customer_type || '-' }}
                </el-descriptions-item>
                <el-descriptions-item label="进线时间">
                  {{ formatDateTime(row.incoming_time) }}
                </el-descriptions-item>
                <el-descriptions-item v-if="row.line_account" label="账号信息" :span="2">
                  <div>
                    <el-tag type="info" style="margin-right: 8px">
                      {{ row.line_account.platform_type === 'line' ? 'Line' : 'Line Business' }}
                    </el-tag>
                    <span>{{ row.line_account.line_id }}</span>
                    <span v-if="row.line_account.display_name" style="margin-left: 8px; color: #909399">
                      ({{ row.line_account.display_name }})
                    </span>
                  </div>
                </el-descriptions-item>
                <el-descriptions-item v-if="row.group" label="分组信息" :span="2">
                  <div>
                    <el-tag type="primary" style="margin-right: 8px">
                      {{ row.group.activation_code }}
                    </el-tag>
                    <span v-if="row.group.remark">{{ row.group.remark }}</span>
                  </div>
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="incoming_line_id" label="进线Line ID" width="180" show-overflow-tooltip />
        <el-table-column prop="display_name" label="显示名称" width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="name-cell">
              <el-avatar
                v-if="row.avatar_url"
                :src="row.avatar_url"
                :size="30"
                style="margin-right: 8px"
              />
              <span>{{ row.display_name || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="phone_number" label="手机号" width="120" />
        <el-table-column prop="is_duplicate" label="是否重复" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_duplicate ? 'warning' : 'success'" size="small">
              {{ row.is_duplicate ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duplicate_scope" label="去重范围" width="100">
          <template #default="{ row }">
            <el-tag :type="row.duplicate_scope === 'global' ? 'warning' : 'info'" size="small">
              {{ row.duplicate_scope === 'global' ? '全局' : '当前' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="showGroupColumn" prop="group" label="分组" width="150">
          <template #default="{ row }">
            <el-tag v-if="row.group" type="primary" size="small">
              {{ row.group.activation_code }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="line_account" label="账号" width="180">
          <template #default="{ row }">
            <div v-if="row.line_account">
              <el-tag
                :type="row.line_account.platform_type === 'line' ? 'success' : 'warning'"
                size="small"
                style="margin-right: 8px"
              >
                {{ row.line_account.platform_type === 'line' ? 'Line' : 'Line Business' }}
              </el-tag>
              <span>{{ row.line_account.line_id }}</span>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="incoming_time" label="进线时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.incoming_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { getIncomingLogs, getOverviewStats } from '@/api/stats'
import { getGroups } from '@/api/group'
import { formatDateTime } from '@/utils/format'
import { useWebSocketStore } from '@/store/websocket'
import { useAuthStore } from '@/store/auth'
import dayjs from 'dayjs'

const authStore = useAuthStore()

// 数据
const loading = ref(false)
const tableData = ref([])
const groupList = ref([])
const overviewStats = ref({})
const dateRange = ref(null)

// 筛选表单
const filterForm = reactive({
  group_id: null,
  line_account_id: null,
  is_duplicate: null,
  start_time: '',
  end_time: '',
  search: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// WebSocket Store
const wsStore = useWebSocketStore()

// 是否显示分组列（子账号不显示）
const showGroupColumn = computed(() => {
  return authStore.isAdmin || authStore.user?.role === 'user'
})

// 获取分组列表
const fetchGroups = async () => {
  try {
    const res = await getGroups({ page: 1, page_size: 100 })
    if (res.code === 1000 && res.data) {
      groupList.value = res.data.list || res.data.data || []
    }
  } catch (error) {
    console.error('获取分组列表失败:', error)
  }
}

// 获取总览统计
const fetchOverviewStats = async () => {
  try {
    const res = await getOverviewStats()
    if (res.code === 1000 && res.data) {
      overviewStats.value = res.data
    }
  } catch (error) {
    console.error('获取总览统计失败:', error)
  }
}

// 获取进线日志列表
const fetchIncomingLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filterForm
    }

    // 移除空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })

    const res = await getIncomingLogs(params)
    if (res.code === 1000) {
      tableData.value = res.data.list || res.data.data || []
      pagination.total = res.data.total || 0
    } else {
      ElMessage.error(res.message || '获取进线日志列表失败')
    }
  } catch (error) {
    console.error('获取进线日志列表失败:', error)
    ElMessage.error('获取进线日志列表失败')
  } finally {
    loading.value = false
  }
}

// 时间范围变化
const handleDateRangeChange = (dates) => {
  if (dates && dates.length === 2) {
    filterForm.start_time = dates[0]
    filterForm.end_time = dates[1]
  } else {
    filterForm.start_time = ''
    filterForm.end_time = ''
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchIncomingLogs()
}

// 重置
const handleReset = () => {
  filterForm.group_id = null
  filterForm.line_account_id = null
  filterForm.is_duplicate = null
  filterForm.start_time = ''
  filterForm.end_time = ''
  filterForm.search = ''
  dateRange.value = null
  pagination.page = 1
  fetchIncomingLogs()
}

// 刷新
const handleRefresh = () => {
  fetchIncomingLogs()
  fetchOverviewStats()
}

// 分页变化
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchIncomingLogs()
}

const handlePageChange = (page) => {
  pagination.page = page
  fetchIncomingLogs()
}

// 查看详情
const handleViewDetail = (row) => {
  // 展开行
  // 这里可以通过ref来控制表格的展开状态
  ElMessage.info('点击左侧展开按钮查看详细信息')
}

// 初始化WebSocket消息处理器
const initWebSocket = () => {
  // 注册消息处理器
  wsStore.registerMessageHandler('leads-list', (message) => {
    // 处理进线更新消息
    if (message.type === 'incoming_update') {
      // 刷新列表和统计
      fetchIncomingLogs()
      fetchOverviewStats()
    } else if (message.type === 'stats_update') {
      // 更新统计
      fetchOverviewStats()
    }
  })
}

// 生命周期
onMounted(() => {
  fetchGroups()
  fetchOverviewStats()
  fetchIncomingLogs()
  initWebSocket()
})

onUnmounted(() => {
  // 取消注册消息处理器
  wsStore.unregisterMessageHandler('leads-list')
})
</script>

<style scoped lang="less">
@import '@/styles/list-page.less';

.leads-list {
  .stats-cards {
    margin-bottom: 20px;

    .stats-card {
      border-radius: 8px;
      transition: all 0.3s;
      
      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
      }

      .stats-item {
        text-align: center;

        .stats-label {
          font-size: 14px;
          color: #909399;
          margin-bottom: 8px;
        }

        .stats-value {
          font-size: 28px;
          font-weight: bold;
          color: #303133;

          &.highlight {
            color: #409eff;
          }

          &.warning {
            color: #e6a23c;
          }
        }
      }
    }
  }

  .expand-content {
    padding: 20px;
    background: #f8f9fa;
    border-radius: 8px;
  }

  .name-cell {
    display: flex;
    align-items: center;
  }
}
</style>
