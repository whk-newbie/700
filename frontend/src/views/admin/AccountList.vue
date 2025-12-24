<template>
  <div class="list-page-container account-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>账号列表</span>
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
          </el-form-item>
        </el-form>
      </div>

      <!-- 操作按钮区域 -->
      <div class="action-buttons">
        <el-button type="primary" :disabled="loading" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          新增账号
        </el-button>
        <el-button
          type="danger"
          :disabled="loading || selectedRows.length === 0"
          @click="handleBatchDelete"
        >
          批量移除 ({{ selectedRows.length }})
        </el-button>
        <el-button
          type="warning"
          :disabled="loading || selectedRows.length === 0"
          @click="handleBatchOffline"
        >
          强制下线 ({{ selectedRows.length }})
        </el-button>
      </div>

      <!-- 数据表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        @selection-change="handleSelectionChange"
        stripe
        empty-text="暂无数据"
      >
        <el-table-column type="selection" width="55" />
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
        <el-table-column prop="activation_code" label="激活码" width="120">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.activation_code }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="添加好友链接" width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="friend-link-cell" v-if="row.add_friend_link">
              <el-link
                :href="row.add_friend_link"
                target="_blank"
                type="primary"
                :underline="false"
                style="margin-right: 8px"
              >
                {{ row.add_friend_link }}
              </el-link>
              <el-button
                type="text"
                size="small"
                @click="handleCopyLink(row.add_friend_link)"
                title="复制链接"
              >
                <el-icon><DocumentCopy /></el-icon>
              </el-button>
            </div>
            <span v-else style="color: #909399">-</span>
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
        <el-table-column prop="reset_time" label="重置时间" width="120">
          <template #default="{ row }">
            <el-tag 
              :type="row.reset_time ? 'success' : 'info'" 
              size="small"
            >
              {{ row.reset_time || '使用分组' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_active_at" label="最后活跃" width="180">
          <template #default="{ row }">
            {{ row.last_active_at ? formatDateTime(row.last_active_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ row.created_at ? formatDateTime(row.created_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewQR(row)">
              二维码
            </el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="info" link size="small" @click="handleViewLogs(row)">
              日志
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
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

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item label="所属分组" prop="group_id">
          <el-select
            v-model="formData.group_id"
            placeholder="请选择分组"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="group in groupList"
              :key="group.id"
              :label="`${group.activation_code}${group.remark ? ' - ' + group.remark : ''}`"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="平台类型" prop="platform_type">
          <el-radio-group v-model="formData.platform_type">
            <el-radio label="line">Line</el-radio>
            <el-radio label="line_business">Line Business</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Line ID" prop="line_id">
          <el-input
            v-model="formData.line_id"
            placeholder="请输入Line ID"
          />
        </el-form-item>
        <el-form-item label="添加好友链接">
          <el-input
            v-model="formData.add_friend_link"
            placeholder="请输入添加好友链接（如：https://line.me/ti/p/~U1234567890abcdef）"
          />
        </el-form-item>
        <el-form-item label="显示名称">
          <el-input v-model="formData.display_name" placeholder="请输入显示名称" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="formData.phone_number" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="Profile URL">
          <el-input v-model="formData.profile_url" placeholder="请输入Profile URL" />
        </el-form-item>
        <el-form-item label="头像URL">
          <el-input v-model="formData.avatar_url" placeholder="请输入头像URL" />
        </el-form-item>
        <el-form-item label="个人简介">
          <el-input
            v-model="formData.bio"
            type="textarea"
            :rows="2"
            placeholder="请输入个人简介"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="状态消息">
          <el-input
            v-model="formData.status_message"
            type="textarea"
            :rows="2"
            placeholder="请输入状态消息"
            maxlength="255"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="账号备注">
          <el-input
            v-model="formData.account_remark"
            type="textarea"
            :rows="2"
            placeholder="请输入账号备注"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="重置时间">
          <el-time-picker
            v-model="formData.reset_time"
            format="HH:mm:ss"
            value-format="HH:mm:ss"
            placeholder="选择重置时间（留空则使用分组的重置时间）"
            style="width: 100%"
            clearable
          />
          <div style="font-size: 12px; color: #909399; margin-top: 4px">
            留空则使用所属分组的重置时间
          </div>
        </el-form-item>
        <el-form-item v-if="formData.id" label="在线状态">
          <el-select v-model="formData.online_status" style="width: 100%">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="用户登出" value="user_logout" />
            <el-option label="异常离线" value="abnormal_offline" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 二维码预览对话框 -->
    <el-dialog
      v-model="qrDialogVisible"
      title="二维码预览"
      width="400px"
      align-center
    >
      <div class="qr-preview">
        <div v-if="currentQRCode" class="qr-image">
          <img :src="qrImageUrl" alt="二维码" style="max-width: 100%; height: auto" />
        </div>
        <div v-else class="qr-placeholder">
          <el-empty description="暂无二维码" />
        </div>
      </div>
      <template #footer>
        <el-button @click="qrDialogVisible = false">关闭</el-button>
        <el-button 
          v-if="currentQRCode" 
          type="danger" 
          @click="handleDeleteQR" 
          :loading="deletingQR"
        >
          删除二维码
        </el-button>
        <el-button 
          v-if="currentQRCode" 
          type="warning" 
          @click="handleRegenerateQR" 
          :loading="generatingQR"
        >
          重新生成
        </el-button>
        <el-button v-if="currentQRCode" type="primary" @click="handleDownloadQR">
          下载
        </el-button>
        <el-button v-if="!currentQRCode" type="primary" @click="handleGenerateQR" :loading="generatingQR">
          生成二维码
        </el-button>
      </template>
    </el-dialog>

    <!-- 日志记录对话框 -->
    <el-dialog
      v-model="logDialogVisible"
      title="账号日志记录"
      width="800px"
    >
      <div class="log-content">
        <el-empty v-if="!logData.length" description="暂无日志记录" />
        <el-timeline v-else>
          <el-timeline-item
            v-for="(log, index) in logData"
            :key="index"
            :timestamp="log.timestamp"
            placement="top"
          >
            <el-card>
              <h4>{{ log.title }}</h4>
              <p>{{ log.content }}</p>
            </el-card>
          </el-timeline-item>
        </el-timeline>
      </div>
      <template #footer>
        <el-button @click="logDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, DocumentCopy } from '@element-plus/icons-vue'
import {
  getLineAccounts,
  createLineAccount,
  updateLineAccount,
  deleteLineAccount,
  generateQRCode,
  batchDeleteLineAccounts,
  batchUpdateLineAccounts
} from '@/api/lineAccount'
import { getGroups } from '@/api/group'
import { formatDateTime } from '@/utils/format'
import { useWebSocketStore } from '@/store/websocket'

const route = useRoute()

// WebSocket Store
const wsStore = useWebSocketStore()

// 数据
const loading = ref(false)
const submitting = ref(false)
const generatingQR = ref(false)
const deletingQR = ref(false)
const tableData = ref([])
const selectedRows = ref([])
const groupList = ref([])

// 筛选表单
const filterForm = reactive({
  platform_type: '',
  group_id: null,
  online_status: '',
  search: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 对话框
const dialogVisible = ref(false)
const qrDialogVisible = ref(false)
const logDialogVisible = ref(false)
const dialogTitle = computed(() => (formData.id ? '编辑账号' : '新增账号'))
const formRef = ref(null)
const currentQRCode = ref(null)
const currentAccountForQR = ref(null)
const logData = ref([])

// 表单数据
const formData = reactive({
  id: null,
  group_id: null,
  platform_type: 'line',
  line_id: '',
  display_name: '',
  phone_number: '',
  profile_url: '',
  avatar_url: '',
  bio: '',
  status_message: '',
  add_friend_link: '',
  account_remark: '',
  reset_time: null, // 重置时间，为空时使用分组的重置时间
  online_status: 'offline'
})

// 表单验证规则
const formRules = {
  group_id: [{ required: true, message: '请选择所属分组', trigger: 'change' }],
  platform_type: [{ required: true, message: '请选择平台类型', trigger: 'change' }],
  line_id: [{ required: true, message: '请输入Line ID', trigger: 'blur' }]
}

// 二维码图片URL
const qrImageUrl = computed(() => {
  if (!currentQRCode.value) return ''
  // 后端返回的路径格式：/static/qrcodes/1.png 或 /qrcodes/1.png（旧数据）
  // 如果是旧格式（/qrcodes/...），需要转换为 /static/qrcodes/...
  let path = String(currentQRCode.value).trim()
  
  // 兼容旧数据：/qrcodes/... -> /static/qrcodes/...
  if (path.startsWith('/qrcodes/') && !path.startsWith('/static/qrcodes/')) {
    path = path.replace(/^\/qrcodes\//, '/static/qrcodes/')
  }
  
  // 确保路径以 /static/qrcodes/ 开头
  if (!path.startsWith('/static/qrcodes/') && path.includes('qrcodes')) {
    // 如果路径包含 qrcodes 但没有 /static 前缀，添加前缀
    path = '/static' + path
  }
  
  // 开发环境：通过vite代理访问（/static已配置代理）
  return path
})

// 加载分组列表（用于下拉选择）
const loadGroups = async () => {
  try {
    // 使用合理的分页大小，最多获取100条
    const res = await getGroups({ page: 1, page_size: 100 })
    if (res.code === 1000) {
      groupList.value = Array.isArray(res.data?.list) ? res.data.list : []
      // 如果总数超过100，提示用户
      if (res.data?.pagination?.total > 100) {
        console.warn(`分组总数 ${res.data.pagination.total} 条，仅显示前100条用于选择`)
      }
    }
  } catch (error) {
    console.error('加载分组列表失败:', error)
  }
}

// 加载账号列表
const loadAccounts = async () => {
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

    const res = await getLineAccounts(params)
    if (res.code === 1000) {
      tableData.value = Array.isArray(res.data?.list) ? res.data.list : []
      pagination.total = Number(res.data?.pagination?.total) || 0
    } else {
      tableData.value = []
      pagination.total = 0
    }
  } catch (error) {
    console.error('加载账号列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadAccounts()
}

// 重置筛选
const handleReset = () => {
  filterForm.platform_type = ''
  filterForm.group_id = null
  filterForm.online_status = ''
  filterForm.search = ''
  pagination.page = 1
  loadAccounts()
}

// 分页变化
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  loadAccounts()
}

const handlePageChange = (page) => {
  pagination.page = page
  loadAccounts()
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedRows.value = selection
}

// 新增
const handleAdd = () => {
  resetForm()
  dialogVisible.value = true
}

// 编辑
const handleEdit = (row) => {
  resetForm()
  formData.id = row.id
  formData.group_id = row.group_id
  formData.platform_type = row.platform_type
  formData.line_id = row.line_id
  formData.display_name = row.display_name || ''
  formData.phone_number = row.phone_number || ''
  formData.profile_url = row.profile_url || ''
  formData.avatar_url = row.avatar_url || ''
  formData.bio = row.bio || ''
  formData.status_message = row.status_message || ''
  formData.add_friend_link = row.add_friend_link || ''
  formData.account_remark = row.account_remark || ''
  formData.reset_time = row.reset_time || null
  formData.online_status = row.online_status || 'offline'
  dialogVisible.value = true
}

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除账号 "${row.line_id}" 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await deleteLineAccount(row.id)
    if (res.code === 1000) {
      ElMessage.success('删除成功')
      loadAccounts()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

// 批量删除
const handleBatchDelete = async () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请选择要删除的账号')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedRows.value.length} 个账号吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const ids = selectedRows.value.map(row => row.id)
    const res = await batchDeleteLineAccounts(ids)
    if (res.code === 1000) {
      ElMessage.success(
        `批量删除完成：成功 ${res.data?.success_count || 0} 个，失败 ${res.data?.fail_count || 0} 个`
      )
      selectedRows.value = []
      loadAccounts()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      // 如果后端接口不存在，提示用户
      if (error.message && error.message.includes('404')) {
        ElMessage.warning('批量删除接口暂未实现，请单个删除')
      }
    }
  }
}

// 批量强制下线
const handleBatchOffline = async () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请选择要强制下线的账号')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要强制下线选中的 ${selectedRows.value.length} 个账号吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const ids = selectedRows.value.map(row => row.id)
    const res = await batchUpdateLineAccounts({
      ids,
      online_status: 'offline'
    })
    if (res.code === 1000) {
      ElMessage.success(
        `批量强制下线完成：成功 ${res.data?.success_count || 0} 个，失败 ${res.data?.fail_count || 0} 个`
      )
      selectedRows.value = []
      loadAccounts()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量强制下线失败:', error)
      // 如果后端接口不存在，提示用户
      if (error.message && error.message.includes('404')) {
        ElMessage.warning('批量强制下线接口暂未实现')
      }
    }
  }
}

// 查看二维码
const handleViewQR = (row) => {
  currentAccountForQR.value = row
  // 获取二维码路径，如果是旧格式会自动转换
  let qrPath = row.qr_code_path || null
  if (qrPath && qrPath.startsWith('/qrcodes/') && !qrPath.startsWith('/static/qrcodes/')) {
    // 兼容旧数据：立即转换
    qrPath = qrPath.replace(/^\/qrcodes\//, '/static/qrcodes/')
  }
  currentQRCode.value = qrPath
  qrDialogVisible.value = true
}

// 生成二维码
const handleGenerateQR = async () => {
  if (!currentAccountForQR.value) return

  generatingQR.value = true
  try {
    const res = await generateQRCode(currentAccountForQR.value.id)
    if (res.code === 1000) {
      currentQRCode.value = res.data?.qr_code_path || null
      ElMessage.success('二维码生成成功')
      // 刷新列表
      loadAccounts()
    }
  } catch (error) {
    console.error('生成二维码失败:', error)
  } finally {
    generatingQR.value = false
  }
}

// 重新生成二维码
const handleRegenerateQR = async () => {
  if (!currentAccountForQR.value) return

  // 确认对话框
  try {
    await ElMessageBox.confirm(
      '确定要重新生成二维码吗？旧的二维码将被替换。',
      '确认重新生成',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 执行生成
    generatingQR.value = true
    try {
      const res = await generateQRCode(currentAccountForQR.value.id)
      if (res.code === 1000) {
        // 更新二维码路径（强制刷新图片）
        const newQRPath = res.data?.qr_code_path || null
        currentQRCode.value = null // 先清空，强制刷新
        // 使用 nextTick 确保图片重新加载
        await nextTick()
        currentQRCode.value = newQRPath
        
        // 更新当前账号的二维码路径
        if (currentAccountForQR.value) {
          currentAccountForQR.value.qr_code_path = newQRPath
        }
        
        ElMessage.success('二维码重新生成成功')
        // 刷新列表
        loadAccounts()
      }
    } catch (error) {
      console.error('重新生成二维码失败:', error)
      ElMessage.error('重新生成二维码失败')
    } finally {
      generatingQR.value = false
    }
  } catch (error) {
    // 用户取消
    if (error !== 'cancel') {
      console.error('确认对话框错误:', error)
    }
  }
}

// 删除二维码
const handleDeleteQR = async () => {
  if (!currentAccountForQR.value) return

  // 确认对话框
  try {
    await ElMessageBox.confirm(
      '确定要删除二维码吗？删除后可以重新生成。',
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 执行删除（通过更新接口将 qr_code_path 设置为空）
    deletingQR.value = true
    try {
      const res = await updateLineAccount(currentAccountForQR.value.id, {
        qr_code_path: ''
      })
      
      if (res.code === 1000) {
        // 清空当前显示的二维码
        currentQRCode.value = null
        
        // 更新当前账号的二维码路径
        if (currentAccountForQR.value) {
          currentAccountForQR.value.qr_code_path = ''
        }
        
        ElMessage.success('二维码删除成功')
        // 刷新列表
        loadAccounts()
      }
    } catch (error) {
      console.error('删除二维码失败:', error)
      ElMessage.error('删除二维码失败')
    } finally {
      deletingQR.value = false
    }
  } catch (error) {
    // 用户取消
    if (error !== 'cancel') {
      console.error('确认对话框错误:', error)
    }
  }
}

// 下载二维码
const handleDownloadQR = () => {
  if (!qrImageUrl.value) return

  const link = document.createElement('a')
  link.href = qrImageUrl.value
  link.download = `qrcode_${currentAccountForQR.value?.line_id || 'unknown'}.png`
  link.click()
}

// 查看日志
const handleViewLogs = (row) => {
  // TODO: 这里应该调用后端接口获取日志数据
  // 目前先使用模拟数据
  logData.value = [
    {
      timestamp: formatDateTime(row.created_at),
      title: '账号创建',
      content: `账号 ${row.line_id} 被创建`
    },
    {
      timestamp: row.last_active_at ? formatDateTime(row.last_active_at) : '-',
      title: '最后活跃',
      content: `账号最后活跃时间：${row.last_active_at ? formatDateTime(row.last_active_at) : '未知'}`
    }
  ]
  logDialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const data = { ...formData }

      // 移除 id 字段（不提交）
      delete data.id

      // 对于空字符串，保留它们以便可以清空字段
      // 对于 reset_time，如果是 null，在更新时需要提交 null 以便清空；在创建时则不提交
      // 其他字段如果是 null 或 undefined，则移除
      Object.keys(data).forEach(key => {
        if (key === 'reset_time') {
          // reset_time 特殊处理：更新时保留 null，创建时如果是 null 则移除
          if (!formData.id && (data[key] === null || data[key] === undefined)) {
            delete data[key]
          }
        } else if (data[key] === null || data[key] === undefined) {
          delete data[key]
        }
      })

      let res
      if (formData.id) {
        // 更新 - 允许更新所有字段（包括 reset_time 为 null）
        res = await updateLineAccount(formData.id, data)
      } else {
        // 创建 - 移除空字符串（创建时不需要空字段）
        Object.keys(data).forEach(key => {
          if (data[key] === '') {
            delete data[key]
          }
        })
        res = await createLineAccount(data)
      }

      if (res.code === 1000) {
        ElMessage.success(formData.id ? '更新成功' : '创建成功')
        dialogVisible.value = false
        loadAccounts()
      }
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      submitting.value = false
    }
  })
}

// 复制链接
const handleCopyLink = async (link) => {
  try {
    await navigator.clipboard.writeText(link)
    ElMessage.success('链接已复制到剪贴板')
  } catch (error) {
    // 降级方案：使用传统方法
    const textarea = document.createElement('textarea')
    textarea.value = link
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    try {
      document.execCommand('copy')
      ElMessage.success('链接已复制到剪贴板')
    } catch (err) {
      ElMessage.error('复制失败，请手动复制')
    }
    document.body.removeChild(textarea)
  }
}

// 重置表单
const resetForm = () => {
  formData.id = null
  formData.group_id = null
  formData.platform_type = 'line'
  formData.line_id = ''
  formData.display_name = ''
  formData.phone_number = ''
  formData.profile_url = ''
  formData.avatar_url = ''
  formData.bio = ''
  formData.status_message = ''
  formData.add_friend_link = ''
  formData.account_remark = ''
  formData.reset_time = null
  formData.online_status = 'offline'
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

// 初始化WebSocket消息处理器
const initWebSocket = () => {
  wsStore.registerMessageHandler('account-list', (message) => {
    if (message.type === 'account_status_change') {
      handleAccountStatusChange(message.data)
    } else if (message.type === 'account_stats_update') {
      handleAccountStatsUpdate(message.data)
    } else if (message.type === 'account_deleted') {
      handleAccountDeleted(message.data)
    }
  })
}

// 处理账号状态变化消息
const handleAccountStatusChange = (data) => {
  const { line_account_id, online_status, group_id } = data

  // 查找对应的账号并更新状态
  const accountIndex = tableData.value.findIndex(account => account.line_id === line_account_id)

  if (accountIndex !== -1) {
    // 更新账号状态
    tableData.value[accountIndex].online_status = online_status

    // 如果状态变为online，更新last_active_at
    if (online_status === 'online') {
      tableData.value[accountIndex].last_active_at = new Date().toISOString()
    }

    console.log(`账号状态已更新: ${line_account_id} -> ${online_status}`)
  }
}

// 处理账号统计更新消息
const handleAccountStatsUpdate = (data) => {
  const { line_id, total_incoming, today_incoming, duplicate_incoming, today_duplicate } = data

  // 查找对应的账号并更新统计信息
  const accountIndex = tableData.value.findIndex(account => account.line_id === line_id)

  if (accountIndex !== -1) {
    // 更新账号的统计信息
    tableData.value[accountIndex].total_incoming = total_incoming
    tableData.value[accountIndex].today_incoming = today_incoming
    tableData.value[accountIndex].duplicate_incoming = duplicate_incoming
    tableData.value[accountIndex].today_duplicate = today_duplicate

    console.log(`账号统计已更新: ${line_id}`)
  }
}

// 处理账号删除消息
const handleAccountDeleted = (data) => {
  const { group_id, account_id, line_account_id } = data

  // 从列表中移除对应的账号
  const accountIndex = tableData.value.findIndex(account => account.id === account_id)
  if (accountIndex !== -1) {
    const deletedAccount = tableData.value[accountIndex]
    tableData.value.splice(accountIndex, 1)

    // 更新分页总数
    pagination.total = Math.max(0, pagination.total - 1)

    console.log(`账号已删除并从列表移除: ${deletedAccount.line_id} (ID: ${account_id})`)
  }
}

// 初始化
onMounted(() => {
  loadGroups()

  // 如果从分组管理页面跳转过来，自动筛选该分组
  if (route.query.group_id) {
    filterForm.group_id = Number(route.query.group_id)
  }

  loadAccounts()
  initWebSocket()
})

onUnmounted(() => {
  wsStore.unregisterMessageHandler('account-list')
})
</script>

<style scoped lang="less">
@import '@/styles/list-page.less';

.account-list {
  .card-header {
    .header-actions {
      display: flex;
      gap: 10px;
    }
  }

  .qr-preview {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 300px;

    .qr-image {
      padding: 20px;
      background-color: #fff;
      border-radius: 4px;
    }

    .qr-placeholder {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 20px;
    }
  }

  .log-content {
    max-height: 500px;
    overflow-y: auto;
  }

  .friend-link-cell {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .friend-link-form-item {
    display: flex;
    align-items: center;
    width: 100%;

    :deep(.el-input-group__append) {
      padding: 0;
    }
  }
}
</style>
