import Vue from 'vue'
import Router from 'vue-router'
import Home from '@/components/Home'
import Item from '@/components/Item'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home
    },
    {
      path: '/items/:id',
      name: 'Item',
      component: Item
    }
  ]
})
