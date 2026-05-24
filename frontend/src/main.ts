import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHashHistory } from 'vue-router'
import 'virtual:uno.css'
import '@unocss/reset/tailwind.css'
import './assets/globals.css'
import App from './App.vue'
import ServersView from './pages/ServersView.vue'
import ToolsView from './pages/ToolsView.vue'
import EventsView from './pages/EventsView.vue'
import ImportView from './pages/ImportView.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/servers' },
    { path: '/servers', component: ServersView },
    { path: '/servers/:id/tools', component: ToolsView },
    { path: '/servers/:id/events', component: EventsView },
    { path: '/import', component: ImportView },
  ],
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
