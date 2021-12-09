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

// eslint-disable-next-line
Vue.filter('memory', function(value: any) {
  if (value < 1e3) {
    return value + 'B'
  }

  if (value < 1e6) {
    return (value / 1e3).toFixed(0) + 'K'
  }

  if (value < 1e9) {
    return (value / 1e6).toFixed(0) + 'M'
  }

  return (value / 1e9).toFixed(1) + 'G'
})

// eslint-disable-next-line
Vue.filter('nanocpus', function(value: any) {
  return (value / 1e9).toFixed(1) + 'x'
})
