import Vue from 'vue'
import moment from 'moment'

// eslint-disable-next-line
Vue.filter('datetime', function(value: any) {
  if (value) {
    return moment(String(value)).format('MM/DD/YYYY hh:mm')
  }
})

// eslint-disable-next-line
Vue.filter('timeago', function(value: any) {
  if (value) {
    return moment(String(value)).fromNow()
  }
})

// eslint-disable-next-line
Vue.filter('percent', function(value: any) {
  return (Math.floor(value * 10000) / 100) + '%'
})
