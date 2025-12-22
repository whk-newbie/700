import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/subaccount-login',
    name: 'SubAccountLogin',
    component: () => import('@/views/auth/SubAccountLogin.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/views/admin/Layout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
        meta: { title: '首页' }
      },
      {
        path: 'groups',
        name: 'GroupManage',
        component: () => import('@/views/admin/GroupManage.vue'),
        meta: { title: '分组管理' }
      },
      {
        path: 'accounts',
        name: 'AccountList',
        component: () => import('@/views/admin/AccountList.vue'),
        meta: { title: '账号列表' }
      },
      {
        path: 'leads',
        name: 'LeadsList',
        component: () => import('@/views/admin/LeadsList.vue'),
        meta: { title: '线索列表' }
      },
      {
        path: 'contact-pool',
        name: 'ContactPool',
        component: () => import('@/views/admin/ContactPool.vue'),
        meta: { title: '底库管理' }
      },
      {
        path: 'customers',
        name: 'CustomerList',
        component: () => import('@/views/admin/CustomerList.vue'),
        meta: { title: '客户列表' }
      },
      {
        path: 'follow-ups',
        name: 'FollowUpRecords',
        component: () => import('@/views/admin/FollowUpRecords.vue'),
        meta: { title: '跟进记录' }
      },
      {
        path: 'users',
        name: 'UserManage',
        component: () => import('@/views/admin/UserManage.vue'),
        meta: { title: '用户管理', requiresAdmin: true }
      },
      {
        path: 'llm-config',
        name: 'LLMConfig',
        component: () => import('@/views/admin/LLMConfig.vue'),
        meta: { title: '大模型配置', requiresAdmin: true }
      },
      {
        path: 'llm-logs',
        name: 'LLMCallLogs',
        component: () => import('@/views/admin/LLMCallLogs.vue'),
        meta: { title: '大模型调用记录', requiresAdmin: true }
      }
    ]
  },
  {
    path: '/subaccount',
    component: () => import('@/views/subaccount/Layout.vue'),
    redirect: '/subaccount/dashboard',
    meta: { requiresAuth: true, requiresSubAccount: true },
    children: [
      {
        path: 'dashboard',
        name: 'SubAccountDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
        meta: { title: '首页' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  // 检查是否需要认证
  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
      // 未登录，重定向到登录页
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }

    // 检查管理员权限
    if (to.meta.requiresAdmin && authStore.user?.role !== 'admin') {
      next({ name: 'Dashboard' })
      return
    }

    // 检查子账号权限
    if (to.meta.requiresSubAccount && authStore.user?.role !== 'subaccount') {
      next({ name: 'Dashboard' })
      return
    }
  } else {
    // 如果已登录，访问登录页时重定向到首页
    if (to.name === 'Login' && authStore.isAuthenticated) {
      next({ name: 'Dashboard' })
      return
    }
  }

  next()
})

export default router

