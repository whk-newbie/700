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
          <el-menu-item v-if="isAdmin" index="/groups">
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
const isUser = computed(() => authStore.user?.role === 'user')
const isSubAccount = computed(() => authStore.isSubAccount)
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
  overflow: hidden;
}

.sidebar {
  background: #fff;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.06);
  border-right: 1px solid #e4e7ed;
  
  .logo {
    height: 64px;
    line-height: 64px;
    text-align: center;
    color: #303133;
    font-size: 20px;
    font-weight: 700;
    background: #fff;
    border-bottom: 1px solid #e4e7ed;
    position: relative;
    z-index: 1;
    transition: all 0.3s;
    
    span {
      display: inline-block;
      letter-spacing: 1px;
    }
  }
  
  .sidebar-menu {
    border: none;
    background: #fff;
    padding: 8px 0;
    
    :deep(.el-menu-item) {
      color: #606266;
      margin: 2px 12px;
      border-radius: 6px;
      height: 44px;
      line-height: 44px;
      transition: all 0.2s;
      position: relative;
      
      .el-icon {
        font-size: 18px;
        margin-right: 12px;
        color: #606266;
        transition: color 0.2s;
      }
      
      &:hover {
        background: #f5f7fa;
        color: #303133;
        
        .el-icon {
          color: #303133;
        }
      }
      
      &.is-active {
        background: #ecf5ff;
        color: #409eff;
        font-weight: 600;
        
        .el-icon {
          color: #409eff;
        }
      }
    }
    
    :deep(.el-sub-menu) {
      margin: 2px 12px;
      
      .el-sub-menu__title {
        color: #606266;
        border-radius: 6px;
        height: 44px;
        line-height: 44px;
        transition: all 0.2s;
        
        .el-icon {
          font-size: 18px;
          margin-right: 12px;
          color: #606266;
          transition: color 0.2s;
        }
        
        &:hover {
          background: #f5f7fa;
          color: #303133;
          
          .el-icon {
            color: #303133;
          }
        }
      }
      
      &.is-opened {
        .el-sub-menu__title {
          color: #409eff;
          background: #f5f7fa;
          font-weight: 600;
          
          .el-icon {
            color: #409eff;
          }
        }
      }
    }
    
    :deep(.el-sub-menu .el-menu-item) {
      background: #fff;
      padding-left: 20px !important;
      margin: 2px 12px;
      border-radius: 6px;
      height: 40px;
      line-height: 40px;
      transition: all 0.2s;
      
      .el-icon {
        font-size: 16px;
        margin-right: 10px;
        color: #909399;
      }
      
      &:hover {
        background: #f5f7fa;
        color: #303133;
        
        .el-icon {
          color: #606266;
        }
      }
      
      &.is-active {
        background: #ecf5ff;
        color: #409eff;
        font-weight: 600;
        
        .el-icon {
          color: #409eff;
        }
      }
    }
    
    // 折叠状态下的样式
    &.el-menu--collapse {
      :deep(.el-sub-menu__title) {
        padding-left: 20px !important;
        justify-content: center;
      }
      
      :deep(.el-menu-item) {
        padding-left: 20px !important;
        justify-content: center;
      }
    }
  }
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 24px;
  height: 64px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  position: relative;
  z-index: 100;
  
  .header-left {
    display: flex;
    align-items: center;
    gap: 24px;
    flex: 1;
    min-width: 0;
    
    .collapse-icon {
      font-size: 22px;
      cursor: pointer;
      flex-shrink: 0;
      color: #606266;
      transition: all 0.3s;
      padding: 8px;
      border-radius: 6px;
      
      &:hover {
        background: #f5f7fa;
        color: #409eff;
        transform: scale(1.1);
      }
    }
    
    :deep(.el-breadcrumb) {
      flex: 1;
      overflow: hidden;
      
      .el-breadcrumb__item {
        .el-breadcrumb__inner {
          color: #909399;
          font-weight: 500;
          transition: color 0.3s;
          
          &.is-link {
            color: #606266;
            
            &:hover {
              color: #409eff;
            }
          }
        }
        
        &:last-child {
          .el-breadcrumb__inner {
            color: #303133;
            font-weight: 600;
          }
        }
      }
      
      .el-breadcrumb__separator {
        color: #c0c4cc;
        margin: 0 8px;
      }
    }
  }
  
  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-shrink: 0;
    
    :deep(.el-tag) {
      border-radius: 16px;
      padding: 6px 12px;
      font-weight: 500;
      border: none;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
      transition: all 0.3s;
      
      &:hover {
        transform: translateY(-1px);
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
      }
    }
    
    .user-info {
      display: flex;
      align-items: center;
      gap: 10px;
      cursor: pointer;
      white-space: nowrap;
      padding: 8px 12px;
      border-radius: 8px;
      transition: all 0.3s;
      color: #606266;
      font-weight: 500;
      
      .el-icon {
        font-size: 18px;
        transition: transform 0.3s;
      }
      
      &:hover {
        background: #f5f7fa;
        color: #409eff;
        
        .el-icon:last-child {
          transform: rotate(180deg);
        }
      }
    }
    
    :deep(.el-dropdown-menu) {
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      border: 1px solid #e4e7ed;
      padding: 8px 0;
      
      .el-dropdown-menu__item {
        padding: 12px 20px;
        transition: all 0.3s;
        
        &:hover {
          background: #f5f7fa;
          color: #409eff;
        }
      }
    }
  }
}

.main-content {
  background: linear-gradient(180deg, #f5f7fa 0%, #f0f2f5 100%);
  padding: 24px;
  overflow-y: auto;
  min-height: calc(100vh - 64px);
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
    padding: 0 16px;
    height: 56px;
    
    .header-left {
      gap: 12px;
      
      .collapse-icon {
        font-size: 20px;
        padding: 6px;
      }
      
      :deep(.el-breadcrumb) {
        font-size: 13px;
        
        .el-breadcrumb__separator {
          margin: 0 4px;
        }
      }
    }
    
    .header-right {
      gap: 8px;
      
      :deep(.el-tag) {
        display: none; // 小屏幕隐藏WebSocket状态
      }
      
      .user-info {
        padding: 6px 8px;
        font-size: 14px;
        
        span {
          display: none; // 小屏幕隐藏用户名
        }
      }
    }
  }
  
  .main-content {
    padding: 16px;
  }
  
  .sidebar {
    .logo {
      height: 56px;
      line-height: 56px;
      font-size: 16px;
    }
    
    .sidebar-menu {
      :deep(.el-menu-item) {
        height: 44px;
        line-height: 44px;
        margin: 2px 4px;
        font-size: 14px;
      }
      
      :deep(.el-sub-menu) {
        margin: 2px 4px;
        
        .el-sub-menu__title {
          height: 44px;
          line-height: 44px;
          font-size: 14px;
        }
      }
      
      :deep(.el-sub-menu .el-menu-item) {
        height: 40px;
        line-height: 40px;
        font-size: 13px;
      }
    }
  }
}
</style>

