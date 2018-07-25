<template>
  <div id="yams">
    <div class="loader" :class="{loading}">
      <i class="fas fa-sync fa-spin" />
    </div>
    <header>
      <nav class="navbar navbar-expand-lg navbar-dark fixed-top bg-dark flex-md-nowrap shadow px-4">
        <router-link :to="{name: 'index'}" class="navbar-brand">YAMS Console</router-link>
        <button v-if="token" type="button" class="navbar-toggler" data-toggle="collapse" data-target="#navbar">
          <span class="navbar-toggler-icon" />
        </button>
        <div v-if="token" class="collapse navbar-collapse" id="navbar">
          <ul class="navbar-nav mr-auto">
            <li class="nav-item" :class="{active: $route.name === 'profiles'}">
              <router-link :to="{name: 'profiles'}" class="nav-link">
                <i class="fas fa-fw fa-cogs" /> Profiles
              </router-link>
            </li>
            <li v-if="nav.profile" class="nav-item" :class="{active: $route.name === 'routes'}">
              <router-link :to="{name: 'routes', params: {id: nav.profile.id}}" class="nav-link">
                <i class="fas fa-fw fa-exchange-alt" /> Routes
              </router-link>
            </li>
            <li v-if="nav.profile" class="nav-item" :class="{active: $route.name === 'assets'}">
              <router-link :to="{name: 'assets', params: {id: $root.nav.profile.id}}" class="nav-link">
                <i class="fas fa-fw fa-copy" /> Assets
              </router-link>
            </li>
            <li v-if="auth.role === 'admin'" class="nav-item" :class="{active: $route.name === 'users'}">
              <router-link :to="{name: 'users'}" class="nav-link">
                <i class="fas fa-fw fa-users" /> Users
              </router-link>
            </li>
            <li class="nav-item">
              <a class="nav-link" @click="onChangePasswordClick">
                <i class="fas fa-fw fa-key" /> Password
              </a>
            </li>
          </ul>
          <ul class="navbar-nav">
            <li class="nav-item">
              <a class="nav-link" @click="doLogout()">
                <i class="fas fa-fw fa-sign-out-alt" /> Sign out
              </a>
            </li>
          </ul>
        </div>
      </nav>
    </header>
    <transition name="page" mode="out-in">
      <router-view />
    </transition>
    <component :is="modal" ref="modal" />
  </div>
</template>

<script>
  import $ from 'jquery'
  import Cookie from 'js-cookie'
  import router from './router'
  import RequestError from './components/modals/RequestError'
  import ChangePassword from './components/modals/ChangePassword'

  export default {
    name: 'Yams',
    metaInfo: {
      title: 'Index',
      titleTemplate: '%s — YAMS Console'
    },
    data () {
      return {
        auth: {},
        token: Cookie.get('token'),
        modal: null,
        dirty: false,
        loading: true,
        nav: {
          profile: null
        }
      }
    },
    methods: {
      startLoading () {
        this.loading = true
      },
      stopLoading () {
        this.loading = false
      },
      resetNavigation () {
        this.nav.profile = null
      },
      httpError (response, errorMap = {}, form) {
        if (response.ok) {
          return
        }

        const status = response.status
        const data = response.data

        errorMap = Object.assign({
          401: () => this.doLogout(),
          403: () => router.history.updateRoute(router.match('403')),
          404: () => router.history.updateRoute(router.match('404')),
          422: () => {
            if (form && data.invalid) {
              for (let name in data.invalid) {
                form[name].setCustomValidity(data.invalid[name])
              }
              for (let input of Array.from(form.elements)) {
                if (!input.validity.valid) {
                  input.focus()
                  break
                }
              }
            }
          }
        }, errorMap)

        // call mapped function
        if (status in errorMap) {
          errorMap[status](data, response.headers)
        } else {
          const args = [response.url]
          if (status && response.statusText) {
            args.push(`${status} ${response.statusText}`)
          }
          setTimeout(() => {
            // may catch previous modal in active state
            this.showModal(RequestError, this, 'show', ...args)
          })
        }
      },
      resetFormValidity (form, addWasValidatedClass) {
        for (let input of Array.from(form.elements)) {
          if (input.name in form.elements) {
            input.setCustomValidity('')
          }
        }
        if (addWasValidatedClass) {
          form.classList.add('was-validated')
        } else {
          form.classList.remove('was-validated')
        }
        setTimeout(() => {  // updates plugin layout
          const $selectpicker = $('.selectpicker', form)
          const $parent = $selectpicker.parent()
          $parent.next('.invalid-feedback').appendTo($parent)
          $selectpicker.selectpicker('refresh')
        })
      },
      lockSubmit (button) {
        const $spinner = $('<i class="fas fa-cog fa-spin ml-1" />')
        $spinner.appendTo(button)
        button.disabled = true
      },
      unlockSubmit(button) {
        $('.fa-spin', button).remove()
        button.disabled = false
      },
      resetDirty () {
        this.$nextTick(() => {
          this.dirty = false
        })
      },
      showModal (modal, context, method = 'show', ...args) {
        this.modal = modal
        this.$nextTick(() => {
          this.$refs.modal.$parent = context
          this.$refs.modal[method](...args)
        })
      },
      onModalHide (e) {
        if (this.dirty && !window.confirm('Changes that you made may not be saved. Do you want to proceed?')) {
          e.preventDefault()  // this will prevent `hide.bs.modal` event to trigger
          return false
        }
        this.dirty = false
        return true
      },
      onChangePasswordClick (e) {
        this.showModal(ChangePassword, this)
      },
      doGetAuth (callback) {
        if (!this.token) {
          return
        }
        this.$http.get('/api/auth')
          .then(response => response.json())
          .then(auth => {
            this.auth = auth
            if (typeof callback === 'function') {
              callback(auth)
            }
          })
          .catch(this.httpError)
      },
      doLogout () {
        if (!this.token) {
          return
        }
        this.auth = {}
        this.token = ''
        Cookie.remove('token')
        router.push({name: 'login'})
      },
      refreshToken (callback) {
        if (!this.token) {
          return
        }
        this.$http.post('/api/auth/refresh')
          .then(response => response.json())
          .then(auth => {
            Cookie.set('token', auth.token)
            this.token = auth.token
            if (typeof callback === 'function') {
              callback(auth.token)
            }
          })
          .catch(this.httpError)
      }
    },
    created () {
      // refresh token and get auth info
      this.refreshToken(() => this.doGetAuth())

      // event to protect dirty forms
      addEventListener('beforeunload', (e) => {
        if (this.dirty) {
          e.returnValue = this.dirty
          return this.dirty
        }
      })

      // task to refresh token every 45 minutes
      setInterval(() => this.refreshToken(), 45 * 60 * 1000)
    }
  }
</script>

<style lang="scss">
  html,
  body,
  #yams {
    height: 100%;
  }

  main {
    padding-top: 56px;
    font-size: .875rem;
  }

  a {
    cursor: pointer;
  }

  .tooltip {
    font-size: .7rem;
  }

  .loader {
    position: fixed;
    display: none;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    align-items: center;
    justify-content: center;
    font-size: 10rem;
    background: #000;
    z-index: 999999999;
    opacity: .25;

    &.loading {
      display: flex;
    }
  }

  $color-notice: #ae9602;
  $color-surprise: #7d69cb;

  %bg-notice {
    background-color: $color-notice;
  }

  %bg-surprise {
    background-color: $color-surprise;
  }

  .bg-notice,
  .badge-notice {
    @extend %bg-notice;
  }

  .bg-surprise,
  .badge-surprise {
    @extend %bg-surprise;
  }

  .badge-notice,
  .badge-surprise,
  .badge-notice[href]:hover,
  .badge-surprise[href]:hover {
    color: #fff;
    text-decoration: none;
  }

  .text-notice {
    color: $color-notice;
  }

  .text-surprise {
    color: $color-surprise;
  }

  .page-enter-active,
  .page-leave-active {
    transition: opacity .25s;
  }
  .page-enter,
  .page-leave-to {
    opacity: 0;
  }

  .btn-group-xs > .btn,
  .btn-xs {
    padding: .25rem .4rem;
    font-size: .875rem;
    line-height: .5;
    border-radius: .2rem;

    .badge {
      line-height: inherit;
    }
  }

  .text-line-through {
    text-decoration: line-through;
  }
</style>
