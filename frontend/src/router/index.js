import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/store/auth'

// 懒加载组件
const Login = () => import('@/views/auth/Login.vue')
const SubAccountLogin = () => import('@/views/auth/SubAccountLogin.vue')
const Layout = () => import('@/views/admin/Layout.vue')
const Dashboard = () => import('@/views/admin/Dashboard.vue')
const GroupManage = () => import('@/views/admin/GroupManage.vue')
const AccountList = () => import('@/views/admin/AccountList.vue')
const LeadsList = () => import('@/views/admin/LeadsList.vue')
const ContactPool = () => import('@/views/admin/ContactPool.vue')
const CustomerList = () => import('@/views/admin/CustomerList.vue')
const FollowUpRecords = () => import('@/views/admin/FollowUpRecords.vue')
const UserManage = () => import('@/views/admin/UserManage.vue')
const LLMConfig = () => import('@/views/admin/LLMConfig.vue')
const LLMCallLogs = () => import('@/views/admin/LLMCallLogs.vue')

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/subaccount-login',
    name: 'SubAccountLogin',
    component: SubAccountLogin,
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: Dashboard,
        meta: { title: '首页', icon: 'HomeFilled' }
      },
      {
        path: 'groups',
        name: 'GroupManage',
        component: GroupManage,
        meta: { title: '分组管理', icon: 'FolderOpened', roles: ['admin', 'user'] }
      },
      {
        path: 'accounts',
        name: 'AccountList',
        component: AccountList,
        meta: { title: '账号列表', icon: 'Monitor', roles: ['admin', 'user'] }
      },
      {
        path: 'leads',
        name: 'LeadsList',
        component: LeadsList,
        meta: { title: '线索列表', icon: 'DataLine', roles: ['admin', 'user', 'subaccount'] }
      },
      {
        path: 'contact-pool',
        name: 'ContactPool',
        component: ContactPool,
        meta: { title: '底库管理', icon: 'Files', roles: ['admin', 'user'] }
      },
      {
        path: 'customers',
        name: 'CustomerList',
        component: CustomerList,
        meta: { title: '客户列表', icon: 'User', roles: ['admin', 'user', 'subaccount'] }
      },
      {
        path: 'follow-ups',
        name: 'FollowUpRecords',
        component: FollowUpRecords,
        meta: { title: '跟进记录', icon: 'DocumentCopy', roles: ['admin', 'user', 'subaccount'] }
      },
      {
        path: 'users',
        name: 'UserManage',
        component: UserManage,
        meta: { title: '用户管理', icon: 'UserFilled', roles: ['admin'] }
      },
      {
        path: 'llm-config',
        name: 'LLMConfig',
        component: LLMConfig,
        meta: { title: '大模型配置', icon: 'Setting', roles: ['admin'] }
      },
      {
        path: 'llm-logs',
        name: 'LLMCallLogs',
        component: LLMCallLogs,
        meta: { title: '调用记录', icon: 'Document', roles: ['admin'] }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // 检查是否需要登录
  if (to.meta.requiresAuth !== false && !authStore.isAuthenticated) {
    next('/login')
    return
  }

  // 检查角色权限
  if (to.meta.roles && !to.meta.roles.includes(authStore.user?.role)) {
    next('/dashboard')
    return
  }

  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - Line管理系统`
  }

  next()
})

export default router
