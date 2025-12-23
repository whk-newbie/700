<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '200px'" class="sidebar">
      <div class="logo">
        <span v-if="!isCollapse">Line管理系统</span>
        <span v-else>L</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :default-openeds="defaultOpenedMenus"
        :collapse="isCollapse"
        router
        class="sidebar-menu"
        :unique-opened="true"
      >
        <!-- 首页 -->
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <template #title>首页</template>
        </el-menu-item>

        <!-- 基础管理 -->
        <el-sub-menu index="basic">
          <template #title>
            <el-icon><Folder /></el-icon>
            <span>基础管理</span>
          </template>
          <el-menu-item index="/groups">
            <el-icon><FolderOpened /></el-icon>
            <template #title>分组管理</template>
          </el-menu-item>
          <el-menu-item index="/accounts">
            <el-icon><User /></el-icon>
            <template #title>账号列表</template>
          </el-menu-item>
        </el-sub-menu>

        <!-- 数据统计 -->
        <el-sub-menu index="stats">
          <template #title>
            <el-icon><DataAnalysis /></el-icon>
            <span>数据统计</span>
          </template>
          <el-menu-item index="/leads">
            <el-icon><List /></el-icon>
            <template #title>线索列表</template>
          </el-menu-item>
          <el-menu-item index="/contact-pool">
            <el-icon><Box /></el-icon>
            <template #title>底库管理</template>
          </el-menu-item>
        </el-sub-menu>

        <!-- 客户管理 -->
        <el-sub-menu index="customer">
          <template #title>
            <el-icon><Avatar /></el-icon>
            <span>客户管理</span>
          </template>
          <el-menu-item index="/customers">
            <el-icon><UserFilled /></el-icon>
            <template #title>客户列表</template>
          </el-menu-item>
          <el-menu-item index="/follow-ups">
            <el-icon><Document /></el-icon>
            <template #title>跟进记录</template>
          </el-menu-item>
        </el-sub-menu>

        <!-- 系统设置（仅管理员） -->
        <el-sub-menu v-if="isAdmin" index="system">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </template>
          <el-menu-item index="/users">
            <el-icon><UserFilled /></el-icon>
            <template #title>用户管理</template>
          </el-menu-item>
          <el-menu-item index="/llm-config">
            <el-icon><Cpu /></el-icon>
            <template #title>大模型配置</template>
          </el-menu-item>
          <el-menu-item index="/llm-logs">
            <el-icon><DataLine /></el-icon>
            <template #title>调用记录</template>
          </el-menu-item>
        </el-sub-menu>
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
            <el-breadcrumb-item v-for="(item, index) in breadcrumbList" :key="index">
              <span v-if="index === breadcrumbList.length - 1">{{ item.title }}</span>
              <router-link v-else :to="item.path">{{ item.title }}</router-link>
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <!-- WebSocket连接状态 -->
          <el-tag 
            :type="wsConnected ? 'success' : 'danger'" 
            size="small" 
            style="margin-right: 12px"
          >
            <el-icon style="margin-right: 4px">
              <component :is="wsConnected ? 'CircleCheck' : 'CircleClose'" />
            </el-icon>
            {{ wsConnected ? '已连接' : '未连接' }}
          </el-tag>
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
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
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
  FolderOpened,
  User,
  UserFilled,
  List,
  Box,
  Avatar,
  Document,
  Setting,
  Cpu,
  DataLine,
  DataAnalysis,
  Expand,
  Fold,
  ArrowDown,
  CircleCheck,
  CircleClose
} from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import { useWebSocketStore } from '@/store/websocket'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const isCollapse = ref(false)
const user = computed(() => authStore.user)
const isAdmin = computed(() => authStore.isAdmin)
const wsConnected = computed(() => wsStore.connected)

// 计算当前激活的菜单项（支持二级菜单）
const activeMenu = computed(() => {
  const path = route.path
  // 如果是二级菜单路径，返回完整路径
  return path
})

// 计算默认打开的菜单（根据当前路径自动展开对应的子菜单）
const defaultOpenedMenus = computed(() => {
  const path = route.path
  const opened = []
  
  if (path.startsWith('/groups') || path.startsWith('/accounts')) {
    opened.push('basic')
  } else if (path.startsWith('/leads') || path.startsWith('/contact-pool')) {
    opened.push('stats')
  } else if (path.startsWith('/customers') || path.startsWith('/follow-ups')) {
    opened.push('customer')
  } else if (path.startsWith('/users') || path.startsWith('/llm-')) {
    opened.push('system')
  }
  
  return opened
})

// 菜单分组映射（用于面包屑导航）
const menuGroupMap = {
  '/groups': { group: '基础管理', path: '/groups' },
  '/accounts': { group: '基础管理', path: '/accounts' },
  '/leads': { group: '数据统计', path: '/leads' },
  '/contact-pool': { group: '数据统计', path: '/contact-pool' },
  '/customers': { group: '客户管理', path: '/customers' },
  '/follow-ups': { group: '客户管理', path: '/follow-ups' },
  '/users': { group: '系统设置', path: '/users' },
  '/llm-config': { group: '系统设置', path: '/llm-config' },
  '/llm-logs': { group: '系统设置', path: '/llm-logs' }
}

// 面包屑导航
const breadcrumbList = computed(() => {
  const matched = route.matched.filter(item => item.meta && item.meta.title)
  const list = []
  const currentPath = route.path
  
  // 如果当前路径在菜单分组中，添加分组名称
  if (menuGroupMap[currentPath]) {
    list.push({
      title: menuGroupMap[currentPath].group,
      path: ''
    })
  }
  
  matched.forEach((item, index) => {
    if (index === 0) {
      // 首页
      list.push({ title: '首页', path: '/' })
    } else {
      list.push({
        title: item.meta.title,
        path: item.path
      })
    }
  })
  
  return list
})

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
      // 断开WebSocket连接
      wsStore.disconnect()
      authStore.logout()
      router.push({ name: 'Login' })
    } catch {
      // 用户取消
    }
  }
}

// 初始化WebSocket连接
import { onMounted, onUnmounted } from 'vue'
onMounted(() => {
  if (authStore.isAuthenticated) {
    wsStore.connect()
  }
})

onUnmounted(() => {
  // 不在Layout卸载时断开，因为可能只是路由切换
  // wsStore.disconnect()
})
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
      
      :deep(.el-sub-menu) {
        .el-sub-menu__title {
          color: #bfcbd9;
          
          &:hover {
            background-color: #263445;
          }
        }
        
        &.is-opened {
          .el-sub-menu__title {
            color: #409eff;
          }
        }
      }
      
      :deep(.el-sub-menu .el-menu-item) {
        background-color: #1f2d3d;
        padding-left: 50px !important;
        
        &:hover {
          background-color: #263445;
        }
        
        &.is-active {
          background-color: #409eff;
          color: #fff;
        }
      }
      
      // 折叠状态下的样式
      &.el-menu--collapse {
        :deep(.el-sub-menu__title) {
          padding-left: 20px !important;
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
  height: 60px;
  
  .header-left {
    display: flex;
    align-items: center;
    gap: 20px;
    flex: 1;
    min-width: 0; // 允许flex子元素收缩
    
    .collapse-icon {
      font-size: 20px;
      cursor: pointer;
      flex-shrink: 0;
    }
    
    :deep(.el-breadcrumb) {
      flex: 1;
      overflow: hidden;
      
      .el-breadcrumb__inner {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }
  }
  
  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
    
    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
      white-space: nowrap;
    }
  }
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
}

// 路由过渡动画
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

// 响应式布局
@media (max-width: 768px) {
  .header {
    padding: 0 10px;
    
    .header-left {
      gap: 10px;
      
      .collapse-icon {
        font-size: 18px;
      }
    }
    
    .header-right {
      :deep(.el-tag) {
        display: none; // 小屏幕隐藏WebSocket状态
      }
    }
  }
  
  .main-content {
    padding: 10px;
  }
  
  .sidebar {
    :deep(.el-menu-item) {
      padding-left: 20px !important;
    }
    
    :deep(.el-sub-menu .el-menu-item) {
      padding-left: 50px !important;
    }
  }
}
</style>

