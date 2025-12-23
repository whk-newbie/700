<template>
  <div class="follow-up-records">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>跟进记录</span>
          <div class="header-actions">
            <el-button type="primary" :disabled="loading" @click="handleAdd">
              <el-icon><Plus /></el-icon>
              新增记录
            </el-button>
            <el-button type="primary" :disabled="loading" @click="handleRefresh">
              <el-icon><Refresh /></el-icon>
              刷新
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
              placeholder="跟进内容"
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
        stripe
        empty-text="暂无数据"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="activation_code" label="激活码" width="120">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.activation_code }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="主账号信息" width="200">
          <template #default="{ row }">
            <div class="account-info">
              <div class="account-header">
                <el-avatar
                  v-if="row.line_account_avatar_url"
                  :src="row.line_account_avatar_url"
                  :size="32"
                  shape="square"
                />
                <el-avatar v-else :size="32" shape="square">
                  <el-icon><User /></el-icon>
                </el-avatar>
                <el-tag
                  :type="row.platform_type === 'line' ? 'primary' : 'success'"
                  size="small"
                  style="margin-left: 8px"
                >
                  {{ row.platform_type === 'line' ? 'Line' : 'Line Business' }}
                </el-tag>
              </div>
              <div class="account-name" v-if="row.line_account_display_name">
                {{ row.line_account_display_name }}
              </div>
              <div class="account-id" v-if="row.line_account_line_id">
                {{ row.line_account_line_id }}
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="line_account_line_id" label="主账号" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link
              v-if="row.line_account_line_id"
              type="primary"
              :underline="false"
              @click="handleViewAccount(row)"
            >
              {{ row.line_account_line_id }}
            </el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="线索账号信息" width="200">
          <template #default="{ row }">
            <div class="account-info" v-if="row.customer_display_name || row.customer_line_id">
              <div class="account-header">
                <el-avatar
                  v-if="row.customer_avatar_url"
                  :src="row.customer_avatar_url"
                  :size="32"
                  shape="square"
                />
                <el-avatar v-else :size="32" shape="square">
                  <el-icon><User /></el-icon>
                </el-avatar>
              </div>
              <div class="account-name" v-if="row.customer_display_name">
                {{ row.customer_display_name }}
              </div>
              <div class="account-id" v-if="row.customer_line_id">
                {{ row.customer_line_id }}
              </div>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="customer_line_id" label="线索账号" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link
              v-if="row.customer_line_id"
              type="primary"
              :underline="false"
              @click="handleViewCustomer(row)"
            >
              {{ row.customer_line_id }}
            </el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建日期" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="content-cell">{{ row.content || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
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
        <el-form-item label="所属分组" prop="group_id">
          <el-select
            v-model="formData.group_id"
            placeholder="请选择分组"
            filterable
            style="width: 100%"
            :disabled="!!formData.id"
            @change="handleGroupChange"
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
          <el-radio-group v-model="formData.platform_type" :disabled="!!formData.id">
            <el-radio label="line">Line</el-radio>
            <el-radio label="line_business">Line Business</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Line账号" v-if="!formData.id">
          <el-select
            v-model="formData.line_account_id"
            placeholder="请选择Line账号（可选）"
            filterable
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="account in lineAccountList"
              :key="account.id"
              :label="`${account.display_name || account.line_id} (${account.line_id})`"
              :value="account.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="客户" v-if="!formData.id">
          <el-select
            v-model="formData.customer_id"
            placeholder="请选择客户（可选）"
            filterable
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="customer in customerList"
              :key="customer.id"
              :label="`${customer.display_name || customer.customer_id} (${customer.customer_id})`"
              :value="customer.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="跟进内容" prop="content">
          <el-input
            v-model="formData.content"
            type="textarea"
            :rows="6"
            placeholder="请输入跟进内容"
            maxlength="2000"
            show-word-limit
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, User } from '@element-plus/icons-vue'
import {
  getFollowUps,
  createFollowUp,
  updateFollowUp,
  deleteFollowUp
} from '@/api/followUp'
import { getGroups } from '@/api/group'
import { getLineAccounts } from '@/api/lineAccount'
import { getCustomers } from '@/api/customer'
import { formatDateTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()

// 数据
const loading = ref(false)
const tableData = ref([])
const groupList = ref([])
const lineAccountList = ref([])
const customerList = ref([])
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 筛选表单
const filterForm = reactive({
  platform_type: '',
  group_id: null,
  customer_id: null,
  line_account_id: null,
  search: '',
  start_time: '',
  end_time: ''
})

const dateRange = ref(null)

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('新增记录')
const formRef = ref(null)
const submitting = ref(false)

// 表单数据
const formData = reactive({
  id: null,
  group_id: null,
  platform_type: 'line',
  line_account_id: null,
  customer_id: null,
  content: ''
})

// 表单验证规则
const formRules = {
  group_id: [{ required: true, message: '请选择所属分组', trigger: 'change' }],
  platform_type: [{ required: true, message: '请选择平台类型', trigger: 'change' }],
  content: [{ required: true, message: '请输入跟进内容', trigger: 'blur' }]
}

// 获取分组列表
const fetchGroups = async () => {
  try {
    const res = await getGroups({ page: 1, page_size: 100 })
    if (res.code === 1000) {
      groupList.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取分组列表失败:', error)
  }
}

// 获取Line账号列表（用于新增时选择）
const fetchLineAccounts = async (groupId) => {
  if (!groupId) {
    lineAccountList.value = []
    return
  }
  try {
    const res = await getLineAccounts({ group_id: groupId, page: 1, page_size: 100 })
    if (res.code === 1000) {
      lineAccountList.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取Line账号列表失败:', error)
  }
}

// 获取客户列表（用于新增时选择）
const fetchCustomers = async (groupId) => {
  if (!groupId) {
    customerList.value = []
    return
  }
  try {
    const res = await getCustomers({ group_id: groupId, page: 1, page_size: 100 })
    if (res.code === 1000) {
      customerList.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取客户列表失败:', error)
  }
}

// 获取跟进记录列表
const fetchFollowUps = async () => {
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

    const res = await getFollowUps(params)
    if (res.code === 1000) {
      tableData.value = res.data.list || []
      pagination.total = res.data.pagination?.total || 0
    } else {
      ElMessage.error(res.message || '获取跟进记录列表失败')
    }
  } catch (error) {
    console.error('获取跟进记录列表失败:', error)
    ElMessage.error('获取跟进记录列表失败')
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
  fetchFollowUps()
}

// 重置
const handleReset = () => {
  filterForm.platform_type = ''
  filterForm.group_id = null
  filterForm.customer_id = null
  filterForm.line_account_id = null
  filterForm.search = ''
  filterForm.start_time = ''
  filterForm.end_time = ''
  dateRange.value = null
  pagination.page = 1
  fetchFollowUps()
}

// 刷新
const handleRefresh = () => {
  fetchFollowUps()
}

// 分页变化
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchFollowUps()
}

const handlePageChange = (page) => {
  pagination.page = page
  fetchFollowUps()
}

// 新增
const handleAdd = () => {
  dialogTitle.value = '新增记录'
  formData.id = null
  formData.group_id = null
  formData.platform_type = 'line'
  formData.line_account_id = null
  formData.customer_id = null
  formData.content = ''
  lineAccountList.value = []
  customerList.value = []
  dialogVisible.value = true
}

// 编辑
const handleEdit = (row) => {
  dialogTitle.value = '编辑记录'
  formData.id = row.id
  formData.group_id = row.group_id
  formData.platform_type = row.platform_type
  formData.content = row.content
  // 编辑时不能修改分组和平台类型
  dialogVisible.value = true
}

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除这条跟进记录吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await deleteFollowUp(row.id)
    if (res.code === 1000) {
      ElMessage.success('删除成功')
      fetchFollowUps()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除跟进记录失败:', error)
      ElMessage.error('删除跟进记录失败')
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
      if (formData.id) {
        // 编辑
        const res = await updateFollowUp(formData.id, {
          content: formData.content
        })
        if (res.code === 1000) {
          ElMessage.success('更新成功')
          dialogVisible.value = false
          fetchFollowUps()
        } else {
          ElMessage.error(res.message || '更新失败')
        }
      } else {
        // 新增
        const res = await createFollowUp({
          group_id: formData.group_id,
          platform_type: formData.platform_type,
          line_account_id: formData.line_account_id || undefined,
          customer_id: formData.customer_id || undefined,
          content: formData.content
        })
        if (res.code === 1000) {
          ElMessage.success('创建成功')
          dialogVisible.value = false
          fetchFollowUps()
        } else {
          ElMessage.error(res.message || '创建失败')
        }
      }
    } catch (error) {
      console.error('提交失败:', error)
      ElMessage.error('提交失败')
    } finally {
      submitting.value = false
    }
  })
}

// 查看账号
const handleViewAccount = (row) => {
  if (row.line_account_id) {
    router.push({
      name: 'AccountList',
      query: {
        line_account_id: row.line_account_id
      }
    })
  }
}

// 查看客户
const handleViewCustomer = (row) => {
  if (row.customer_id) {
    router.push({
      name: 'CustomerList',
      query: {
        customer_id: row.customer_id
      }
    })
  }
}

// 监听分组变化，加载对应的账号和客户列表
const handleGroupChange = async () => {
  if (formData.group_id) {
    await Promise.all([
      fetchLineAccounts(formData.group_id),
      fetchCustomers(formData.group_id)
    ])
  } else {
    lineAccountList.value = []
    customerList.value = []
  }
}

// 初始化
onMounted(() => {
  fetchGroups()
  
  // 如果路由中有customer_id参数，自动筛选
  if (route.query.customer_id) {
    filterForm.customer_id = Number(route.query.customer_id)
  }
  
  fetchFollowUps()
})
</script>

<style scoped lang="less">
.follow-up-records {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-actions {
    display: flex;
    gap: 10px;
  }

  .filter-section {
    margin-bottom: 20px;
    padding: 20px;
    background-color: #f5f7fa;
    border-radius: 4px;
  }

  .account-info {
    .account-header {
      display: flex;
      align-items: center;
      margin-bottom: 4px;
    }

    .account-name {
      font-size: 14px;
      color: #606266;
      margin-top: 4px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .account-id {
      font-size: 12px;
      color: #909399;
      margin-top: 2px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .content-cell {
    max-width: 300px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
