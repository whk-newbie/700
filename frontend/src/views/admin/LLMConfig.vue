<template>
  <div class="llm-config-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>OpenAI API 配置</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- API配置标签页 -->
        <el-tab-pane label="API配置" name="config">
          <div class="config-content">
            <el-alert
              title="API Key 管理说明"
              description="OpenAI API Key 将使用RSA加密存储在服务器端，确保安全性。API Key 仅用于转发用户请求到OpenAI API。"
              type="info"
              show-icon
              :closable="false"
              style="margin-bottom: 24px;"
            />

            <!-- 当前状态显示 -->
            <div class="status-section">
              <h3>当前配置状态</h3>
              <div class="status-info">
                <el-row :gutter="20">
                  <el-col :span="12">
                    <div class="status-item">
                      <span class="label">API Key 状态:</span>
                      <el-tag :type="apiKeyStatus.hasKey ? 'success' : 'danger'" size="large">
                        {{ apiKeyStatus.hasKey ? '已配置' : '未配置' }}
                      </el-tag>
                    </div>
                  </el-col>
                  <el-col :span="12">
                    <div class="status-item">
                      <span class="label">最后更新时间:</span>
                      <span class="value">{{ apiKeyStatus.updatedAt || '从未配置' }}</span>
                    </div>
                  </el-col>
                </el-row>
              </div>
            </div>

            <!-- 配置表单 -->
            <div class="config-section">
              <h3>更新 API Key</h3>
              <el-form
                ref="configFormRef"
                :model="configFormData"
                :rules="configFormRules"
                label-width="140px"
              >
                <el-form-item label="OpenAI API Key" prop="api_key">
                  <el-input
                    v-model="configFormData.api_key"
                    type="password"
                    placeholder="请输入新的API Key"
                    show-password
                    style="width: 400px"
                  />
                  <div class="form-tip">
                    API Key 将在前端使用RSA公钥加密后传输，确保安全性
                  </div>
                </el-form-item>
                <el-form-item>
                  <el-button
                    type="primary"
                    @click="handleSubmit"
                    :loading="saving"
                    :disabled="!configFormData.api_key"
                  >
                    更新配置
                  </el-button>
                  <el-button @click="handleRefresh" :loading="loading">
                    刷新状态
                  </el-button>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </el-tab-pane>

        <!-- 调用日志标签页 -->
        <el-tab-pane label="调用日志" name="logs">
          <!-- 筛选区域 -->
          <div class="filter-section">
            <el-form :model="logFilterForm" :inline="true" class="filter-form">
              <el-form-item label="状态">
                <el-select
                  v-model="logFilterForm.status"
                  placeholder="全部"
                  clearable
                  style="width: 120px"
                >
                  <el-option label="成功" value="success" />
                  <el-option label="失败" value="error" />
                </el-select>
              </el-form-item>
              <el-form-item label="激活码">
                <el-input
                  v-model="logFilterForm.activation_code"
                  placeholder="激活码"
                  clearable
                  style="width: 150px"
                />
              </el-form-item>
              <el-form-item label="开始时间">
                <el-date-picker
                  v-model="logFilterForm.start_time"
                  type="datetime"
                  placeholder="选择开始时间"
                  format="YYYY-MM-DD HH:mm:ss"
                  value-format="YYYY-MM-DD HH:mm:ss"
                  style="width: 200px"
                />
              </el-form-item>
              <el-form-item label="结束时间">
                <el-date-picker
                  v-model="logFilterForm.end_time"
                  type="datetime"
                  placeholder="选择结束时间"
                  format="YYYY-MM-DD HH:mm:ss"
                  value-format="YYYY-MM-DD HH:mm:ss"
                  style="width: 200px"
                />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleLogSearch" :loading="logLoading">
                  <el-icon><Search /></el-icon>
                  搜索
                </el-button>
                <el-button @click="handleLogReset">重置</el-button>
                <el-button type="default" :disabled="logLoading" @click="handleLogRefresh">
                  <el-icon><Refresh /></el-icon>
                  刷新
                </el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- 日志列表表格 -->
          <el-table
            v-loading="logLoading"
            :data="logTableData"
            style="width: 100%"
            stripe
            empty-text="暂无数据"
          >
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="activation_code" label="激活码" width="150" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : 'danger'">
                  {{ row.status === 'success' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="prompt_tokens" label="请求Tokens" width="120" />
            <el-table-column prop="completion_tokens" label="响应Tokens" width="120" />
            <el-table-column prop="tokens_used" label="总Tokens" width="120" />
            <el-table-column prop="duration_ms" label="耗时(ms)" width="120" />
            <el-table-column prop="call_time" label="调用时间" width="180">
              <template #default="{ row }">
                {{ formatDateTime(row.call_time) }}
              </template>
            </el-table-column>
            <el-table-column prop="error_message" label="错误信息" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.error_message || '-' }}
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="pagination" v-if="logPagination.total > 0">
            <el-pagination
              v-model:current-page="logPagination.page"
              v-model:page-size="logPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="Number(logPagination.total)"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleLogSizeChange"
              @current-change="handleLogPageChange"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getOpenAIAPIKey, updateOpenAIAPIKey, getLLMCallLogs } from '@/api/llm'
import { formatDateTime } from '@/utils/format'

// 标签页
const activeTab = ref('config')

// API配置相关
const loading = ref(false)
const saving = ref(false)
const apiKeyStatus = reactive({
  hasKey: false,
  updatedAt: ''
})

// 配置表单相关
const configFormRef = ref(null)
const configFormData = reactive({
  api_key: ''
})

const configFormRules = {
  api_key: [
    { required: true, message: '请输入OpenAI API Key', trigger: 'blur' }
  ]
}

// 调用日志相关
const logLoading = ref(false)
const logTableData = ref([])
const logPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})
const logFilterForm = reactive({
  status: '',
  activation_code: '',
  start_time: '',
  end_time: ''
})

// 获取API Key状态
const fetchAPIKeyStatus = async () => {
  loading.value = true
  try {
    const res = await getOpenAIAPIKey()
    if (res.code === 1000) {
      apiKeyStatus.hasKey = res.data.has_key || false
      apiKeyStatus.updatedAt = res.data.updated_at || ''
    } else {
      ElMessage.error(res.message || '获取API Key状态失败')
    }
  } catch (error) {
    console.error('获取API Key状态失败:', error)
    ElMessage.error('获取API Key状态失败')
  } finally {
    loading.value = false
  }
}

// 更新API Key
const handleSubmit = async () => {
  if (!configFormRef.value) return

  await configFormRef.value.validate(async (valid) => {
    if (!valid) return

    saving.value = true
    try {
      const res = await updateOpenAIAPIKey(configFormData.api_key)
      if (res.code === 1000) {
        ElMessage.success('API Key 更新成功')
        configFormData.api_key = '' // 清空表单
        if (configFormRef.value) {
          configFormRef.value.clearValidate()
        }
        // 刷新状态
        await fetchAPIKeyStatus()
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } catch (error) {
      console.error('更新API Key失败:', error)
      ElMessage.error('更新失败')
    } finally {
      saving.value = false
    }
  })
}

// 刷新状态
const handleRefresh = () => {
  fetchAPIKeyStatus()
}

// 获取调用日志列表
const fetchLogs = async () => {
  logLoading.value = true
  try {
    const params = {
      page: logPagination.page,
      page_size: logPagination.pageSize,
      ...logFilterForm
    }
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })

    const res = await getLLMCallLogs(params)
    if (res.code === 1000) {
      logTableData.value = res.data.list || []
      logPagination.total = res.data.pagination?.total || 0
    } else {
      ElMessage.error(res.message || '获取调用日志失败')
    }
  } catch (error) {
    console.error('获取调用日志失败:', error)
    ElMessage.error('获取调用日志失败')
  } finally {
    logLoading.value = false
  }
}

// 标签页切换
const handleTabChange = (tabName) => {
  if (tabName === 'logs') {
    fetchLogs()
  }
}

// 日志相关操作
const handleLogSearch = () => {
  logPagination.page = 1
  fetchLogs()
}

const handleLogReset = () => {
  logFilterForm.status = ''
  logFilterForm.activation_code = ''
  logFilterForm.start_time = ''
  logFilterForm.end_time = ''
  logPagination.page = 1
  fetchLogs()
}

const handleLogRefresh = () => {
  fetchLogs()
}

const handleLogSizeChange = (size) => {
  logPagination.pageSize = size
  logPagination.page = 1
  fetchLogs()
}

const handleLogPageChange = (page) => {
  logPagination.page = page
  fetchLogs()
}

// 初始化
onMounted(() => {
  fetchAPIKeyStatus()
})
</script>

<style scoped lang="less">
.llm-config-container {
  .card-header {
    .header-actions {
      display: flex;
      gap: 10px;
    }
  }

  .config-content {
    .status-section {
      margin-bottom: 32px;

      h3 {
        margin: 0 0 16px 0;
        color: #303133;
        font-size: 16px;
        font-weight: 500;
      }

      .status-info {
        .status-item {
          display: flex;
          align-items: center;
          margin-bottom: 12px;

          .label {
            font-weight: 500;
            color: #606266;
            margin-right: 12px;
            min-width: 120px;
          }

          .value {
            color: #303133;
          }
        }
      }
    }

    .config-section {
      h3 {
        margin: 0 0 16px 0;
        color: #303133;
        font-size: 16px;
        font-weight: 500;
      }

      .form-tip {
        font-size: 12px;
        color: #909399;
        margin-top: 4px;
      }
    }
  }

  .filter-section {
    margin-bottom: 20px;

    .filter-form {
      .el-form-item {
        margin-bottom: 12px;
      }
    }
  }

  .pagination {
    margin-top: 20px;
    text-align: right;
  }
}
</style>
