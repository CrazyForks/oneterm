import { UserLayout, BasicLayout, RouteView } from '@/layouts'
import appConfig from '@/config/app'
import { getAppAclRouter } from './utils'
import store from '../store'

export const generatorDynamicRouter = async () => {
  const packages = []
  const { apps = undefined } = store.getters.userInfo
  for (const appName of appConfig.buildModules) {
    if (!apps || !apps.length || apps.includes(appName)) {
      const module = await import(`@/modules/${appName}/index.js`)
      const r = await module.default.route()

      if (r.length) {
        if (module.default.name !== 'acl' && appConfig.buildAclToModules) {
          r[0].children.push(getAppAclRouter(module.default.name))
        }
        packages.push(...r)
      } else {
        if (module.default.name !== 'acl' && appConfig.buildAclToModules) {
          r.children.push(getAppAclRouter(module.default.name))
        }
        packages.push(r)
      }
    }
  }
  let routes = packages
  routes = routes.concat([
    { path: '*', redirect: '/404', hidden: true },
    {
      path: '/setting',
      component: BasicLayout,
      redirect: '/setting/companyinfo',
      meta: {},
      children: [
        {
          hidden: true,
          path: '/setting/person',
          name: 'setting_person',
          meta: { title: 'cs.menu.person', },
          component: () => import(/* webpackChunkName: "setting" */ '@/views/setting/person/index')
        },
        {
          path: '/setting/companyinfo',
          name: 'company_info',
          meta: { title: 'cs.menu.companyInfo', appName: 'backend', icon: 'ops-setting-companyInfo', selectedIcon: 'ops-setting-companyInfo', permission: ['公司信息', 'backend_admin'] },
          component: () => import(/* webpackChunkName: "setting" */ '@/views/setting/companyInfo/index')
        },
        {
          path: '/setting/companystructure',
          name: 'company_structure',
          meta: { title: 'cs.menu.companyStructure', appName: 'backend', icon: 'ops-setting-companyStructure', selectedIcon: 'ops-setting-companyStructure', permission: ['公司架构', 'backend_admin'] },
          component: () => import(/* webpackChunkName: "setting" */ '@/views/setting/companyStructure/index')
        },
        {
          path: '/setting/notice',
          name: 'notice',
          component: RouteView,
          meta: { title: 'cs.menu.notice', appName: 'backend', icon: 'ops-setting-notice', selectedIcon: 'ops-setting-notice', permission: ['通知设置', 'backend_admin'] },
          redirect: '/setting/notice/email',
          children: [{
            path: '/setting/notice/basic',
            name: 'notice_basic',
            meta: { title: 'cs.menu.basic', icon: 'ops-setting-basic', selectedIcon: 'ops-setting-basic-selected' },
            component: () => import(/* webpackChunkName: "setting" */ '@/views/setting/notice/basic')
          }, {
            path: '/setting/notice/email',
            name: 'notice_email',
            meta: { title: 'cs.menu.email', icon: 'ops-setting-notice-email', selectedIcon: 'ops-setting-notice-email-selected' },
            component: () => import(/* webpackChunkName: "setting" */ '@/views/setting/notice/email/index')
          }]
        },
        {
          path: '/setting/auth',
          name: 'company_auth',
          meta: { title: 'cs.menu.auth', appName: 'backend', icon: 'ops-setting-auth', selectedIcon: 'ops-setting-auth', permission: ['acl_admin'] },
          component: () => import(/* webpackChunkName: "setting" */ '@/views/setting/auth/index')
        },
      ]
    }, ])
  return routes
}

// basic route (module based), added according to app config configuration
const constantModuleRouteMap = []
if (appConfig.buildModules.includes('oneterm')) {
  constantModuleRouteMap.push({
    path: '/oneterm/share/:protocol/:id',
    name: 'oneterm_share',
    hidden: true,
    component: () => import('@/modules/oneterm/views/share'),
    meta: { title: 'oneterm.menu.share', keepAlive: false }
  })
}

/**
 * basic route
 */
export const constantRouterMap = [
  {
    path: '/',
    redirect: appConfig.redirectTo,
    // redirect: () => { return store.getters.appRoutes[0] },
  },
  {
    path: '/user/login',
    name: 'login',
    component: () => import(/* webpackChunkName: "user" */ '@/views/user/Login'),
  },
  {
    path: '/user/logout',
    name: 'logout',
    component: () => import(/* webpackChunkName: "user" */ '@/views/user/Logout'),
  },
  {
    path: '/user',
    component: UserLayout,
    redirect: '/user/login',
    hidden: true,
    children: [
      {
        path: 'register',
        name: 'register',
        component: () => import(/* webpackChunkName: "user" */ '@/views/user/Register'),
      },
      {
        path: 'register-result',
        name: 'registerResult',
        component: () => import(/* webpackChunkName: "user" */ '@/views/user/RegisterResult'),
      },
    ],
  },
  {
    path: '/404',
    name: '404',
    component: () => import(/* webpackChunkName: "fail" */ '@/views/exception/404'),
  },
  {
    path: '/403',
    name: '403',
    component: () => import(/* webpackChunkName: "fail" */ '@/views/exception/403'),
  },
  {
    path: '/500',
    name: '500',
    component: () => import(/* webpackChunkName: "fail" */ '@/views/exception/500'),
  },
  ...constantModuleRouteMap,
]
