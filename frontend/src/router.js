import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import Analyze from './views/Analyze.vue'
import Settings from './views/Settings.vue'
import SkippedTalents from './views/SkippedTalents.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/analyze/:code', component: Analyze },
  { path: '/results/:code/:fightId/:player', component: Analyze },
  { path: '/settings', component: Settings },
  { path: '/skipped-talents', component: SkippedTalents },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
