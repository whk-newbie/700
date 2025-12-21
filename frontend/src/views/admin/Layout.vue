<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside width="200px" class="layout-sidebar">
      <div class="sidebar-header">
        <h3>Line管理系统</h3>
      </div>

      <el-menu
        :default-active="$route.path"
        class="sidebar-menu"
        :collapse="sidebarCollapsed"
        router
      >
        <el-menu-item
          v-for="item in menuItems"
          :key="item.path"
          :index="item.path"
          :disabled="!hasPermission(item.roles)"
        >
          <component :is="item.icon" />
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区域 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="layout-header">
        <div class="header-left">
          <el-button
            type="text"
            @click="toggleSidebar"
            class="sidebar-toggle"
          >
            <Fold v-if="!sidebarCollapsed" />
            <Expand v-else />
          </el-button>
        </div>

        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar size="small" :src="user?.avatar">
                {{ user?.username?.charAt(0)?.toUpperCase() }}
              </el-avatar>
              <span class="username">{{ user?.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                <el-dropdown-item command="change-password">修改密码</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Fold,
  Expand,
  HomeFilled,
  FolderOpened,
  Monitor,
  DataLine,
  Files,
  User,
  DocumentCopy,
  UserFilled,
  Setting,
  Document,
  ArrowDown
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()

// 状态
const sidebarCollapsed = ref(false)

// 用户信息
const user = computed(() => authStore.user)

// 菜单项配置
const menuItems = [
  {
    path: '/dashboard',
    title: '首页',
    icon: HomeFilled,
    roles: ['admin', 'user', 'subaccount']
  },
  {
    path: '/groups',
    title: '分组管理',
    icon: FolderOpened,
    roles: ['admin', 'user']
  },
  {
    path: '/accounts',
    title: '账号列表',
    icon: Monitor,
    roles: ['admin', 'user']
  },
  {
    path: '/leads',
    title: '线索列表',
    icon: DataLine,
    roles: ['admin', 'user', 'subaccount']
  },
  {
    path: '/contact-pool',
    title: '底库管理',
    icon: Files,
    roles: ['admin', 'user']
  },
  {
    path: '/customers',
    title: '客户列表',
    icon: User,
    roles: ['admin', 'user', 'subaccount']
  },
  {
    path: '/follow-ups',
    title: '跟进记录',
    icon: DocumentCopy,
    roles: ['admin', 'user', 'subaccount']
  },
  {
    path: '/users',
    title: '用户管理',
    icon: UserFilled,
    roles: ['admin']
  },
  {
    path: '/llm-config',
    title: '大模型配置',
    icon: Setting,
    roles: ['admin']
  },
  {
    path: '/llm-logs',
    title: '调用记录',
    icon: Document,
    roles: ['admin']
  }
]

// 权限检查
const hasPermission = (roles) => {
  if (!roles) return true
  return roles.includes(authStore.userRole)
}

// 切换侧边栏
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// 处理下拉菜单命令
const handleCommand = (command) => {
  switch (command) {
    case 'profile':
      // TODO: 个人资料页面
      ElMessage.info('个人资料功能开发中')
      break
    case 'change-password':
      // TODO: 修改密码弹窗
      ElMessage.info('修改密码功能开发中')
      break
    case 'logout':
      handleLogout()
      break
  }
}

// 退出登录
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    authStore.logout()
    ElMessage.success('已退出登录')
  } catch {
    // 用户取消操作
  }
}

// 初始化
onMounted(() => {
  // 检查用户权限，过滤菜单项
  // 可以在这里根据用户角色动态调整菜单
})
</script>

<style lang="less" scoped>
.layout-container {
  height: 100vh;

  .layout-sidebar {
    background-color: #fff;
    border-right: 1px solid @border-color-lighter;
    transition: width @animation-duration;

    .sidebar-header {
      height: 60px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-bottom: 1px solid @border-color-lighter;
      background-color: @primary-color;
      color: #fff;

      h3 {
        margin: 0;
        font-size: 16px;
        font-weight: 600;
      }
    }

    .sidebar-menu {
      border-right: none;
      margin-top: 10px;
    }
  }

  .layout-header {
    background-color: #fff;
    border-bottom: 1px solid @border-color-lighter;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 20px;

    .header-left {
      .sidebar-toggle {
        color: @text-regular;
      }
    }

    .header-right {
      .user-info {
        display: flex;
        align-items: center;
        cursor: pointer;
        padding: 8px 12px;
        border-radius: @border-radius-base;
        transition: background-color @animation-duration-fast;

        &:hover {
          background-color: @background-color-page;
        }

        .username {
          margin: 0 8px;
          font-size: 14px;
          color: @text-regular;
        }
      }
    }
  }

  .layout-main {
    background-color: @background-color-base;
    padding: 20px;
    overflow-y: auto;
  }
}

:deep(.el-menu-item) {
  height: 48px;
  line-height: 48px;

  &.is-disabled {
    opacity: 0.6;
  }
}
</style>
