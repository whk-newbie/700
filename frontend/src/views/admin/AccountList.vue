<template>
  <div class="account-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>账号列表</span>
          <div class="header-actions">
            <el-button
              v-if="selectedRows.length > 0"
              type="danger"
              :disabled="loading"
              @click="handleBatchDelete"
            >
              批量移除 ({{ selectedRows.length }})
            </el-button>
            <el-button
              v-if="selectedRows.length > 0"
              type="warning"
              :disabled="loading"
              @click="handleBatchOffline"
            >
              强制下线 ({{ selectedRows.length }})
            </el-button>
            <el-button type="primary" :disabled="loading" @click="handleAdd">
              <el-icon><Plus /></el-icon>
              新增账号
            </el-button>
          </div>
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
            :disabled="!!formData.id"
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
            :disabled="!!formData.id"
          />
        </el-form-item>
        <el-form-item label="显示名称">
          <el-input v-model="formData.display_name" placeholder="请输入显示名称" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="formData.phone_number" placeholder="请输入手机号" />
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
          <el-button type="primary" @click="handleGenerateQR" :loading="generatingQR">
            生成二维码
          </el-button>
        </div>
      </div>
      <template #footer>
        <el-button @click="qrDialogVisible = false">关闭</el-button>
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
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
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

const route = useRoute()

// 数据
const loading = ref(false)
const submitting = ref(false)
const generatingQR = ref(false)
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
  avatar_url: '',
  bio: '',
  status_message: '',
  account_remark: '',
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
  // 后端返回的路径格式：/static/qrcodes/1.png
  // 开发环境：通过vite代理访问（/static已配置代理）
  // 生产环境：需要拼接完整URL
  if (currentQRCode.value.startsWith('/')) {
    // 开发环境直接使用相对路径，通过代理访问
    return currentQRCode.value
  }
  return currentQRCode.value
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
  formData.avatar_url = row.avatar_url || ''
  formData.bio = row.bio || ''
  formData.status_message = row.status_message || ''
  formData.account_remark = row.account_remark || ''
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
  currentQRCode.value = row.qr_code_path || null
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

      // 移除空值
      Object.keys(data).forEach(key => {
        if (data[key] === '' || data[key] === null || data[key] === undefined) {
          delete data[key]
        }
      })

      // 编辑时不需要group_id和line_id（这些字段不可修改）
      if (formData.id) {
        delete data.group_id
        delete data.line_id
        delete data.id
      }

      let res
      if (formData.id) {
        // 更新
        res = await updateLineAccount(formData.id, data)
      } else {
        // 创建
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

// 重置表单
const resetForm = () => {
  formData.id = null
  formData.group_id = null
  formData.platform_type = 'line'
  formData.line_id = ''
  formData.display_name = ''
  formData.phone_number = ''
  formData.avatar_url = ''
  formData.bio = ''
  formData.status_message = ''
  formData.account_remark = ''
  formData.online_status = 'offline'
  if (formRef.value) {
    formRef.value.clearValidate()
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
})
</script>

<style scoped lang="less">
.account-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .header-actions {
      display: flex;
      gap: 10px;
    }
  }

  .filter-section {
    margin-bottom: 20px;
    padding: 20px;
    background-color: #f5f7fa;
    border-radius: 4px;

    .filter-form {
      margin: 0;
    }
  }

  .stats-info {
    font-size: 12px;
    line-height: 1.8;

    strong {
      font-weight: 600;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
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
}
</style>
