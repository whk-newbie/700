<template>
  <div class="llm-call-logs">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>大模型调用记录</span>
        </div>
      </template>

      <!-- 统计卡片 -->
      <div class="stats-cards">
        <el-card class="stats-card">
          <div class="stats-item">
            <div class="stats-label">总调用次数</div>
            <div class="stats-value">{{ stats.totalCalls }}</div>
          </div>
        </el-card>
        <el-card class="stats-card">
          <div class="stats-item">
            <div class="stats-label">成功次数</div>
            <div class="stats-value success">{{ stats.successCalls }}</div>
          </div>
        </el-card>
        <el-card class="stats-card">
          <div class="stats-item">
            <div class="stats-label">失败次数</div>
            <div class="stats-value danger">{{ stats.errorCalls }}</div>
          </div>
        </el-card>
        <el-card class="stats-card">
          <div class="stats-item">
            <div class="stats-label">总Token数</div>
            <div class="stats-value">{{ formatNumber(stats.totalTokens) }}</div>
          </div>
        </el-card>
        <el-card class="stats-card">
          <div class="stats-item">
            <div class="stats-label">平均耗时(ms)</div>
            <div class="stats-value">{{ formatNumber(stats.avgDuration) }}</div>
          </div>
        </el-card>
      </div>

      <!-- 筛选区域 -->
      <div class="filter-section">
        <el-form :model="filterForm" :inline="true" class="filter-form">
          <el-form-item label="配置">
            <el-select
              v-model="filterForm.config_id"
              placeholder="全部"
              clearable
              filterable
              style="width: 200px"
            >
              <el-option
                v-for="config in configList"
                :key="config.id"
                :label="config.name"
                :value="config.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="模板">
            <el-select
              v-model="filterForm.template_id"
              placeholder="全部"
              clearable
              filterable
              style="width: 200px"
            >
              <el-option
                v-for="template in templateList"
                :key="template.id"
                :label="template.template_name"
                :value="template.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="激活码">
            <el-input
              v-model="filterForm.activation_code"
              placeholder="请输入激活码"
              clearable
              style="width: 150px"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select
              v-model="filterForm.status"
              placeholder="全部"
              clearable
              style="width: 120px"
            >
              <el-option label="成功" value="success" />
              <el-option label="失败" value="error" />
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
          <el-form-item>
            <el-button type="primary" @click="handleSearch" :loading="loading">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
            <el-button @click="handleReset">重置</el-button>
            <el-button type="default" :disabled="loading" @click="handleRefresh">
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
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="config_id" label="配置" width="150" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getConfigName(row.config_id) }}
          </template>
        </el-table-column>
        <el-table-column prop="template_id" label="模板" width="150" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.template_id ? getTemplateName(row.template_id) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="activation_code" label="激活码" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.activation_code" type="info" size="small">
              {{ row.activation_code }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'">
              {{ row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="tokens_used" label="Token数" width="120">
          <template #default="{ row }">
            {{ row.tokens_used ? formatNumber(row.tokens_used) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="duration_ms" label="耗时(ms)" width="120">
          <template #default="{ row }">
            {{ row.duration_ms ? formatNumber(row.duration_ms) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="call_time" label="调用时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.call_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination" v-if="pagination.total > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="Number(pagination.total)"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="调用记录详情"
      width="900px"
      :close-on-click-modal="false"
    >
      <div v-if="currentLog">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本信息" name="info">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="ID">
                {{ currentLog.id }}
              </el-descriptions-item>
              <el-descriptions-item label="配置">
                {{ getConfigName(currentLog.config_id) }}
              </el-descriptions-item>
              <el-descriptions-item label="模板">
                {{ currentLog.template_id ? getTemplateName(currentLog.template_id) : '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="激活码">
                <el-tag v-if="currentLog.activation_code" type="info" size="small">
                  {{ currentLog.activation_code }}
                </el-tag>
                <span v-else>-</span>
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="currentLog.status === 'success' ? 'success' : 'danger'">
                  {{ currentLog.status === 'success' ? '成功' : '失败' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="调用时间">
                {{ formatDateTime(currentLog.call_time) }}
              </el-descriptions-item>
              <el-descriptions-item label="耗时(ms)">
                {{ currentLog.duration_ms ? formatNumber(currentLog.duration_ms) : '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="错误信息" v-if="currentLog.status === 'error'">
                <el-text type="danger">{{ currentLog.error_message || '-' }}</el-text>
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane label="Token统计" name="tokens">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="总Token数">
                {{ currentLog.tokens_used ? formatNumber(currentLog.tokens_used) : '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="Prompt Token">
                {{ currentLog.prompt_tokens ? formatNumber(currentLog.prompt_tokens) : '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="Completion Token">
                {{ currentLog.completion_tokens ? formatNumber(currentLog.completion_tokens) : '-' }}
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane label="请求消息" name="request">
            <el-input
              v-model="requestMessagesText"
              type="textarea"
              :rows="10"
              readonly
            />
          </el-tab-pane>
          <el-tab-pane label="请求参数" name="params">
            <el-input
              v-model="requestParamsText"
              type="textarea"
              :rows="10"
              readonly
            />
          </el-tab-pane>
          <el-tab-pane label="响应内容" name="response">
            <el-input
              v-model="currentLog.response_content"
              type="textarea"
              :rows="10"
              readonly
            />
          </el-tab-pane>
          <el-tab-pane label="响应数据" name="responseData">
            <el-input
              v-model="responseDataText"
              type="textarea"
              :rows="10"
              readonly
            />
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getLLMCallLogs, getLLMConfigs, getLLMTemplates } from '@/api/llm'
import { formatDateTime, formatNumber } from '@/utils/format'

// 数据
const loading = ref(false)
const tableData = ref([])
const configList = ref([])
const templateList = ref([])
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 统计
const stats = reactive({
  totalCalls: 0,
  successCalls: 0,
  errorCalls: 0,
  totalTokens: 0,
  avgDuration: 0
})

// 筛选表单
const filterForm = reactive({
  config_id: null,
  template_id: null,
  activation_code: '',
  status: '',
  start_time: '',
  end_time: ''
})
const dateRange = ref(null)

// 详情对话框
const detailDialogVisible = ref(false)
const activeTab = ref('info')
const currentLog = ref(null)
const requestMessagesText = ref('')
const requestParamsText = ref('')
const responseDataText = ref('')

// 获取配置名称
const getConfigName = (configId) => {
  if (!configId) return '-'
  const config = configList.value.find(c => c.id === configId)
  return config ? config.name : `配置ID: ${configId}`
}

// 获取模板名称
const getTemplateName = (templateId) => {
  if (!templateId) return '-'
  const template = templateList.value.find(t => t.id === templateId)
  return template ? template.template_name : `模板ID: ${templateId}`
}

// 获取配置列表
const fetchConfigs = async () => {
  try {
    const res = await getLLMConfigs({ page: 1, page_size: 100 })
    if (res.code === 1000) {
      configList.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取配置列表失败:', error)
  }
}

// 获取模板列表
const fetchTemplates = async () => {
  try {
    const res = await getLLMTemplates({ page: 1, page_size: 100 })
    if (res.code === 1000) {
      templateList.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取模板列表失败:', error)
  }
}

// 获取调用记录列表
const fetchLogs = async () => {
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

    const res = await getLLMCallLogs(params)
    if (res.code === 1000) {
      tableData.value = res.data.list || []
      pagination.total = res.data.pagination?.total || 0
      
      // 计算统计信息
      calculateStats()
    } else {
      ElMessage.error(res.message || '获取调用记录失败')
    }
  } catch (error) {
    console.error('获取调用记录失败:', error)
    ElMessage.error('获取调用记录失败')
  } finally {
    loading.value = false
  }
}

// 计算统计信息
const calculateStats = () => {
  let totalCalls = 0
  let successCalls = 0
  let errorCalls = 0
  let totalTokens = 0
  let totalDuration = 0
  let durationCount = 0

  tableData.value.forEach(log => {
    totalCalls++
    if (log.status === 'success') {
      successCalls++
    } else {
      errorCalls++
    }
    if (log.tokens_used) {
      totalTokens += log.tokens_used
    }
    if (log.duration_ms) {
      totalDuration += log.duration_ms
      durationCount++
    }
  })

  stats.totalCalls = totalCalls
  stats.successCalls = successCalls
  stats.errorCalls = errorCalls
  stats.totalTokens = totalTokens
  stats.avgDuration = durationCount > 0 ? Math.round(totalDuration / durationCount) : 0
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchLogs()
}

// 重置
const handleReset = () => {
  filterForm.config_id = null
  filterForm.template_id = null
  filterForm.activation_code = ''
  filterForm.status = ''
  filterForm.start_time = ''
  filterForm.end_time = ''
  dateRange.value = null
  pagination.page = 1
  fetchLogs()
}

// 刷新
const handleRefresh = () => {
  fetchLogs()
}

// 分页变化
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchLogs()
}

const handlePageChange = (page) => {
  pagination.page = page
  fetchLogs()
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

// 查看详情
const handleViewDetail = (row) => {
  currentLog.value = row
  activeTab.value = 'info'
  
  // 格式化请求消息
  try {
    if (row.request_messages && Array.isArray(row.request_messages)) {
      requestMessagesText.value = JSON.stringify(row.request_messages, null, 2)
    } else {
      requestMessagesText.value = row.request_messages ? String(row.request_messages) : '-'
    }
  } catch {
    requestMessagesText.value = '-'
  }
  
  // 格式化请求参数
  try {
    if (row.request_params && typeof row.request_params === 'object') {
      requestParamsText.value = JSON.stringify(row.request_params, null, 2)
    } else {
      requestParamsText.value = row.request_params ? String(row.request_params) : '-'
    }
  } catch {
    requestParamsText.value = '-'
  }
  
  // 格式化响应数据
  try {
    if (row.response_data && typeof row.response_data === 'object') {
      responseDataText.value = JSON.stringify(row.response_data, null, 2)
    } else {
      responseDataText.value = row.response_data ? String(row.response_data) : '-'
    }
  } catch {
    responseDataText.value = '-'
  }
  
  detailDialogVisible.value = true
}

// 初始化
onMounted(() => {
  fetchConfigs()
  fetchTemplates()
  fetchLogs()
})
</script>

<style scoped lang="less">
.llm-call-logs {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-actions {
    display: flex;
    gap: 10px;
  }

  .stats-cards {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 20px;
    margin-bottom: 20px;

    .stats-card {
      .stats-item {
        text-align: center;

        .stats-label {
          font-size: 14px;
          color: #909399;
          margin-bottom: 10px;
        }

        .stats-value {
          font-size: 24px;
          font-weight: bold;
          color: #409eff;

          &.success {
            color: #67c23a;
          }

          &.danger {
            color: #f56c6c;
          }
        }
      }
    }
  }

  .filter-section {
    margin-bottom: 20px;
    padding: 20px;
    background-color: #f5f7fa;
    border-radius: 4px;
  }

  :deep(.el-table) {
    min-height: 400px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: center;
  }
}
</style>
