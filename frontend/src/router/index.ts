import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

// 用户端路由
const userRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/UserLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/user/Home.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'files',
        name: 'FileList',
        component: () => import('@/views/user/FileList.vue'),
        meta: { title: '资料列表' },
      },
      {
        path: 'files/:id',
        name: 'FileDetail',
        component: () => import('@/views/user/FileDetail.vue'),
        meta: { title: '资料详情' },
      },
      {
        path: 'search',
        name: 'Search',
        component: () => import('@/views/user/Search.vue'),
        meta: { title: '搜索结果' },
      },
      {
        path: 'upload',
        name: 'Upload',
        component: () => import('@/views/user/Upload.vue'),
        meta: { title: '上传资料' },
      },
      {
        path: 'my-uploads',
        name: 'MyUploads',
        component: () => import('@/views/user/MyUploads.vue'),
        meta: { title: '我的上传' },
      },
    ],
  },
]

// 管理端路由
const adminRoutes: RouteRecordRaw[] = [
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/views/admin/Login.vue'),
    meta: { title: '管理员登录' },
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/admin/submissions',
      },
      {
        path: 'submissions',
        name: 'AdminSubmissions',
        component: () => import('@/views/admin/Submissions.vue'),
        meta: { title: '审核管理' },
      },
      {
        path: 'files',
        name: 'AdminFiles',
        component: () => import('@/views/admin/Files.vue'),
        meta: { title: '资料管理' },
      },
      {
        path: 'tags',
        name: 'AdminTags',
        component: () => import('@/views/admin/Tags.vue'),
        meta: { title: 'Tag 管理' },
      },
      {
        path: 'reports',
        name: 'AdminReports',
        component: () => import('@/views/admin/Reports.vue'),
        meta: { title: '举报管理' },
      },
      {
        path: 'announcements',
        name: 'AdminAnnouncements',
        component: () => import('@/views/admin/Announcements.vue'),
        meta: { title: '公告管理' },
      },
      {
        path: 'admins',
        name: 'AdminUsers',
        component: () => import('@/views/admin/Admins.vue'),
        meta: { title: '管理员管理' },
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('@/views/admin/Settings.vue'),
        meta: { title: '系统设置' },
      },
      {
        path: 'logs',
        name: 'AdminLogs',
        component: () => import('@/views/admin/Logs.vue'),
        meta: { title: '操作日志' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [...userRoutes, ...adminRoutes],
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  // 设置页面标题
  const title = to.meta.title as string
  document.title = title ? `${title} - OpenShare` : 'OpenShare'

  // 检查是否需要认证
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('admin_token')
    if (!token) {
      next({ name: 'AdminLogin', query: { redirect: to.fullPath } })
      return
    }
  }

  next()
})

export default router
