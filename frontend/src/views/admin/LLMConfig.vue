<template>
  <div class="list-page-container llm-config">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>大模型配置</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 配置列表标签页 -->
        <el-tab-pane label="配置列表" name="configs">
          <!-- 筛选区域 -->
          <div class="filter-section">
            <el-form :model="filterForm" :inline="true" class="filter-form">
              <el-form-item label="提供商">
                <el-select
                  v-model="filterForm.provider"
                  placeholder="全部"
                  clearable
                  style="width: 150px"
                >
                  <el-option label="OpenAI" value="openai" />
                  <el-option label="Anthropic" value="anthropic" />
                  <el-option label="阿里云" value="aliyun" />
                  <el-option label="讯飞" value="xunfei" />
                  <el-option label="百度" value="baidu" />
                  <el-option label="智谱" value="zhipu" />
                  <el-option label="自定义" value="custom" />
                </el-select>
              </el-form-item>
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
                  placeholder="配置名称"
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
                <el-button type="default" :disabled="loading" @click="handleRefresh">
                  <el-icon><Refresh /></el-icon>
                  刷新
                </el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- 操作按钮区域 -->
          <div class="action-buttons">
            <el-button type="primary" :disabled="loading" @click="handleAddConfig">
              <el-icon><Plus /></el-icon>
              新增配置
            </el-button>
          </div>

          <!-- 配置列表表格 -->
          <el-table
            v-loading="loading"
            :data="configTableData"
            style="width: 100%"
            stripe
            empty-text="暂无数据"
          >
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="name" label="配置名称" width="200" show-overflow-tooltip />
            <el-table-column prop="provider" label="提供商" width="120">
              <template #default="{ row }">
                <el-tag>{{ getProviderName(row.provider) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="model" label="模型" width="200" show-overflow-tooltip />
            <el-table-column prop="api_url" label="API地址" width="250" show-overflow-tooltip>
              <template #default="{ row }">
                <el-link :href="row.api_url" target="_blank" type="primary" :underline="false">
                  {{ row.api_url }}
                </el-link>
              </template>
            </el-table-column>
            <el-table-column prop="is_active" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.is_active ? 'success' : 'danger'">
                  {{ row.is_active ? '激活' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="250" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleEditConfig(row)">
                  编辑
                </el-button>
                <el-button type="success" link size="small" @click="handleTestConfig(row)">
                  测试
                </el-button>
                <el-button type="danger" link size="small" @click="handleDeleteConfig(row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="pagination" v-if="configPagination.total > 0">
            <el-pagination
              v-model:current-page="configPagination.page"
              v-model:page-size="configPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="Number(configPagination.total)"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleConfigSizeChange"
              @current-change="handleConfigPageChange"
            />
          </div>
        </el-tab-pane>

        <!-- 模板列表标签页 -->
        <el-tab-pane label="Prompt模板" name="templates">
          <!-- 模板筛选 -->
          <div class="filter-section">
            <el-form :model="templateFilterForm" :inline="true" class="filter-form">
              <el-form-item label="配置">
                <el-select
                  v-model="templateFilterForm.config_id"
                  placeholder="全部"
                  clearable
                  filterable
                  style="width: 250px"
                >
                  <el-option
                    v-for="config in configList"
                    :key="config.id"
                    :label="config.name"
                    :value="config.id"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="状态">
                <el-select
                  v-model="templateFilterForm.is_active"
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
                  v-model="templateFilterForm.search"
                  placeholder="模板名称"
                  clearable
                  style="width: 200px"
                  @keyup.enter="handleTemplateSearch"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleTemplateSearch" :loading="templateLoading">
                  <el-icon><Search /></el-icon>
                  搜索
                </el-button>
                <el-button @click="handleTemplateReset">重置</el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- 操作按钮区域 -->
          <div class="action-buttons">
            <el-button type="primary" :disabled="templateLoading" @click="handleAddTemplate">
              <el-icon><Plus /></el-icon>
              新增模板
            </el-button>
          </div>

          <!-- 模板列表表格 -->
          <el-table
            v-loading="templateLoading"
            :data="templateTableData"
            style="width: 100%"
            stripe
            empty-text="暂无数据"
          >
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="template_name" label="模板名称" width="200" show-overflow-tooltip />
            <el-table-column prop="config_id" label="配置" width="150">
              <template #default="{ row }">
                {{ getConfigName(row.config_id) }}
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" width="250" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.description || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="is_active" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.is_active ? 'success' : 'danger'">
                  {{ row.is_active ? '激活' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleEditTemplate(row)">
                  编辑
                </el-button>
                <el-button type="danger" link size="small" @click="handleDeleteTemplate(row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="pagination" v-if="templatePagination.total > 0">
            <el-pagination
              v-model:current-page="templatePagination.page"
              v-model:page-size="templatePagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="Number(templatePagination.total)"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleTemplateSizeChange"
              @current-change="handleTemplatePageChange"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 配置对话框 -->
    <el-dialog
      v-model="configDialogVisible"
      :title="configDialogTitle"
      width="800px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="configFormRef"
        :model="configFormData"
        :rules="configFormRules"
        label-width="140px"
      >
        <el-form-item label="配置名称" prop="name">
          <el-input v-model="configFormData.name" placeholder="请输入配置名称" />
        </el-form-item>
        <el-form-item label="提供商" prop="provider">
          <el-select v-model="configFormData.provider" placeholder="请选择提供商" style="width: 100%">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Anthropic" value="anthropic" />
            <el-option label="阿里云" value="aliyun" />
            <el-option label="讯飞" value="xunfei" />
            <el-option label="百度" value="baidu" />
            <el-option label="智谱" value="zhipu" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="API地址" prop="api_url">
          <el-input v-model="configFormData.api_url" placeholder="请输入API地址" />
        </el-form-item>
        <el-form-item label="API Key" prop="api_key">
          <el-input
            v-model="configFormData.api_key"
            type="password"
            :placeholder="isEditConfig ? '留空则不修改' : '请输入API Key'"
            show-password
          />
        </el-form-item>
        <el-form-item label="模型" prop="model">
          <el-input v-model="configFormData.model" placeholder="请输入模型名称" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="Max Tokens" prop="max_tokens">
              <el-input-number
                v-model="configFormData.max_tokens"
                :min="1"
                :max="100000"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Temperature" prop="temperature">
              <el-input-number
                v-model="configFormData.temperature"
                :min="0"
                :max="2"
                :step="0.1"
                :precision="2"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="Top P" prop="top_p">
              <el-input-number
                v-model="configFormData.top_p"
                :min="0"
                :max="1"
                :step="0.1"
                :precision="2"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Frequency Penalty" prop="frequency_penalty">
              <el-input-number
                v-model="configFormData.frequency_penalty"
                :min="-2"
                :max="2"
                :step="0.1"
                :precision="2"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="Presence Penalty" prop="presence_penalty">
              <el-input-number
                v-model="configFormData.presence_penalty"
                :min="-2"
                :max="2"
                :step="0.1"
                :precision="2"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="超时时间(秒)" prop="timeout_seconds">
              <el-input-number
                v-model="configFormData.timeout_seconds"
                :min="1"
                :max="300"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="最大重试次数" prop="max_retries">
          <el-input-number
            v-model="configFormData.max_retries"
            :min="0"
            :max="10"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="System Prompt">
          <el-input
            v-model="configFormData.system_prompt"
            type="textarea"
            :rows="4"
            placeholder="请输入系统提示词（可选）"
          />
        </el-form-item>
        <el-form-item label="状态" prop="is_active">
          <el-switch
            v-model="configFormData.is_active"
            active-text="激活"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfigSubmit" :loading="configSaving">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 模板对话框 -->
    <el-dialog
      v-model="templateDialogVisible"
      :title="templateDialogTitle"
      width="800px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="templateFormRef"
        :model="templateFormData"
        :rules="templateFormRules"
        label-width="140px"
      >
        <el-form-item label="模板名称" prop="template_name">
          <el-input v-model="templateFormData.template_name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="配置" prop="config_id">
          <el-select
            v-model="templateFormData.config_id"
            placeholder="请选择配置"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="config in configList"
              :key="config.id"
              :label="config.name"
              :value="config.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模板内容" prop="template_content">
          <el-input
            v-model="templateFormData.template_content"
            type="textarea"
            :rows="8"
            placeholder="请输入模板内容，支持变量替换，如：{{variable_name}}"
          />
        </el-form-item>
        <el-form-item label="变量定义">
          <el-input
            v-model="templateVariablesText"
            type="textarea"
            :rows="4"
            placeholder='请输入JSON格式的变量定义，如：{"variable_name": "变量描述"}'
            @blur="handleVariablesChange"
          />
          <div class="form-tip">JSON格式：{"变量名": "变量描述"}</div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="templateFormData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入模板描述（可选）"
          />
        </el-form-item>
        <el-form-item label="状态" prop="is_active">
          <el-switch
            v-model="templateFormData.is_active"
            active-text="激活"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="templateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleTemplateSubmit" :loading="templateSaving">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  getLLMConfigs,
  createLLMConfig,
  updateLLMConfig,
  deleteLLMConfig,
  testLLMConfig,
  getLLMTemplates,
  createLLMTemplate,
  updateLLMTemplate,
  deleteLLMTemplate
} from '@/api/llm'
import { formatDateTime } from '@/utils/format'

// 标签页
const activeTab = ref('configs')

// 配置相关
const loading = ref(false)
const configTableData = ref([])
const configList = ref([])
const configPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})
const filterForm = reactive({
  provider: '',
  is_active: null,
  search: ''
})

// 模板相关
const templateLoading = ref(false)
const templateTableData = ref([])
const templatePagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})
const templateFilterForm = reactive({
  config_id: null,
  is_active: null,
  search: ''
})

// 配置对话框
const configDialogVisible = ref(false)
const configDialogTitle = ref('新增配置')
const isEditConfig = ref(false)
const configFormRef = ref(null)
const configSaving = ref(false)
const configFormData = reactive({
  name: '',
  provider: 'openai',
  api_url: '',
  api_key: '',
  model: '',
  max_tokens: 2000,
  temperature: 0.7,
  top_p: 1.0,
  frequency_penalty: 0.0,
  presence_penalty: 0.0,
  system_prompt: '',
  timeout_seconds: 30,
  max_retries: 3,
  is_active: true
})

const configFormRules = {
  name: [
    { required: true, message: '请输入配置名称', trigger: 'blur' }
  ],
  provider: [
    { required: true, message: '请选择提供商', trigger: 'change' }
  ],
  api_url: [
    { required: true, message: '请输入API地址', trigger: 'blur' },
    { type: 'url', message: '请输入正确的URL', trigger: 'blur' }
  ],
  api_key: [
    { required: true, message: '请输入API Key', trigger: 'blur' }
  ],
  model: [
    { required: true, message: '请输入模型名称', trigger: 'blur' }
  ]
}

// 模板对话框
const templateDialogVisible = ref(false)
const templateDialogTitle = ref('新增模板')
const isEditTemplate = ref(false)
const templateFormRef = ref(null)
const templateSaving = ref(false)
const templateFormData = reactive({
  template_name: '',
  config_id: null,
  template_content: '',
  variables: {},
  description: '',
  is_active: true
})
const templateVariablesText = ref('')

const templateFormRules = {
  template_name: [
    { required: true, message: '请输入模板名称', trigger: 'blur' }
  ],
  config_id: [
    { required: true, message: '请选择配置', trigger: 'change' }
  ],
  template_content: [
    { required: true, message: '请输入模板内容', trigger: 'blur' }
  ]
}

// 获取提供商名称
const getProviderName = (provider) => {
  const map = {
    openai: 'OpenAI',
    anthropic: 'Anthropic',
    aliyun: '阿里云',
    xunfei: '讯飞',
    baidu: '百度',
    zhipu: '智谱',
    custom: '自定义'
  }
  return map[provider] || provider
}

// 获取配置名称
const getConfigName = (configId) => {
  const config = configList.value.find(c => c.id === configId)
  return config ? config.name : `配置ID: ${configId}`
}

// 获取配置列表
const fetchConfigs = async () => {
  loading.value = true
  try {
    const params = {
      page: configPagination.page,
      page_size: configPagination.pageSize,
      ...filterForm
    }
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })

    const res = await getLLMConfigs(params)
    if (res.code === 1000) {
      configTableData.value = res.data.list || []
      configPagination.total = res.data.pagination?.total || 0
      
      // 同时更新配置下拉列表（用于模板选择）
      if (configPagination.page === 1) {
        const allRes = await getLLMConfigs({ page: 1, page_size: 100 })
        if (allRes.code === 1000) {
          configList.value = allRes.data.list || []
        }
      }
    } else {
      ElMessage.error(res.message || '获取配置列表失败')
    }
  } catch (error) {
    console.error('获取配置列表失败:', error)
    ElMessage.error('获取配置列表失败')
  } finally {
    loading.value = false
  }
}

// 获取模板列表
const fetchTemplates = async () => {
  templateLoading.value = true
  try {
    const params = {
      page: templatePagination.page,
      page_size: templatePagination.pageSize,
      ...templateFilterForm
    }
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })

    const res = await getLLMTemplates(params)
    if (res.code === 1000) {
      templateTableData.value = res.data.list || []
      templatePagination.total = res.data.pagination?.total || 0
    } else {
      ElMessage.error(res.message || '获取模板列表失败')
    }
  } catch (error) {
    console.error('获取模板列表失败:', error)
    ElMessage.error('获取模板列表失败')
  } finally {
    templateLoading.value = false
  }
}

// 标签页切换
const handleTabChange = (tabName) => {
  if (tabName === 'configs') {
    fetchConfigs()
  } else if (tabName === 'templates') {
    fetchTemplates()
    // 确保配置列表已加载
    if (configList.value.length === 0) {
      fetchConfigs()
    }
  }
}

// 配置相关操作
const handleSearch = () => {
  configPagination.page = 1
  fetchConfigs()
}

const handleReset = () => {
  filterForm.provider = ''
  filterForm.is_active = null
  filterForm.search = ''
  configPagination.page = 1
  fetchConfigs()
}

const handleRefresh = () => {
  if (activeTab.value === 'configs') {
    fetchConfigs()
  } else {
    fetchTemplates()
  }
}

const handleConfigSizeChange = (size) => {
  configPagination.pageSize = size
  configPagination.page = 1
  fetchConfigs()
}

const handleConfigPageChange = (page) => {
  configPagination.page = page
  fetchConfigs()
}

const handleAddConfig = () => {
  isEditConfig.value = false
  configDialogTitle.value = '新增配置'
  resetConfigForm()
  configDialogVisible.value = true
}

const handleEditConfig = (row) => {
  isEditConfig.value = true
  configDialogTitle.value = '编辑配置'
  Object.assign(configFormData, {
    id: row.id,
    name: row.name,
    provider: row.provider,
    api_url: row.api_url,
    api_key: '', // 不显示API Key
    model: row.model,
    max_tokens: row.max_tokens,
    temperature: row.temperature,
    top_p: row.top_p,
    frequency_penalty: row.frequency_penalty,
    presence_penalty: row.presence_penalty,
    system_prompt: row.system_prompt || '',
    timeout_seconds: row.timeout_seconds,
    max_retries: row.max_retries,
    is_active: row.is_active
  })
  configDialogVisible.value = true
}

const handleDeleteConfig = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除配置 "${row.name}" 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await deleteLLMConfig(row.id)
    if (res.code === 1000) {
      ElMessage.success('删除成功')
      fetchConfigs()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除配置失败:', error)
      ElMessage.error('删除配置失败')
    }
  }
}

const handleTestConfig = async (row) => {
  try {
    ElMessage.info('正在测试连接...')
    const res = await testLLMConfig(row.id)
    if (res.code === 1000) {
      ElMessage.success('连接测试成功')
    } else {
      ElMessage.error(res.message || '连接测试失败')
    }
  } catch (error) {
    console.error('测试连接失败:', error)
    ElMessage.error('测试连接失败')
  }
}

const resetConfigForm = () => {
  Object.assign(configFormData, {
    id: null,
    name: '',
    provider: 'openai',
    api_url: '',
    api_key: '',
    model: '',
    max_tokens: 2000,
    temperature: 0.7,
    top_p: 1.0,
    frequency_penalty: 0.0,
    presence_penalty: 0.0,
    system_prompt: '',
    timeout_seconds: 30,
    max_retries: 3,
    is_active: true
  })
  if (configFormRef.value) {
    configFormRef.value.clearValidate()
  }
}

const handleConfigSubmit = async () => {
  if (!configFormRef.value) return

  await configFormRef.value.validate(async (valid) => {
    if (!valid) return

    configSaving.value = true
    try {
      const submitData = { ...configFormData }
      
      // 编辑时，如果API Key为空，则不传
      if (isEditConfig.value && !submitData.api_key) {
        delete submitData.api_key
      }

      let res
      if (isEditConfig.value) {
        const { id, ...updateData } = submitData
        res = await updateLLMConfig(id, updateData)
      } else {
        res = await createLLMConfig(submitData)
      }

      if (res.code === 1000) {
        ElMessage.success(isEditConfig.value ? '更新成功' : '创建成功')
        configDialogVisible.value = false
        fetchConfigs()
      } else {
        ElMessage.error(res.message || (isEditConfig.value ? '更新失败' : '创建失败'))
      }
    } catch (error) {
      console.error('提交失败:', error)
      ElMessage.error(isEditConfig.value ? '更新失败' : '创建失败')
    } finally {
      configSaving.value = false
    }
  })
}

// 模板相关操作
const handleTemplateSearch = () => {
  templatePagination.page = 1
  fetchTemplates()
}

const handleTemplateReset = () => {
  templateFilterForm.config_id = null
  templateFilterForm.is_active = null
  templateFilterForm.search = ''
  templatePagination.page = 1
  fetchTemplates()
}

const handleTemplateSizeChange = (size) => {
  templatePagination.pageSize = size
  templatePagination.page = 1
  fetchTemplates()
}

const handleTemplatePageChange = (page) => {
  templatePagination.page = page
  fetchTemplates()
}

const handleAddTemplate = () => {
  isEditTemplate.value = false
  templateDialogTitle.value = '新增模板'
  resetTemplateForm()
  templateDialogVisible.value = true
}

const handleEditTemplate = (row) => {
  isEditTemplate.value = true
  templateDialogTitle.value = '编辑模板'
  Object.assign(templateFormData, {
    id: row.id,
    template_name: row.template_name,
    config_id: row.config_id,
    template_content: row.template_content,
    variables: row.variables || {},
    description: row.description || '',
    is_active: row.is_active
  })
  // 将variables对象转换为JSON字符串
  try {
    templateVariablesText.value = JSON.stringify(row.variables || {}, null, 2)
  } catch {
    templateVariablesText.value = ''
  }
  templateDialogVisible.value = true
}

const handleDeleteTemplate = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${row.template_name}" 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await deleteLLMTemplate(row.id)
    if (res.code === 1000) {
      ElMessage.success('删除成功')
      fetchTemplates()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除模板失败:', error)
      ElMessage.error('删除模板失败')
    }
  }
}

const resetTemplateForm = () => {
  Object.assign(templateFormData, {
    id: null,
    template_name: '',
    config_id: null,
    template_content: '',
    variables: {},
    description: '',
    is_active: true
  })
  templateVariablesText.value = ''
  if (templateFormRef.value) {
    templateFormRef.value.clearValidate()
  }
}

const handleVariablesChange = () => {
  try {
    if (templateVariablesText.value.trim()) {
      templateFormData.variables = JSON.parse(templateVariablesText.value)
    } else {
      templateFormData.variables = {}
    }
  } catch (error) {
    ElMessage.warning('变量定义格式错误，请检查JSON格式')
  }
}

const handleTemplateSubmit = async () => {
  if (!templateFormRef.value) return

  await templateFormRef.value.validate(async (valid) => {
    if (!valid) return

    // 处理变量定义
    handleVariablesChange()

    templateSaving.value = true
    try {
      const submitData = { ...templateFormData }

      let res
      if (isEditTemplate.value) {
        const { id, ...updateData } = submitData
        res = await updateLLMTemplate(id, updateData)
      } else {
        res = await createLLMTemplate(submitData)
      }

      if (res.code === 1000) {
        ElMessage.success(isEditTemplate.value ? '更新成功' : '创建成功')
        templateDialogVisible.value = false
        fetchTemplates()
      } else {
        ElMessage.error(res.message || (isEditTemplate.value ? '更新失败' : '创建失败'))
      }
    } catch (error) {
      console.error('提交失败:', error)
      ElMessage.error(isEditTemplate.value ? '更新失败' : '创建失败')
    } finally {
      templateSaving.value = false
    }
  })
}

// 初始化
onMounted(() => {
  fetchConfigs()
})
</script>

<style scoped lang="less">
@import '@/styles/list-page.less';

.llm-config {
  .card-header {
    .header-actions {
      display: flex;
      gap: 10px;
    }
  }

  .form-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 4px;
  }
}
</style>
