<template>
  <div class="list-page-container contact-pool">
    <!-- 统计卡片 -->
    <div class="stats-cards">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card class="stats-card">
            <div class="stats-item">
              <div class="stats-label">导入原始联系人数量</div>
              <div class="stats-value">{{ summary.import_count || 0 }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="stats-card">
            <div class="stats-item">
              <div class="stats-label">平台工单原始联系人数量</div>
              <div class="stats-value highlight">{{ summary.platform_count || 0 }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="stats-card">
            <div class="stats-item">
              <div class="stats-label">原始联系人数量汇总</div>
              <div class="stats-value warning">{{ summary.total_count || 0 }}</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <el-card style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>底库管理</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 统计汇总标签页 -->
        <el-tab-pane label="统计汇总" name="summary">
          <div class="tab-content">
            <!-- 筛选区域 -->
            <div class="filter-section">
              <el-form :model="summaryFilterForm" :inline="true" class="filter-form">
                <el-form-item label="平台">
                  <el-select
                    v-model="summaryFilterForm.platform_type"
                    placeholder="全部"
                    clearable
                    style="width: 150px"
                  >
                    <el-option label="Line" value="line" />
                    <el-option label="Line Business" value="line_business" />
                  </el-select>
                </el-form-item>
                <el-form-item label="搜索">
                  <el-input
                    v-model="summaryFilterForm.search"
                    placeholder="激活码"
                    clearable
                    style="width: 200px"
                    @keyup.enter="handleSummarySearch"
                  >
                    <template #prefix>
                      <el-icon><Search /></el-icon>
                    </template>
                  </el-input>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="handleSummarySearch" :loading="loading">
                    <el-icon><Search /></el-icon>
                    搜索
                  </el-button>
                  <el-button @click="handleSummaryReset">重置</el-button>
                </el-form-item>
              </el-form>
            </div>

            <!-- 操作按钮区域 -->
            <div class="action-buttons">
              <el-button type="primary" :disabled="loading" @click="handleImport">
                <el-icon><Upload /></el-icon>
                导入联系人
              </el-button>
              <el-button type="success" :disabled="loading" @click="handleDownloadTemplate">
                <el-icon><Download /></el-icon>
                下载导入模板
              </el-button>
            </div>

            <!-- 数据表格 -->
            <el-table
              v-loading="loading"
              :data="summaryTableData"
              style="width: 100%"
              stripe
              empty-text="暂无数据"
            >
              <el-table-column prop="activation_code" label="激活码" width="150">
                <template #default="{ row }">
                  <el-tag type="info">{{ row.activation_code }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ row.remark || '-' }}
                </template>
              </el-table-column>
              <el-table-column prop="platform_type" label="平台" width="150">
                <template #default="{ row }">
                  <el-tag :type="row.platform_type === 'line' ? 'primary' : 'success'">
                    {{ row.platform_type === 'line' ? 'Line' : 'Line Business' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="contact_count" label="联系人数量" width="150">
                <template #default="{ row }">
                  <strong style="color: #409eff">{{ row.contact_count || 0 }}</strong>
                </template>
              </el-table-column>
            </el-table>

            <!-- 分页 -->
            <div class="pagination">
              <el-pagination
                v-model:current-page="summaryPagination.page"
                v-model:page-size="summaryPagination.pageSize"
                :page-sizes="[10, 20, 50, 100]"
                :total="summaryPagination.total"
                layout="total, sizes, prev, pager, next, jumper"
                @size-change="handleSummarySizeChange"
                @current-change="handleSummaryPageChange"
              />
            </div>
          </div>
        </el-tab-pane>

        <!-- 详细列表标签页 -->
        <el-tab-pane label="详细列表" name="detail">
          <div class="tab-content">
            <!-- 筛选区域 -->
            <div class="filter-section">
              <el-form :model="detailFilterForm" :inline="true" class="filter-form">
                <el-form-item label="激活码">
                  <el-select
                    v-model="detailFilterForm.activation_code"
                    placeholder="全部"
                    clearable
                    filterable
                    style="width: 200px"
                  >
                    <el-option
                      v-for="group in groupList"
                      :key="group.id"
                      :label="`${group.activation_code}${group.remark ? ' - ' + group.remark : ''}`"
                      :value="group.activation_code"
                    />
                  </el-select>
                </el-form-item>
                <el-form-item label="平台">
                  <el-select
                    v-model="detailFilterForm.platform_type"
                    placeholder="全部"
                    clearable
                    style="width: 150px"
                  >
                    <el-option label="Line" value="line" />
                    <el-option label="Line Business" value="line_business" />
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
                    value-format="YYYY-MM-DD HH:mm:ss"
                    style="width: 400px"
                    @change="handleDateRangeChange"
                  />
                </el-form-item>
                <el-form-item label="搜索">
                  <el-input
                    v-model="detailFilterForm.search"
                    placeholder="用户名或手机号"
                    clearable
                    style="width: 200px"
                    @keyup.enter="handleDetailSearch"
                  >
                    <template #prefix>
                      <el-icon><Search /></el-icon>
                    </template>
                  </el-input>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="handleDetailSearch" :loading="loading">
                    <el-icon><Search /></el-icon>
                    搜索
                  </el-button>
                  <el-button @click="handleDetailReset">重置</el-button>
                </el-form-item>
              </el-form>
            </div>

            <!-- 操作按钮区域 -->
            <div class="action-buttons">
              <el-button type="primary" :disabled="loading" @click="handleImport">
                <el-icon><Upload /></el-icon>
                导入联系人
              </el-button>
            </div>

            <!-- 数据表格 -->
            <el-table
              v-loading="loading"
              :data="detailTableData"
              style="width: 100%"
              stripe
              empty-text="暂无数据"
            >
              <el-table-column prop="line_id" label="Line ID" width="180" show-overflow-tooltip />
              <el-table-column prop="display_name" label="显示名称" width="150" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ row.display_name || '-' }}
                </template>
              </el-table-column>
              <el-table-column prop="phone_number" label="手机号" width="120">
                <template #default="{ row }">
                  {{ row.phone_number || '-' }}
                </template>
              </el-table-column>
              <el-table-column prop="source" label="来源" width="120">
                <template #default="{ row }">
                  <el-tag :type="row.source === '系统上报' ? 'success' : 'info'" size="small">
                    {{ row.source || '-' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="180">
                <template #default="{ row }">
                  {{ formatDateTime(row.created_at) }}
                </template>
              </el-table-column>
            </el-table>

            <!-- 分页 -->
            <div class="pagination">
              <el-pagination
                v-model:current-page="detailPagination.page"
                v-model:page-size="detailPagination.pageSize"
                :page-sizes="[10, 20, 50, 100]"
                :total="detailPagination.total"
                layout="total, sizes, prev, pager, next, jumper"
                @size-change="handleDetailSizeChange"
                @current-change="handleDetailPageChange"
              />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 导入对话框 -->
    <el-dialog
      v-model="importDialogVisible"
      title="导入联系人"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="importFormRef"
        :model="importForm"
        :rules="importFormRules"
        label-width="120px"
      >
        <el-form-item label="文件" prop="file">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            accept=".xlsx,.xls,.csv,.txt"
            drag
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                支持 Excel (.xlsx, .xls)、CSV (.csv)、TXT (.txt) 格式，文件大小不超过10MB
                <br />
                <el-link type="primary" @click="handleDownloadTemplate" style="margin-top: 8px">
                  <el-icon><Download /></el-icon>
                  下载导入模板
                </el-link>
              </div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="平台类型" prop="platform_type">
          <el-select v-model="importForm.platform_type" placeholder="请选择平台类型" style="width: 100%">
            <el-option label="Line" value="line" />
            <el-option label="Line Business" value="line_business" />
          </el-select>
        </el-form-item>
        <el-form-item label="分组" prop="group_id">
          <el-select
            v-model="importForm.group_id"
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
        <el-form-item label="去重范围" prop="dedup_scope">
          <el-select v-model="importForm.dedup_scope" placeholder="请选择去重范围" style="width: 100%">
            <el-option label="当前分组" value="current" />
            <el-option label="全局" value="global" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleImportSubmit" :loading="importing">
          确定导入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Upload, UploadFilled, Download } from '@element-plus/icons-vue'
import {
  getContactPoolSummary,
  getContactPoolList,
  getContactPoolDetail,
  importContacts,
  getImportBatches,
  downloadImportTemplate
} from '@/api/contactPool'
import { getGroups } from '@/api/group'
import { formatDateTime } from '@/utils/format'

// 数据
const loading = ref(false)
const importing = ref(false)
const activeTab = ref('summary')
const summary = ref({})
const summaryTableData = ref([])
const detailTableData = ref([])
const groupList = ref([])
const dateRange = ref(null)

// 统计汇总筛选表单
const summaryFilterForm = reactive({
  platform_type: '',
  search: ''
})

// 详细列表筛选表单
const detailFilterForm = reactive({
  activation_code: '',
  platform_type: '',
  start_time: '',
  end_time: '',
  search: ''
})

// 分页
const summaryPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const detailPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 导入对话框
const importDialogVisible = ref(false)
const importFormRef = ref(null)
const uploadRef = ref(null)
const importForm = reactive({
  file: null,
  platform_type: '',
  group_id: null,
  dedup_scope: ''
})

// 导入表单验证规则
const importFormRules = {
  file: [
    { required: true, message: '请选择要上传的文件', trigger: 'change' }
  ],
  platform_type: [
    { required: true, message: '请选择平台类型', trigger: 'change' }
  ],
  group_id: [
    { required: true, message: '请选择分组', trigger: 'change' }
  ],
  dedup_scope: [
    { required: true, message: '请选择去重范围', trigger: 'change' }
  ]
}

// 加载统计汇总
const loadSummary = async () => {
  try {
    const res = await getContactPoolSummary()
    if (res.code === 1000) {
      summary.value = res.data || {}
    }
  } catch (error) {
    console.error('获取统计汇总失败:', error)
  }
}

// 加载分组列表
const loadGroups = async () => {
  try {
    const res = await getGroups({ page: 1, page_size: 100 })
    if (res.code === 1000) {
      groupList.value = Array.isArray(res.data?.list) ? res.data.list : []
    }
  } catch (error) {
    console.error('加载分组列表失败:', error)
  }
}

// 加载统计汇总列表
const loadSummaryList = async () => {
  loading.value = true
  try {
    const params = {
      page: summaryPagination.page,
      page_size: summaryPagination.pageSize,
      ...summaryFilterForm
    }

    // 移除空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })

    const res = await getContactPoolList(params)
    if (res.code === 1000) {
      summaryTableData.value = Array.isArray(res.data?.list) ? res.data.list : []
      summaryPagination.total = Number(res.data?.pagination?.total) || 0
    } else {
      summaryTableData.value = []
      summaryPagination.total = 0
    }
  } catch (error) {
    console.error('加载统计汇总列表失败:', error)
    ElMessage.error('加载统计汇总列表失败')
  } finally {
    loading.value = false
  }
}

// 加载详细列表
const loadDetailList = async () => {
  loading.value = true
  try {
    const params = {
      page: detailPagination.page,
      page_size: detailPagination.pageSize,
      ...detailFilterForm
    }

    // 移除空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })

    const res = await getContactPoolDetail(params)
    if (res.code === 1000) {
      detailTableData.value = Array.isArray(res.data?.list) ? res.data.list : []
      detailPagination.total = Number(res.data?.pagination?.total) || 0
    } else {
      detailTableData.value = []
      detailPagination.total = 0
    }
  } catch (error) {
    console.error('加载详细列表失败:', error)
    ElMessage.error('加载详细列表失败')
  } finally {
    loading.value = false
  }
}

// 标签页切换
const handleTabChange = (tabName) => {
  if (tabName === 'summary') {
    loadSummaryList()
  } else if (tabName === 'detail') {
    loadDetailList()
  }
}

// 统计汇总搜索
const handleSummarySearch = () => {
  summaryPagination.page = 1
  loadSummaryList()
}

// 统计汇总重置
const handleSummaryReset = () => {
  summaryFilterForm.platform_type = ''
  summaryFilterForm.search = ''
  summaryPagination.page = 1
  loadSummaryList()
}

// 统计汇总分页变化
const handleSummarySizeChange = (size) => {
  summaryPagination.pageSize = size
  summaryPagination.page = 1
  loadSummaryList()
}

const handleSummaryPageChange = (page) => {
  summaryPagination.page = page
  loadSummaryList()
}

// 详细列表搜索
const handleDetailSearch = () => {
  detailPagination.page = 1
  loadDetailList()
}

// 详细列表重置
const handleDetailReset = () => {
  detailFilterForm.activation_code = ''
  detailFilterForm.platform_type = ''
  detailFilterForm.start_time = ''
  detailFilterForm.end_time = ''
  detailFilterForm.search = ''
  dateRange.value = null
  detailPagination.page = 1
  loadDetailList()
}

// 时间范围变化
const handleDateRangeChange = (dates) => {
  if (dates && dates.length === 2) {
    detailFilterForm.start_time = dates[0]
    detailFilterForm.end_time = dates[1]
  } else {
    detailFilterForm.start_time = ''
    detailFilterForm.end_time = ''
  }
}

// 详细列表分页变化
const handleDetailSizeChange = (size) => {
  detailPagination.pageSize = size
  detailPagination.page = 1
  loadDetailList()
}

const handleDetailPageChange = (page) => {
  detailPagination.page = page
  loadDetailList()
}

// 打开导入对话框
const handleImport = () => {
  importForm.file = null
  importForm.platform_type = ''
  importForm.group_id = null
  importForm.dedup_scope = ''
  if (uploadRef.value) {
    uploadRef.value.clearFiles()
  }
  importDialogVisible.value = true
}

// 文件变化
const handleFileChange = (file) => {
  importForm.file = file.raw
}

// 文件移除
const handleFileRemove = () => {
  importForm.file = null
}

// 下载导入模板
const handleDownloadTemplate = async () => {
  try {
    const res = await downloadImportTemplate()
    // 创建blob对象
    const blob = new Blob([res], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    })
    // 创建下载链接
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = '联系人导入模板.xlsx'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ElMessage.success('模板下载成功')
  } catch (error) {
    console.error('下载模板失败:', error)
    ElMessage.error('下载模板失败')
  }
}

// 提交导入
const handleImportSubmit = async () => {
  if (!importFormRef.value) return

  await importFormRef.value.validate(async (valid) => {
    if (!valid) return

    if (!importForm.file) {
      ElMessage.warning('请选择要上传的文件')
      return
    }

    importing.value = true
    try {
      const formData = new FormData()
      formData.append('file', importForm.file)
      formData.append('platform_type', importForm.platform_type)
      formData.append('group_id', importForm.group_id)
      formData.append('dedup_scope', importForm.dedup_scope)

      const res = await importContacts(formData)
      if (res.code === 1000) {
        ElMessage.success('导入成功')
        importDialogVisible.value = false
        
        // 刷新数据
        loadSummary()
        if (activeTab.value === 'summary') {
          loadSummaryList()
        } else if (activeTab.value === 'detail') {
          loadDetailList()
        }
      } else {
        ElMessage.error(res.message || '导入失败')
      }
    } catch (error) {
      console.error('导入失败:', error)
      ElMessage.error(error.response?.data?.message || '导入失败')
    } finally {
      importing.value = false
    }
  })
}

// 初始化
onMounted(() => {
  loadSummary()
  loadGroups()
  loadSummaryList()
})
</script>

<style scoped lang="less">
@import '@/styles/list-page.less';

.contact-pool {
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

  .tab-content {
    // 标签页内容样式已在list-page.less中定义
  }
}
</style>
