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
Vue.filter('bytes', function(value: any) {
  if (Math.abs(value) < 1000) {
    return value + ' B';
  }

  const units = ['K', 'M', 'G']
  let u = -1;
  const r = 10**2;

  do {
    value /= 1000;
    ++u;
  } while (Math.round(Math.abs(value) * r) / r >= 1000 && u < units.length - 1);


  return value.toFixed(2) + units[u];
})
