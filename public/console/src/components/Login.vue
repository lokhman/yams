<template>
  <main role="main" class="bg-light">
    <form method="post" novalidate @submit.prevent="onSubmit">
      <div class="position-relative">
        <input v-model="form.username" type="text" class="form-control" name="username" placeholder="Username" autofocus autocomplete="off">
        <div class="invalid-tooltip">Please provide a valid username</div>
      </div>
      <div class="position-relative">
        <input v-model="form.password" type="password" class="form-control" name="password" placeholder="Password">
        <div class="invalid-tooltip">Please provide a valid password</div>
      </div>
      <button ref="submit" class="btn btn-primary btn-block mt-3" type="submit">Sign in</button>
    </form>
  </main>
</template>

<script>
  import Cookie from 'js-cookie'
  import router from '../router'

  export default {
    name: 'Login',
    metaInfo: {
      title: 'Login'
    },
    data () {
      return {
        form: {
          username: '',
          password: ''
        }
      }
    },
    methods: {
      onSubmit (e) {
        const elForm = e.target
        const elSubmit = this.$refs.submit

        this.$root.resetFormValidity(elForm, true)
        this.$root.lockSubmit(elSubmit)
        this.$http.post('/api/auth', this.form)
          .then(response => response.json())
          .then(auth => {
            Cookie.set('token', auth.token)
            this.$root.token = auth.token
            this.$root.doGetAuth(() => {
              router.push({name: 'index'})
            })
          })
          .catch(response => {
            this.$root.httpError(response, {
              401: () => {
                this.form.password = ''
                e.target.password.setCustomValidity('auth')
                e.target.password.focus()
              }
            }, elForm)
          })
          .finally(() => {
            this.$root.unlockSubmit(elSubmit)
          })
      }
    },
    beforeRouteEnter (to, from, next) {
      next(vm => {
        if (vm.$root.token) {
          next({name: 'index'})
        }
      })
    },
    created () {
      this.$root.stopLoading()
    }
  }
</script>

<style lang="scss" scoped>
  [role='main'] {
    height: 100%;
    display: flex;
    align-items: center;
  }

  form {
    width: 100%;
    max-width: 330px;
    margin: auto;
    padding: 15px;

    [name='username'] {
      border-bottom-left-radius: 0;
      border-bottom-right-radius: 0;
      margin-bottom: -1px;
      z-index: 1;
    }

    [name='password'] {
      border-top-left-radius: 0;
      border-top-right-radius: 0;
    }

    .form-control:focus {
      position: relative;
    }
  }
</style>
