<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '200px'" class="sidebar">
      <div class="logo">
        <span v-if="!isCollapse">Line管理系统</span>
        <span v-else>L</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <template #title>首页</template>
        </el-menu-item>
        <el-menu-item index="/groups">
          <el-icon><Folder /></el-icon>
          <template #title>分组管理</template>
        </el-menu-item>
        <el-menu-item index="/accounts">
          <el-icon><User /></el-icon>
          <template #title>账号列表</template>
        </el-menu-item>
        <el-menu-item index="/leads">
          <el-icon><List /></el-icon>
          <template #title>线索列表</template>
        </el-menu-item>
        <el-menu-item index="/contact-pool">
          <el-icon><Box /></el-icon>
          <template #title>底库管理</template>
        </el-menu-item>
        <el-menu-item index="/customers">
          <el-icon><Avatar /></el-icon>
          <template #title>客户列表</template>
        </el-menu-item>
        <el-menu-item index="/follow-ups">
          <el-icon><Document /></el-icon>
          <template #title>跟进记录</template>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/users">
          <el-icon><Setting /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/llm-config">
          <el-icon><Cpu /></el-icon>
          <template #title>大模型配置</template>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/llm-logs">
          <el-icon><DataLine /></el-icon>
          <template #title>调用记录</template>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon @click="toggleCollapse" class="collapse-icon">
            <Expand v-if="isCollapse" />
            <Fold v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentTitle">{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><Avatar /></el-icon>
              <span>{{ user?.username || '用户' }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import {
  HomeFilled,
  Folder,
  User,
  List,
  Box,
  Avatar,
  Document,
  Setting,
  Cpu,
  DataLine,
  Expand,
  Fold,
  ArrowDown
} from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isCollapse = ref(false)
const user = computed(() => authStore.user)
const isAdmin = computed(() => authStore.isAdmin)

const activeMenu = computed(() => route.path)
const currentTitle = computed(() => route.meta?.title || '')

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleCommand = async (command) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      authStore.logout()
      router.push({ name: 'Login' })
    } catch {
      // 用户取消
    }
  }
}
</script>

<style scoped lang="less">
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  transition: width 0.3s;
  
  .logo {
    height: 60px;
    line-height: 60px;
    text-align: center;
    color: #fff;
    font-size: 18px;
    font-weight: bold;
    background-color: #2b3a4a;
  }
  
  .sidebar-menu {
    border: none;
    background-color: #304156;
    
    :deep(.el-menu-item) {
      color: #bfcbd9;
      
      &:hover {
        background-color: #263445;
      }
      
      &.is-active {
        background-color: #409eff;
        color: #fff;
      }
    }
  }
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
  
  .header-left {
    display: flex;
    align-items: center;
    gap: 20px;
    
    .collapse-icon {
      font-size: 20px;
      cursor: pointer;
    }
  }
  
  .header-right {
    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
    }
  }
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
}
</style>

