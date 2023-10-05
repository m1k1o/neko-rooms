import Vue from 'vue'
import App from './App.vue'
import './plugins/filters.ts'
import sweetalert from './plugins/sweetalert'
import store from './store'
import vuetify from './plugins/vuetify'

import '@/assets/styles/main.scss';

Vue.config.productionTip = false

Vue.use(sweetalert)

new Vue({
  store,
  vuetify,
  render: h => h(App)
}).$mount('#app')
