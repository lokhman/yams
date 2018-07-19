<template>
  <modal ref="modal" :title="title" data-backdrop="static" data-keyboard="false" @hide="$root.onModalHide">
    <form ref="form" novalidate @submit.prevent>
      <div class="form-group">
        <label>Path: *</label>
        <input v-model="form.path" type="text" class="form-control form-control-sm" name="path" maxlength="2048" placeholder="/posts/{id}" autocomplete="off" @keyup="onPathChange">
        <div class="invalid-feedback">Please provide a valid path</div>
        <small v-if="params.length > 0" class="form-text text-muted">
          Parameters:
          <template v-for="(param, index) of params">
            <code>{{ param }}</code><template v-if="index < params.length - 1">, </template>
          </template>
        </small>
      </div>
      <div class="form-group">
        <label>Methods: *</label>
        <select v-model="form.methods" class="form-control form-control-sm selectpicker" name="methods" data-actions-box="true" multiple>
          <option v-for="method of methods" :class="`text-${getMethodClassName(method)}`">{{ method }}</option>
        </select>
        <div class="invalid-feedback">Please choose valid methods</div>
      </div>
      <div class="form-group">
        <label>Timeout (s): *</label>
        <input v-model="form.timeout" type="number" class="form-control form-control-sm" name="timeout" min="1" max="86400" placeholder="60" autocomplete="off">
        <div class="invalid-feedback">Please provide a valid timeout</div>
      </div>
      <div class="form-group">
        <label>Hint:</label>
        <input v-model="form.hint" type="text" class="form-control form-control-sm" name="hint" maxlength="255" placeholder="This route is given as example" autocomplete="off">
        <div class="invalid-feedback">Please provide a valid hint</div>
      </div>
      <div class="form-check">
        <label>
          <input v-model="form.is_enabled" type="checkbox" class="form-check-input">
          enabled
        </label>
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
    name: 'RouteForm',
    components: { Modal },
    data () {
      return {
        id: 0,
        title: 'Route',
        params: [],
        form: {
          path: '',
          methods: [],
          timeout: 60,
          hint: '',
          is_enabled: true
        },
        methods: ['GET', 'HEAD', 'POST', 'PUT', 'DELETE', 'CONNECT', 'OPTIONS', 'TRACE', 'PATCH']
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

        this.form.path = ''
        this.form.methods.length = 0
        this.form.timeout = 60
        this.form.hint = ''
        this.form.is_enabled = true

        this.$root.resetDirty()
        this.$root.resetFormValidity(this.$refs.form)
        this.title = 'New Route'
        this.$refs.modal.show()
      },
      showUpdate (id) {
        this.$root.startLoading()
        this.$http.get(`/api/routes/${id}`)
          .then(response => response.json())
          .then(route => {
            this.id = id

            this.form.path = route.path
            this.form.methods.length = 0
            this.form.methods.push(...route.methods)
            this.form.timeout = route.timeout
            this.form.hint = route.hint
            this.form.is_enabled = route.is_enabled

            this.onPathChange()
            this.title = 'Edit Route'
            this.$root.resetDirty()
            this.$root.resetFormValidity(this.$refs.form)
            this.$refs.modal.show()
          })
          .catch(this.$root.httpError)
          .finally(() => {
            this.$root.stopLoading()
          })
      },
      getMethodClassName (method) {
        return this.$parent.$parent.getMethodClassName(method)
      },
      onPathChange () {
        this.params.length = 0

        const match = this.form.path.match(/{\w+}/g)
        if (match) {
          this.params.push(...match.map(v => v.slice(1, -1)))
        }
      },
      onSubmitClick (e) {
        const elForm = this.$refs.form
        const elSubmit = e.target

        this.form.timeout = +this.form.timeout
        this.form.hint = this.form.hint || null

        this.$root.resetFormValidity(elForm, true)
        this.$root.lockSubmit(elSubmit)

        const xhr = this.id
          ? this.$http.put(`/api/routes/${this.id}`, this.form)
          : this.$http.post(`/api/profiles/${this.$parent.profile.id}/routes`, this.form)

        xhr
          .then(() => {
            this.$root.dirty = false
            this.$parent.doLoadRoutes()
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
