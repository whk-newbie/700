<template>
  <div class="login-container">
    <div class="login-form">
      <div class="login-header">
        <h2>Line账号管理系统</h2>
        <p>管理员登录</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="form"
        :rules="rules"
        label-width="0"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            size="large"
            :prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            style="width: 100%"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <el-link type="primary" @click="showSubAccountLogin = true">
          子账号登录
        </el-link>
      </div>
    </div>

    <!-- 子账号登录弹窗 -->
    <el-dialog
      v-model="showSubAccountLogin"
      title="子账号登录"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="subAccountFormRef"
        :model="subAccountForm"
        :rules="subAccountRules"
        label-width="0"
      >
        <el-form-item prop="activation_code">
          <el-input
            v-model="subAccountForm.activation_code"
            placeholder="请输入激活码"
            size="large"
            clearable
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="subAccountForm.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            show-password
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showSubAccountLogin = false">取消</el-button>
        <el-button type="primary" :loading="loading" @click="handleSubAccountLogin">
          登录
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()

// 表单数据
const form = reactive({
  username: '',
  password: ''
})

const subAccountForm = reactive({
  activation_code: '',
  password: ''
})

// 表单引用
const loginFormRef = ref()
const subAccountFormRef = ref()

// 状态
const loading = ref(false)
const showSubAccountLogin = ref(false)

// 验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

const subAccountRules = {
  activation_code: [
    { required: true, message: '请输入激活码', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

// 登录处理
const handleLogin = async () => {
  if (!loginFormRef.value) return

  try {
    await loginFormRef.value.validate()
    loading.value = true

    await authStore.login(form)

    ElMessage.success('登录成功')
  } catch (error) {
    if (error.message !== 'Validation failed') {
      ElMessage.error(error.message || '登录失败')
    }
  } finally {
    loading.value = false
  }
}

// 子账号登录处理
const handleSubAccountLogin = async () => {
  if (!subAccountFormRef.value) return

  try {
    await subAccountFormRef.value.validate()
    loading.value = true

    await authStore.subAccountLogin(subAccountForm)

    ElMessage.success('登录成功')
    showSubAccountLogin.value = false
  } catch (error) {
    if (error.message !== 'Validation failed') {
      ElMessage.error(error.message || '登录失败')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style lang="less" scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, @primary-color 0%, lighten(@primary-color, 20%) 100%);
  padding: 20px;
}

.login-form {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: @border-radius-base * 3;
  padding: 40px;
  box-shadow: @box-shadow-dark;

  .login-header {
    text-align: center;
    margin-bottom: 30px;

    h2 {
      color: @text-primary;
      margin-bottom: 8px;
      font-size: 24px;
      font-weight: 600;
    }

    p {
      color: @text-secondary;
      margin: 0;
      font-size: 14px;
    }
  }

  .login-footer {
    text-align: center;
    margin-top: 20px;
  }
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}

:deep(.el-input__inner) {
  border-radius: @border-radius-base;
  height: 48px;
}

:deep(.el-button) {
  height: 48px;
  border-radius: @border-radius-base;
}
</style>
