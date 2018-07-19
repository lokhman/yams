import Vue from 'vue'
import Yams from './Yams'
import router from './router'
import VueResource from 'vue-resource'
import VueMeta from 'vue-meta'
import $ from 'jquery'

Vue.use(VueResource)
Vue.use(VueMeta)

Vue.mixin({
  methods: {
    deepCopy (object) {
      return $.extend(true, {}, object)
    },
    formatDate (date) {
      return new Date(date).toLocaleString()
    },
    formatNumber (number) {
      return new Intl.NumberFormat().format(number)
    },
    formatFileSize (bytes, delimiter = ' ') {
      const index = Math.log(bytes) / Math.log(1024) | 0
      const number = +(bytes / Math.pow(1024, index)).toFixed(2)
      return this.formatNumber(number) + delimiter + ['B', 'KB', 'MB', 'GB', 'TB'][index]
    },
    getRandomString () {
      const x = 2147483648
      const left = Math.floor(Math.random() * x).toString(36)
      const right = Math.abs(Math.floor(Math.random() * x) ^ +new Date()).toString(36)
      return left + right
    },
    getStringSize (str) {
      let bytes = 0
      for (let i = 0; i < str.length; i++) {
        const codePoint = str.charCodeAt(i)

        if (codePoint >= 0xD800 && codePoint < 0xE000) {
          if (codePoint < 0xDC00 && i + 1 < str.length) {
            const next = str.charCodeAt(i + 1)

            if (next >= 0xDC00 && next < 0xE000) {
              bytes += 4
              i++
              continue
            }
          }
        }
        bytes += codePoint < 0x80 ? 1 : (codePoint < 0x800 ? 2 : 3)
      }
      return bytes
    }
  }
})

Vue.directive('tooltip', (el, binding) => {
  el.title = binding.value
  $(el).tooltip().removeAttr('title')
})

router.afterEach(() => {
  $('.tooltip').remove()
})

new (Vue.extend(Yams))({
  el: '#yams',
  router
})
