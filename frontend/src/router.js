import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import Analyze from './views/Analyze.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/analyze/:code', component: Analyze },
  { path: '/results/:code/:fightId/:player', component: Analyze },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
