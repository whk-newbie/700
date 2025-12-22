<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h2>Line账号管理系统</h2>
        <p>子账号登录</p>
      </div>
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
      >
        <el-form-item prop="activationCode">
          <el-input
            v-model="loginForm.activationCode"
            placeholder="请输入激活码"
            size="large"
            clearable
          >
            <template #prefix>
              <el-icon><Key /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码（如果分组设置了密码）"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
          <div class="password-hint">
            <el-text type="info" size="small">如果分组未设置密码，可直接使用激活码登录</el-text>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="login-button"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '登录' }}
          </el-button>
        </el-form-item>
      </el-form>
      <div class="login-footer">
        <el-link type="primary" @click="goToAdminLogin">
          管理员/普通用户登录
        </el-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { Key, Lock } from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const loginFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  activationCode: '',
  password: ''
})

const loginRules = {
  activationCode: [
    { required: true, message: '请输入激活码', trigger: 'blur' }
  ],
  password: [
    // 密码是可选的，如果分组设置了密码则需要输入
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  // 只验证激活码，密码可选
  await loginFormRef.value.validateField('activationCode', async (valid) => {
    if (!valid) return
    
    // 如果输入了密码，验证密码长度
    if (loginForm.password && loginForm.password.length > 0) {
      if (loginForm.password.length < 6) {
        ElMessage.warning('密码长度不能少于6位')
        return
      }
    }
    
    loading.value = true
    try {
      // 如果密码为空，传空字符串
      const password = loginForm.password || ''
      const result = await authStore.loginSubAccount(
        loginForm.activationCode,
        password
      )
      
      if (result.success) {
        ElMessage.success('登录成功')
        // 跳转到子账号首页
        router.push('/subaccount/dashboard')
      } else {
        ElMessage.error(result.message || '登录失败')
      }
    } catch (error) {
      ElMessage.error(error.message || '登录失败')
    } finally {
      loading.value = false
    }
  })
}

const goToAdminLogin = () => {
  router.push({ name: 'Login' })
}
</script>

<style scoped lang="less">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
  
  h2 {
    font-size: 24px;
    color: #303133;
    margin-bottom: 8px;
  }
  
  p {
    font-size: 14px;
    color: #909399;
  }
}

.login-form {
  .el-form-item {
    margin-bottom: 20px;
  }
  
  .password-hint {
    margin-top: 8px;
    text-align: center;
  }
  
  .login-button {
    width: 100%;
    margin-top: 10px;
  }
}

.login-footer {
  text-align: center;
  margin-top: 20px;
}
</style>

