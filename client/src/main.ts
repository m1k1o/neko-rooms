import Vue from 'vue'
import App from './App.vue'
import './plugins/filters.ts'
import './plugins/sweetalert.ts'
import router from './router'
import store from './store'
import vuetify from './plugins/vuetify'

import '@/assets/styles/main.scss';

Vue.config.productionTip = false

new Vue({
  router,
  store,
  vuetify,
  render: h => h(App)
}).$mount('#app')
