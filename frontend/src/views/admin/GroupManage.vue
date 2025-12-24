<template>
  <div class="list-page-container group-manage">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>分组管理</span>
        </div>
      </template>

      <!-- 筛选区域 -->
      <div class="filter-section">
        <el-form :model="filterForm" :inline="true" class="filter-form">
          <el-form-item label="状态">
            <el-select
              v-model="filterForm.is_active"
              placeholder="全部"
              clearable
              style="width: 120px"
            >
              <el-option label="激活" :value="true" />
              <el-option label="禁用" :value="false" />
            </el-select>
          </el-form-item>
          <el-form-item label="搜索">
            <el-input
              v-model="filterForm.search"
              placeholder="激活码或备注"
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
          新增分组
        </el-button>
        <el-button
          type="danger"
          :disabled="loading || selectedRows.length === 0"
          @click="handleBatchDelete"
        >
          批量删除 ({{ selectedRows.length }})
        </el-button>
        <el-button
          type="warning"
          :disabled="loading || selectedRows.length === 0"
          @click="handleBatchUpdate"
        >
          批量更新 ({{ selectedRows.length }})
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
        <el-table-column prop="activation_code" label="激活码" width="120">
          <template #default="{ row }">
            <el-tag type="info">{{ row.activation_code }}</el-tag>
            <el-button
              type="text"
              size="small"
              @click="handleRegenerateCode(row)"
              style="margin-left: 8px"
            >
              重新生成
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.remark || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="is_active" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'danger'">
              {{ row.is_active ? '激活' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="account_limit" label="账号限制" width="100">
          <template #default="{ row }">
            {{ row.account_limit == null || row.account_limit === -1 ? '无限制' : row.account_limit }}
          </template>
        </el-table-column>
        <el-table-column label="账号统计" width="150">
          <template #default="{ row }">
            <div class="stats-info">
              <div>总数: <strong>{{ row.total_accounts ?? 0 }}</strong></div>
              <div>在线: <strong style="color: #67c23a">{{ row.online_accounts ?? 0 }}</strong></div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="进线统计" width="220">
          <template #default="{ row }">
            <div class="stats-info">
              <div>当日进线: <strong style="color: #409eff">{{ row.today_incoming ?? 0 }}</strong></div>
              <div>当日重复: <strong style="color: #e6a23c">{{ row.today_duplicate ?? 0 }}</strong></div>
              <div>总进线: <strong>{{ row.total_incoming ?? 0 }}</strong></div>
              <div>总重复: <strong style="color: #e6a23c">{{ row.duplicate_incoming ?? 0 }}</strong></div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="dedup_scope" label="去重范围" width="100">
          <template #default="{ row }">
            <el-tag :type="row.dedup_scope === 'global' ? 'warning' : 'info'">
              {{ row.dedup_scope === 'global' ? '全局' : '当前' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reset_time" label="重置时间" width="120">
          <template #default="{ row }">
            <el-tag type="info" size="small">
              {{ row.reset_time || '09:00:00' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ row.created_at ? formatDateTime(row.created_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">
            {{ row.updated_at ? formatDateTime(row.updated_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">
              查看
            </el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
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
        <el-form-item label="账号限制" prop="account_limit">
          <el-input-number
            v-model="formData.account_limit"
            :min="-1"
            style="width: 100%"
            placeholder="-1表示无限制，0表示显示为0但允许"
          />
          <div style="font-size: 12px; color: #909399; margin-top: 4px;">
            提示：-1表示无限制，0表示显示为0但实际允许，大于0表示有限制
          </div>
        </el-form-item>
        <el-form-item label="状态" prop="is_active">
          <el-switch
            v-model="formData.is_active"
            active-text="激活"
            inactive-text="禁用"
          />
        </el-form-item>
        <el-form-item label="去重范围" prop="dedup_scope">
          <el-radio-group v-model="formData.dedup_scope">
            <el-radio label="current">当前分组</el-radio>
            <el-radio label="global">全局</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="重置时间" prop="reset_time">
          <el-time-picker
            v-model="formData.reset_time"
            format="HH:mm:ss"
            value-format="HH:mm:ss"
            placeholder="选择重置时间"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="2"
            placeholder="请输入备注"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="登录密码" prop="login_password">
          <el-input
            v-model="formData.login_password"
            type="password"
            placeholder="子账号登录密码（6位以上）"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量更新对话框 -->
    <el-dialog
      v-model="batchUpdateDialogVisible"
      title="批量更新分组"
      width="500px"
    >
      <el-form
        ref="batchFormRef"
        :model="batchFormData"
        label-width="120px"
      >
        <el-form-item label="状态">
          <el-select
            v-model="batchFormData.is_active"
            placeholder="不修改"
            clearable
            style="width: 100%"
          >
            <el-option label="激活" :value="true" />
            <el-option label="禁用" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="去重范围">
          <el-select
            v-model="batchFormData.dedup_scope"
            placeholder="不修改"
            clearable
            style="width: 100%"
          >
            <el-option label="当前分组" value="current" />
            <el-option label="全局" value="global" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchUpdateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBatchUpdateSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import {
  getGroups,
  createGroup,
  updateGroup,
  deleteGroup,
  regenerateCode,
  batchDeleteGroups,
  batchUpdateGroups
} from '@/api/group'
import { formatDateTime } from '@/utils/format'
import { useAuthStore } from '@/store/auth'
import { useWebSocketStore } from '@/store/websocket'

const router = useRouter()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()
const isAdmin = computed(() => authStore.isAdmin)

// 数据
const loading = ref(false)
const submitting = ref(false)
const tableData = ref([])
const selectedRows = ref([])

// 筛选表单
const filterForm = reactive({
  is_active: null,
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
const batchUpdateDialogVisible = ref(false)
const dialogTitle = computed(() => (formData.id ? '编辑分组' : '新增分组'))
const formRef = ref(null)
const batchFormRef = ref(null)

// 表单数据
const formData = reactive({
  id: null,
  user_id: null,
  account_limit: null,
  is_active: true,
  dedup_scope: 'current',
  reset_time: '',
  remark: '',
  description: '',
  login_password: ''
})

// 批量更新表单
const batchFormData = reactive({
  is_active: null,
  dedup_scope: ''
})

// 表单验证规则
const formRules = {
  dedup_scope: [
    { required: true, message: '请选择去重范围', trigger: 'change' }
  ]
}

// 加载分组列表
const loadGroups = async () => {
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

    const res = await getGroups(params)
    if (res.code === 1000) {
      // 后端返回结构: { list: [], pagination: { total: 0, ... } }
      tableData.value = Array.isArray(res.data?.list) ? res.data.list : []
      pagination.total = Number(res.data?.pagination?.total) || 0
    } else {
      tableData.value = []
      pagination.total = 0
    }
  } catch (error) {
    console.error('加载分组列表失败:', error)
  } finally {
    loading.value = false
  }
}


// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadGroups()
}

// 重置筛选
const handleReset = () => {
  filterForm.is_active = null
  filterForm.search = ''
  pagination.page = 1
  loadGroups()
}

// 分页变化
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  loadGroups()
}

const handlePageChange = (page) => {
  pagination.page = page
  loadGroups()
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedRows.value = selection
}

// 新增
const handleAdd = () => {
  resetForm()
  // 管理员可以指定用户ID，普通用户自动使用当前用户ID
  if (isAdmin.value) {
    formData.user_id = authStore.user?.id || null
  } else {
    // 普通用户自动设置为当前用户ID
    formData.user_id = authStore.user?.id || null
  }
  dialogVisible.value = true
}

// 查看 - 跳转到账号列表并筛选对应分组
const handleView = (row) => {
  // 跳转到账号列表页面，并传递分组ID作为查询参数
  router.push({
    path: '/accounts',
    query: {
      group_id: row.id
    }
  })
}

// 编辑
const handleEdit = (row) => {
  resetForm()
  formData.id = row.id
  formData.user_id = row.user_id
  formData.account_limit = row.account_limit
  formData.is_active = row.is_active
  formData.dedup_scope = row.dedup_scope || 'current'
  formData.reset_time = row.reset_time || ''
  formData.remark = row.remark || ''
  formData.description = row.description || ''
  formData.login_password = '' // 编辑时不显示密码
  dialogVisible.value = true
}

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除分组 "${row.activation_code}" 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await deleteGroup(row.id)
    if (res.code === 1000) {
      ElMessage.success('删除成功')
      loadGroups()
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
    ElMessage.warning('请选择要删除的分组')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedRows.value.length} 个分组吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const ids = selectedRows.value.map(row => row.id)
    const res = await batchDeleteGroups(ids)
    if (res.code === 1000) {
      ElMessage.success(
        `批量删除完成：成功 ${res.data?.success_count || 0} 个，失败 ${res.data?.fail_count || 0} 个`
      )
      selectedRows.value = []
      loadGroups()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
    }
  }
}

// 批量更新
const handleBatchUpdate = () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请选择要更新的分组')
    return
  }
  batchFormData.is_active = null
  batchFormData.dedup_scope = ''
  batchUpdateDialogVisible.value = true
}

// 批量更新提交
const handleBatchUpdateSubmit = async () => {
  const updateData = {
    ids: selectedRows.value.map(row => row.id)
  }

  if (batchFormData.is_active !== null) {
    updateData.is_active = batchFormData.is_active
  }
  if (batchFormData.dedup_scope !== '') {
    updateData.dedup_scope = batchFormData.dedup_scope
  }

  if (Object.keys(updateData).length === 1) {
    ElMessage.warning('请至少选择一个要更新的字段')
    return
  }

  submitting.value = true
  try {
    const res = await batchUpdateGroups(updateData)
    if (res.code === 1000) {
      ElMessage.success(
        `批量更新完成：成功 ${res.data?.success_count || 0} 个，失败 ${res.data?.fail_count || 0} 个`
      )
      batchUpdateDialogVisible.value = false
      selectedRows.value = []
      loadGroups()
    }
  } catch (error) {
    console.error('批量更新失败:', error)
  } finally {
    submitting.value = false
  }
}

// 重新生成激活码
const handleRegenerateCode = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要重新生成分组 "${row.activation_code}" 的激活码吗？原激活码将失效。`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await regenerateCode(row.id)
    if (res.code === 1000) {
      ElMessage.success(`激活码已重新生成：${res.data?.activation_code}`)
      loadGroups()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重新生成激活码失败:', error)
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const data = { ...formData }
      
      // 创建分组时，自动使用当前登录用户的ID
      if (!formData.id) {
        data.user_id = authStore.user?.id
      }
      
      // 移除空值（但保留 -1，因为 -1 表示无限制）
      Object.keys(data).forEach(key => {
        if (data[key] === '' || data[key] === null || data[key] === undefined) {
          delete data[key]
        }
        // account_limit 为 -1 时保留（表示无限制）
        if (key === 'account_limit' && data[key] === -1) {
          // 保留 -1
        }
      })

      // 编辑时不需要密码字段（如果为空）
      if (formData.id && !data.login_password) {
        delete data.login_password
      }
      // 编辑时不更新user_id（保持原有归属）
      if (formData.id) {
        delete data.user_id
      }

      let res
      if (formData.id) {
        // 更新
        const { id, ...updateData } = data
        res = await updateGroup(id, updateData)
      } else {
        // 创建
        res = await createGroup(data)
      }

      if (res.code === 1000) {
        ElMessage.success(formData.id ? '更新成功' : '创建成功')
        dialogVisible.value = false
        loadGroups()
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
  formData.user_id = null
  formData.account_limit = null
  formData.is_active = true
  formData.dedup_scope = 'current'
  formData.reset_time = ''
  formData.remark = ''
  formData.description = ''
  formData.login_password = ''
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

// 初始化WebSocket消息处理器
const initWebSocket = () => {
  // 注册消息处理器
  wsStore.registerMessageHandler('group-manage', (message) => {
    // 处理分组统计更新消息
    if (message.type === 'group_stats_update') {
      handleGroupStatsUpdate(message.data)
    }
  })
}

// 处理分组统计更新
const handleGroupStatsUpdate = (data) => {
  // 在表格数据中找到对应的分组并更新统计信息
  const groupIndex = tableData.value.findIndex(group => group.id === data.group_id)
  if (groupIndex !== -1) {
    // 更新分组的统计信息
    tableData.value[groupIndex].total_accounts = data.total_accounts
    tableData.value[groupIndex].online_accounts = data.online_accounts
    tableData.value[groupIndex].total_incoming = data.total_incoming
    tableData.value[groupIndex].today_incoming = data.today_incoming
    tableData.value[groupIndex].duplicate_incoming = data.duplicate_incoming
    tableData.value[groupIndex].today_duplicate = data.today_duplicate

    // 触发Vue响应式更新
    tableData.value.splice(groupIndex, 1, { ...tableData.value[groupIndex] })
  }
}

// 初始化
onMounted(() => {
  loadGroups()
  initWebSocket()
})

onUnmounted(() => {
  // 取消注册消息处理器
  wsStore.unregisterMessageHandler('group-manage')
})
</script>

<style scoped lang="less">
@import '@/styles/list-page.less';

.group-manage {
  .card-header {
    .header-actions {
      display: flex;
      gap: 10px;
    }
  }
}
</style>
