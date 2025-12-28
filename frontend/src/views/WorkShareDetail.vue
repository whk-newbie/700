<template>
  <div class="work-share-detail">
    <!-- 密码验证失败才显示密码输入对话框 -->
    <div v-if="showPasswordDialog" class="password-container">
      <el-card class="password-card" shadow="always">
        <template #header>
          <div class="card-title">
            <el-icon :size="24" color="#409eff"><Lock /></el-icon>
            <span>请输入访问密码</span>
          </div>
        </template>
        
        <div class="password-form">
          <el-alert
            v-if="shareInfo.remark"
            :title="shareInfo.remark"
            type="info"
            :closable="false"
            style="margin-bottom: 20px"
          >
            <template v-if="shareInfo.description">
              {{ shareInfo.description }}
            </template>
          </el-alert>

          <el-alert
            type="warning"
            :closable="false"
            style="margin-bottom: 20px"
          >
            密码验证失败，请重新输入正确的密码
          </el-alert>

          <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef">
            <el-form-item prop="password">
              <el-input
                v-model="passwordForm.password"
                type="password"
                placeholder="请输入访问密码"
                size="large"
                show-password
                @keyup.enter="handleVerifyPassword"
                :disabled="verifying"
              >
                <template #prefix>
                  <el-icon><Lock /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            
            <el-form-item>
              <el-button
                type="primary"
                size="large"
                style="width: 100%"
                @click="handleVerifyPassword"
                :loading="verifying"
              >
                {{ verifying ? '验证中...' : '重新验证' }}
              </el-button>
            </el-form-item>
          </el-form>

          <div class="tips">
            <el-icon><InfoFilled /></el-icon>
            <span>提示：请向分享者获取正确的访问密码</span>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 验证成功后显示内容 -->
    <div v-else-if="verified">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <el-icon class="is-loading" :size="40">
          <Loading />
        </el-icon>
        <p>加载中...</p>
      </div>

      <!-- 错误状态 -->
      <div v-else-if="error" class="error-container">
        <el-result icon="error" title="加载失败" :sub-title="error">
          <template #extra>
            <el-button type="primary" @click="loadAccounts">重试</el-button>
          </template>
        </el-result>
      </div>

      <!-- 正常显示 -->
      <div v-else class="content-container">
        <!-- 顶部信息栏 -->
        <el-card class="header-card" shadow="never">
          <div class="header-info">
            <div class="group-info">
              <h2>{{ shareInfo.remark || '分组详情' }}</h2>
              <p class="description" v-if="shareInfo.description">{{ shareInfo.description }}</p>
              <div class="meta-info">
                <el-tag type="info">激活码: {{ shareInfo.activation_code }}</el-tag>
                <el-tag type="success">浏览次数: {{ shareInfo.view_count }}</el-tag>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 账号列表 -->
        <el-card class="account-card" shadow="never">
          <template #header>
            <div class="card-header">
              <span>账号列表</span>
              <el-tag type="primary">共 {{ tableData.length }} 个账号</el-tag>
            </div>
          </template>

          <!-- 筛选区域 -->
          <div class="filter-section">
            <el-form :model="filterForm" :inline="true" class="filter-form">
              <el-form-item label="平台">
                <el-select
                  v-model="filterForm.platform_type"
                  placeholder="全部"
                  clearable
                  style="width: 120px"
                >
                  <el-option label="Line" value="line" />
                  <el-option label="Line Business" value="line_business" />
                </el-select>
              </el-form-item>
              <el-form-item label="在线状态">
                <el-select
                  v-model="filterForm.online_status"
                  placeholder="全部"
                  clearable
                  style="width: 140px"
                >
                  <el-option label="在线" value="online" />
                  <el-option label="离线" value="offline" />
                  <el-option label="用户登出" value="user_logout" />
                  <el-option label="异常离线" value="abnormal_offline" />
                </el-select>
              </el-form-item>
              <el-form-item label="搜索">
                <el-input
                  v-model="filterForm.search"
                  placeholder="Line ID或显示名称"
                  clearable
                  style="width: 200px"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleSearch">
                  <el-icon><Search /></el-icon>
                  搜索
                </el-button>
                <el-button @click="handleReset">重置</el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- 数据表格 -->
          <el-table
            v-loading="tableLoading"
            :data="filteredTableData"
            style="width: 100%"
            stripe
            empty-text="暂无数据"
          >
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="line_id" label="Line ID" width="180" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tag type="info" size="small">{{ row.line_id }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="display_name" label="显示名称" width="150" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.display_name || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="platform_type" label="平台" width="120">
              <template #default="{ row }">
                <el-tag :type="row.platform_type === 'line' ? 'primary' : 'success'">
                  {{ row.platform_type === 'line' ? 'Line' : 'Line Business' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="online_status" label="在线状态" width="120">
              <template #default="{ row }">
                <el-tag
                  :type="
                    row.online_status === 'online'
                      ? 'success'
                      : row.online_status === 'abnormal_offline'
                      ? 'danger'
                      : 'info'
                  "
                >
                  {{
                    row.online_status === 'online'
                      ? '在线'
                      : row.online_status === 'offline'
                      ? '离线'
                      : row.online_status === 'user_logout'
                      ? '用户登出'
                      : '异常离线'
                  }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="进线统计" width="180">
              <template #default="{ row }">
                <div class="stats-info">
                  <div>今日: <strong style="color: #409eff">{{ row.today_incoming ?? 0 }}</strong></div>
                  <div>总计: <strong>{{ row.total_incoming ?? 0 }}</strong></div>
                  <div>重复: <strong style="color: #e6a23c">{{ row.duplicate_incoming ?? 0 }}</strong></div>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="last_active_at" label="最后活跃" width="180">
              <template #default="{ row }">
                {{ row.last_active_at ? formatDateTime(row.last_active_at) : '-' }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>
    </div>

    <!-- 初始加载状态（自动验证中） -->
    <div v-else class="loading-container">
      <el-icon class="is-loading" :size="40">
        <Loading />
      </el-icon>
      <p>正在验证访问权限...</p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading, Search, Lock, InfoFilled } from '@element-plus/icons-vue'
import { getShareInfo, verifySharePassword } from '@/api/share'
import { getLineAccounts } from '@/api/lineAccount'
import { formatDateTime } from '@/utils/format'

const route = useRoute()

// 状态
const verified = ref(false) // 是否已验证密码
const showPasswordDialog = ref(false) // 是否显示密码输入对话框（只在验证失败时显示）
const verifying = ref(false) // 验证中
const loading = ref(false)
const tableLoading = ref(false)
const error = ref('')
const shareInfo = ref({})
const tableData = ref([])
const wsConnection = ref(null)

// 密码表单
const passwordFormRef = ref(null)
const passwordForm = reactive({
  password: ''
})

const passwordRules = {
  password: [
    { required: true, message: '请输入访问密码', trigger: 'blur' }
  ]
}

// 筛选表单
const filterForm = reactive({
  platform_type: '',
  online_status: '',
  search: ''
})

// 过滤后的表格数据
const filteredTableData = computed(() => {
  let data = tableData.value

  // 平台筛选
  if (filterForm.platform_type) {
    data = data.filter(item => item.platform_type === filterForm.platform_type)
  }

  // 在线状态筛选
  if (filterForm.online_status) {
    data = data.filter(item => item.online_status === filterForm.online_status)
  }

  // 搜索筛选
  if (filterForm.search) {
    const searchLower = filterForm.search.toLowerCase()
    data = data.filter(item => {
      return (
        item.line_id?.toLowerCase().includes(searchLower) ||
        item.display_name?.toLowerCase().includes(searchLower)
      )
    })
  }

  return data
})

// 加载分享基本信息并自动验证
const loadShareInfo = async () => {
  const code = route.query.code
  if (!code) {
    error.value = '缺少分享码参数'
    showPasswordDialog.value = true
    return
  }

  try {
    // 先获取基本信息
    const res = await getShareInfo(code)
    if (res.code === 1000) {
      shareInfo.value = res.data
      // 默认密码就是分享码，自动填充
      passwordForm.password = code
      // 自动验证密码
      await autoVerifyPassword()
    } else {
      ElMessage.error(res.message || '获取分享信息失败')
      showPasswordDialog.value = true
    }
  } catch (err) {
    console.error('加载分享信息失败:', err)
    ElMessage.error('加载失败，请检查分享链接是否正确')
    showPasswordDialog.value = true
  }
}

// 自动验证密码（页面加载时）
const autoVerifyPassword = async () => {
  const code = route.query.code
  
  try {
    const res = await verifySharePassword(code, passwordForm.password)
    if (res.code === 1000) {
      shareInfo.value = res.data
      verified.value = true
      
      // 保存分享 token 到 localStorage
      if (res.data.share_token) {
        localStorage.setItem('share_token', res.data.share_token)
      }
      
      // 验证成功后加载账号列表
      await loadAccounts()
      
      // 初始化 WebSocket
      initWebSocket(code)
    } else {
      // 密码错误，显示密码输入对话框
      showPasswordDialog.value = true
      ElMessage.warning('密码验证失败，请输入正确的密码')
    }
  } catch (err) {
    console.error('自动验证密码失败:', err)
    // 验证失败，显示密码输入对话框
    showPasswordDialog.value = true
  }
}

// 手动验证密码（用户在对话框中输入后点击按钮）
const handleVerifyPassword = async () => {
  if (!passwordFormRef.value) return

  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return

    const code = route.query.code
    verifying.value = true

    try {
      const res = await verifySharePassword(code, passwordForm.password)
      if (res.code === 1000) {
        shareInfo.value = res.data
        verified.value = true
        showPasswordDialog.value = false // 隐藏密码对话框
        ElMessage.success('验证成功')
        
        // 保存分享 token 到 localStorage
        if (res.data.share_token) {
          localStorage.setItem('share_token', res.data.share_token)
        }
        
        // 验证成功后加载账号列表
        await loadAccounts()
        
        // 初始化 WebSocket
        initWebSocket(code)
      } else {
        ElMessage.error(res.message || '密码错误')
      }
    } catch (err) {
      console.error('验证密码失败:', err)
      ElMessage.error(err.response?.data?.message || '密码错误')
    } finally {
      verifying.value = false
    }
  })
}

// 加载账号列表
const loadAccounts = async () => {
  if (!shareInfo.value.group_id) return

  tableLoading.value = true
  error.value = ''
  
  try {
    const res = await getLineAccounts({
      group_id: shareInfo.value.group_id,
      page: 1,
      page_size: 50 // 每页50条
    })
    if (res.code === 1000) {
      tableData.value = Array.isArray(res.data?.list) ? res.data.list : []
      loading.value = false
    } else {
      error.value = res.message || '加载账号列表失败'
    }
  } catch (err) {
    console.error('加载账号列表失败:', err)
    error.value = '加载账号列表失败'
  } finally {
    tableLoading.value = false
  }
}

// 搜索
const handleSearch = () => {
  // 触发计算属性重新计算
}

// 重置筛选
const handleReset = () => {
  filterForm.platform_type = ''
  filterForm.online_status = ''
  filterForm.search = ''
}

// 初始化 WebSocket 连接（使用分享码，不使用 token）
const initWebSocket = (code) => {
  // 构建 WebSocket URL
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = import.meta.env.VITE_WS_BASE_URL || window.location.host
  const wsUrl = `${protocol}//${host}/api/ws/share?code=${code}`

  console.log('连接 WebSocket:', wsUrl)

  try {
    wsConnection.value = new WebSocket(wsUrl)

    wsConnection.value.onopen = () => {
      console.log('WebSocket 连接成功')
    }

    wsConnection.value.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        handleWebSocketMessage(message)
      } catch (err) {
        console.error('解析 WebSocket 消息失败:', err)
      }
    }

    wsConnection.value.onerror = (error) => {
      console.error('WebSocket 错误:', error)
    }

    wsConnection.value.onclose = () => {
      console.log('WebSocket 连接关闭')
      // 5秒后重连（仅在已验证的情况下）
      if (verified.value) {
        setTimeout(() => {
          initWebSocket(code)
        }, 5000)
      }
    }

    // 定时发送心跳
    const heartbeatInterval = setInterval(() => {
      if (wsConnection.value?.readyState === WebSocket.OPEN) {
        wsConnection.value.send(JSON.stringify({
          type: 'heartbeat',
          timestamp: Date.now()
        }))
      }
    }, 30000) // 30秒一次心跳

    // 清理定时器
    onUnmounted(() => {
      clearInterval(heartbeatInterval)
    })
  } catch (err) {
    console.error('创建 WebSocket 连接失败:', err)
  }
}

// 处理 WebSocket 消息
const handleWebSocketMessage = (message) => {
  console.log('收到 WebSocket 消息:', message)

  switch (message.type) {
    case 'connected':
      console.log('WebSocket 已连接')
      break
    case 'account_status_change':
      handleAccountStatusChange(message.data)
      break
    case 'account_stats_update':
      handleAccountStatsUpdate(message.data)
      break
    default:
      console.log('未处理的消息类型:', message.type)
  }
}

// 处理账号状态变化
const handleAccountStatusChange = (data) => {
  const { line_account_id, online_status } = data
  const account = tableData.value.find(item => item.line_id === line_account_id)
  if (account) {
    account.online_status = online_status
    if (online_status === 'online') {
      account.last_active_at = new Date().toISOString()
    }
  }
}

// 处理账号统计更新
const handleAccountStatsUpdate = (data) => {
  const { line_id, total_incoming, today_incoming, duplicate_incoming, today_duplicate } = data
  const account = tableData.value.find(item => item.line_id === line_id)
  if (account) {
    account.total_incoming = total_incoming
    account.today_incoming = today_incoming
    account.duplicate_incoming = duplicate_incoming
    account.today_duplicate = today_duplicate
  }
}

// 初始化
onMounted(() => {
  loadShareInfo()
})

// 清理
onUnmounted(() => {
  if (wsConnection.value) {
    wsConnection.value.close()
  }
})
</script>

<style scoped lang="less">
.work-share-detail {
  min-height: 100vh;
  background: #f0f2f5;

  .password-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: 20px;

    .password-card {
      width: 100%;
      max-width: 450px;

      .card-title {
        display: flex;
        align-items: center;
        gap: 10px;
        font-size: 18px;
        font-weight: 500;
      }

      .password-form {
        padding: 20px 0;

        .tips {
          display: flex;
          align-items: center;
          gap: 8px;
          color: #909399;
          font-size: 14px;
          margin-top: 10px;
        }
      }
    }
  }

  .loading-container,
  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    gap: 20px;

    p {
      font-size: 16px;
      color: #909399;
    }
  }

  .content-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 20px;

    .header-card {
      margin-bottom: 20px;

      .header-info {
        .group-info {
          h2 {
            margin: 0 0 10px 0;
            font-size: 24px;
            color: #303133;
          }

          .description {
            margin: 0 0 15px 0;
            color: #606266;
            font-size: 14px;
            line-height: 1.5;
          }

          .meta-info {
            display: flex;
            gap: 10px;
          }
        }
      }
    }

    .account-card {
      .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;

        span {
          font-size: 16px;
          font-weight: 500;
        }
      }

      .filter-section {
        margin-bottom: 20px;
        padding-bottom: 20px;
        border-bottom: 1px solid #ebeef5;

        .filter-form {
          margin: 0;
        }
      }

      .stats-info {
        font-size: 12px;
        line-height: 1.6;

        div {
          margin: 2px 0;
        }

        strong {
          font-weight: 600;
        }
      }
    }
  }
}
</style>
