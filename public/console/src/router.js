import Vue from 'vue'
import VueRouter from 'vue-router'

import Index from './components/Index'
import Login from './components/Login'
import Profiles from './components/Profiles'
import Routes from './components/Routes'
import Assets from './components/Assets'
import Users from './components/Users'

import Error403 from './components/errors/403'
import Error404 from './components/errors/404'

Vue.use(VueRouter)

export default new VueRouter({
  mode: 'history',
  base: '/web/',
  routes: [
    {path: '/login', name: 'login', component: Login},
    {path: '/', component: Index, children: [
        {path: '/', name: 'index', redirect: {name: 'profiles'}},
        {path: '/profiles', name: 'profiles', component: Profiles},
        {path: '/profiles/:id/routes', name: 'routes', component: Routes},
        {path: '/profiles/:id/assets', name: 'assets', component: Assets},
        {path: '/users', name: 'users', component: Users}
      ]},
    {path: '/403', component: Error403},
    {path: '*', component: Error404}
  ],
  scrollBehavior () {
    return { x: 0, y: 0 }
  }
})
