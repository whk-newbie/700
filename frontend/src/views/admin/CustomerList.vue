<template>
  <div class="customer-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>客户列表</span>
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
          <el-form-item label="客户类型">
            <el-select
              v-model="filterForm.customer_type"
              placeholder="全部"
              clearable
              style="width: 180px"
            >
              <el-option label="新增线索-实时" value="新增线索-实时" />
              <el-option label="新增线索-补录" value="新增线索-补录" />
              <el-option label="新增线索-重复" value="新增线索-重复" />
              <el-option label="新增线索-导入重复" value="新增线索-导入重复" />
            </el-select>
          </el-form-item>
          <el-form-item label="搜索">
            <el-input
              v-model="filterForm.search"
              placeholder="客户ID或显示名称"
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
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="customer_id" label="客户ID" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.customer_id }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="头像" width="80">
          <template #default="{ row }">
            <el-avatar
              v-if="row.avatar_url"
              :src="row.avatar_url"
              :size="40"
              shape="square"
            />
            <el-avatar v-else :size="40" shape="square">
              <el-icon><User /></el-icon>
            </el-avatar>
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
        <el-table-column prop="customer_type" label="客户类型" width="150">
          <template #default="{ row }">
            <el-tag
              v-if="row.customer_type"
              :type="getCustomerTypeTagType(row.customer_type)"
              size="small"
            >
              {{ row.customer_type }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="phone_number" label="手机号" width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.phone_number || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">
              查看详情
            </el-button>
            <el-button type="success" link size="small" @click="handleViewFollowUps(row)">
              跟进记录
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

    <!-- 客户详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="detailDialogTitle"
      width="700px"
      :close-on-click-modal="false"
    >
      <div v-if="currentCustomer">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本信息" name="info">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="客户ID">
                <el-tag type="info" size="small">{{ currentCustomer.customer_id }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="显示名称">
                {{ currentCustomer.display_name || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="平台类型">
                <el-tag :type="currentCustomer.platform_type === 'line' ? 'primary' : 'success'">
                  {{ currentCustomer.platform_type === 'line' ? 'Line' : 'Line Business' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="客户类型">
                <el-tag
                  v-if="currentCustomer.customer_type"
                  :type="getCustomerTypeTagType(currentCustomer.customer_type)"
                  size="small"
                >
                  {{ currentCustomer.customer_type }}
                </el-tag>
                <span v-else>-</span>
              </el-descriptions-item>
              <el-descriptions-item label="激活码">
                <el-tag type="info" size="small">{{ currentCustomer.activation_code }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="手机号">
                {{ currentCustomer.phone_number || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="性别">
                {{ currentCustomer.gender === 'male' ? '男' : currentCustomer.gender === 'female' ? '女' : currentCustomer.gender || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="国家">
                {{ currentCustomer.country || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="生日">
                {{ currentCustomer.birthday || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="地址" :span="2">
                {{ currentCustomer.address || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="头像URL" :span="2">
                <el-link
                  v-if="currentCustomer.avatar_url"
                  :href="currentCustomer.avatar_url"
                  target="_blank"
                  type="primary"
                >
                  {{ currentCustomer.avatar_url }}
                </el-link>
                <span v-else>-</span>
              </el-descriptions-item>
              <el-descriptions-item label="昵称备注" :span="2">
                {{ currentCustomer.nickname_remark || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="备注" :span="2">
                {{ currentCustomer.remark || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">
                {{ formatDateTime(currentCustomer.created_at) }}
              </el-descriptions-item>
              <el-descriptions-item label="更新时间">
                {{ formatDateTime(currentCustomer.updated_at) }}
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane label="编辑信息" name="edit">
            <el-form
              ref="editFormRef"
              :model="editFormData"
              :rules="editFormRules"
              label-width="120px"
            >
              <el-form-item label="显示名称">
                <el-input v-model="editFormData.display_name" placeholder="请输入显示名称" />
              </el-form-item>
              <el-form-item label="头像URL">
                <el-input v-model="editFormData.avatar_url" placeholder="请输入头像URL" />
              </el-form-item>
              <el-form-item label="手机号">
                <el-input v-model="editFormData.phone_number" placeholder="请输入手机号" />
              </el-form-item>
              <el-form-item label="客户类型">
                <el-select v-model="editFormData.customer_type" placeholder="请选择客户类型" style="width: 100%">
                  <el-option label="新增线索-实时" value="新增线索-实时" />
                  <el-option label="新增线索-补录" value="新增线索-补录" />
                  <el-option label="新增线索-重复" value="新增线索-重复" />
                  <el-option label="新增线索-导入重复" value="新增线索-导入重复" />
                </el-select>
              </el-form-item>
              <el-form-item label="性别">
                <el-select v-model="editFormData.gender" placeholder="请选择性别" style="width: 100%">
                  <el-option label="男" value="male" />
                  <el-option label="女" value="female" />
                  <el-option label="未知" value="unknown" />
                </el-select>
              </el-form-item>
              <el-form-item label="国家">
                <el-input v-model="editFormData.country" placeholder="请输入国家代码（如：TW）" />
              </el-form-item>
              <el-form-item label="生日">
                <el-date-picker
                  v-model="editFormData.birthday"
                  type="date"
                  placeholder="请选择生日"
                  format="YYYY-MM-DD"
                  value-format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </el-form-item>
              <el-form-item label="地址">
                <el-input
                  v-model="editFormData.address"
                  type="textarea"
                  :rows="2"
                  placeholder="请输入地址"
                />
              </el-form-item>
              <el-form-item label="昵称备注">
                <el-input
                  v-model="editFormData.nickname_remark"
                  type="textarea"
                  :rows="2"
                  placeholder="请输入昵称备注"
                  maxlength="200"
                  show-word-limit
                />
              </el-form-item>
              <el-form-item label="备注">
                <el-input
                  v-model="editFormData.remark"
                  type="textarea"
                  :rows="3"
                  placeholder="请输入备注"
                  maxlength="500"
                  show-word-limit
                />
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button
          v-if="activeTab === 'edit'"
          type="primary"
          @click="handleSaveEdit"
          :loading="saving"
        >
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, User } from '@element-plus/icons-vue'
import { getCustomers, getCustomer, updateCustomer, deleteCustomer } from '@/api/customer'
import { getGroups } from '@/api/group'
import { formatDateTime } from '@/utils/format'

const router = useRouter()

// 数据
const loading = ref(false)
const tableData = ref([])
const groupList = ref([])
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 筛选表单
const filterForm = reactive({
  platform_type: '',
  group_id: null,
  customer_type: '',
  search: ''
})

// 详情对话框
const detailDialogVisible = ref(false)
const detailDialogTitle = ref('客户详情')
const activeTab = ref('info')
const currentCustomer = ref(null)
const editFormRef = ref(null)
const saving = ref(false)

// 编辑表单数据
const editFormData = reactive({
  display_name: '',
  avatar_url: '',
  phone_number: '',
  customer_type: '',
  gender: '',
  country: '',
  birthday: '',
  address: '',
  nickname_remark: '',
  remark: ''
})

// 编辑表单验证规则
const editFormRules = {
  // 可以添加验证规则
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

// 获取客户列表
const fetchCustomers = async () => {
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

    const res = await getCustomers(params)
    if (res.code === 1000) {
      tableData.value = res.data.list || []
      pagination.total = res.data.pagination?.total || 0
    } else {
      ElMessage.error(res.message || '获取客户列表失败')
    }
  } catch (error) {
    console.error('获取客户列表失败:', error)
    ElMessage.error('获取客户列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchCustomers()
}

// 重置
const handleReset = () => {
  filterForm.platform_type = ''
  filterForm.group_id = null
  filterForm.customer_type = ''
  filterForm.search = ''
  pagination.page = 1
  fetchCustomers()
}

// 刷新
const handleRefresh = () => {
  fetchCustomers()
}

// 分页变化
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchCustomers()
}

const handlePageChange = (page) => {
  pagination.page = page
  fetchCustomers()
}

// 查看详情
const handleViewDetail = async (row) => {
  try {
    const res = await getCustomer(row.id)
    if (res.code === 1000) {
      currentCustomer.value = res.data
      // 填充编辑表单
      editFormData.display_name = res.data.display_name || ''
      editFormData.avatar_url = res.data.avatar_url || ''
      editFormData.phone_number = res.data.phone_number || ''
      editFormData.customer_type = res.data.customer_type || ''
      editFormData.gender = res.data.gender || ''
      editFormData.country = res.data.country || ''
      editFormData.birthday = res.data.birthday || ''
      editFormData.address = res.data.address || ''
      editFormData.nickname_remark = res.data.nickname_remark || ''
      editFormData.remark = res.data.remark || ''
      activeTab.value = 'info'
      detailDialogVisible.value = true
    } else {
      ElMessage.error(res.message || '获取客户详情失败')
    }
  } catch (error) {
    console.error('获取客户详情失败:', error)
    ElMessage.error('获取客户详情失败')
  }
}

// 保存编辑
const handleSaveEdit = async () => {
  if (!editFormRef.value) return
  
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return

    if (!currentCustomer.value) return

    saving.value = true
    try {
      const updateData = { ...editFormData }
      // 移除空字符串
      Object.keys(updateData).forEach(key => {
        if (updateData[key] === '') {
          updateData[key] = undefined
        }
      })

      const res = await updateCustomer(currentCustomer.value.id, updateData)
      if (res.code === 1000) {
        ElMessage.success('更新成功')
        detailDialogVisible.value = false
        fetchCustomers()
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } catch (error) {
      console.error('更新客户失败:', error)
      ElMessage.error('更新客户失败')
    } finally {
      saving.value = false
    }
  })
}

// 查看跟进记录
const handleViewFollowUps = (row) => {
  // 跳转到跟进记录页面，并筛选该客户
  router.push({
    name: 'FollowUpRecords',
    query: {
      customer_id: row.id
    }
  })
}

// 获取客户类型标签类型
const getCustomerTypeTagType = (customerType) => {
  if (!customerType) return 'info'
  if (customerType.includes('实时')) return 'success'
  if (customerType.includes('补录')) return 'primary'
  if (customerType.includes('重复')) return 'warning'
  if (customerType.includes('导入重复')) return 'danger'
  return 'info'
}

// 初始化
onMounted(() => {
  fetchGroups()
  fetchCustomers()
})
</script>

<style scoped lang="less">
.customer-list {
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
