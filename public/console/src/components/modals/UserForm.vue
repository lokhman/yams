<template>
  <modal ref="modal" :title="title" data-backdrop="static" data-keyboard="false" @hide="$root.onModalHide">
    <form ref="form" novalidate @submit.prevent>
      <div class="form-group">
        <label>Username: *</label>
        <input v-model="form.username" type="text" class="form-control form-control-sm" name="username" maxlength="32" placeholder="new.user123" autocomplete="off">
        <div class="invalid-feedback">Please provide a valid username</div>
      </div>
      <div class="form-group">
        <label>Password:</label>
        <div class="input-group input-group-sm">
          <input v-model="form.password" :type="showPassword ? 'text' : 'password'" class="form-control" name="password" maxlength="72" placeholder="new password" autocomplete="off">
          <div class="input-group-append">
            <button type="button" class="btn btn-outline-info" @click="showPassword = !showPassword">
              <span v-show="showPassword"><i class="fas fa-eye-slash" /></span>
              <span v-show="!showPassword"><i class="fas fa-eye" /></span>
            </button>
          </div>
          <div class="invalid-feedback">Please provide a valid password</div>
        </div>
      </div>
      <div class="form-group">
        <label>Role: *</label>
        <select v-model="form.role" class="form-control form-control-sm" name="role">
          <option v-for="role of roles">{{ role }}</option>
        </select>
        <div class="invalid-feedback">Please choose a valid role</div>
      </div>
      <div class="form-group">
        <label>ACL:</label>
        <select v-model="form.acl" class="form-control form-control-sm selectpicker" name="acl" data-actions-box="true" multiple>
          <option v-for="profile of $parent.profiles" :value="profile.id">{{ profile.name }}</option>
        </select>
        <div class="invalid-feedback">Please choose valid profiles</div>
      </div>
    </form>

    <template slot="buttons">
      <button type="button" class="btn btn-primary" :disabled="!$root.dirty" @click="onSubmitClick">
        <i class="fas fa-save" /> Save & Close
      </button>
    </template>
  </modal>
</template>

<script>
  import Modal from '../bootstrap/Modal'

  export default {
    name: 'UserForm',
    components: { Modal },
    data () {
      return {
        id: 0,
        title: 'User',
        showPassword: false,
        form: {
          username: '',
          password: '',
          role: '',
          acl: []
        },
        roles: ['admin', 'manager', 'developer']
      }
    },
    watch: {
      form: {
        handler () { this.$root.dirty = true },
        deep: true
      }
    },
    methods: {
      showCreate () {
        this.id = 0
        this.showPassword = true

        this.form.username = ''
        this.form.password = this.getRandomString().slice(-8)
        this.form.role = ''
        this.form.acl.length = 0

        this.$root.resetDirty()
        this.$root.resetFormValidity(this.$refs.form)
        this.title = 'New User'
        this.$refs.modal.show()
      },
      showUpdate (id) {
        this.$root.startLoading()
        this.$http.get(`/api/users/${id}`)
          .then(response => response.json())
          .then(user => {
            this.id = id
            this.showPassword = false

            this.form.username = user.username
            this.form.password = ''
            this.form.role = user.role

            this.form.acl.length = 0
            this.form.acl.push(...user.acl || [])

            this.title = 'Edit User'
            this.$root.resetDirty()
            this.$root.resetFormValidity(this.$refs.form)
            this.$refs.modal.show()
          })
          .catch(this.$root.httpError)
          .finally(() => {
            this.$root.stopLoading()
          })
      },
      onSubmitClick (e) {
        const elForm = this.$refs.form
        const elSubmit = e.target

        this.$root.resetFormValidity(elForm, true)
        this.$root.lockSubmit(elSubmit)

        const xhr = this.id
          ? this.$http.put(`/api/users/${this.id}`, this.form)
          : this.$http.post('/api/users', this.form)

        xhr
          .then(() => {
            this.$root.dirty = false
            this.$parent.doLoadUsers()
            this.$refs.modal.hide()
          })
          .catch(response => {
            this.$root.httpError(response, {}, elForm)
          })
          .finally(() => {
            this.$root.unlockSubmit(elSubmit)
          })
      }
    }
  }
</script>
