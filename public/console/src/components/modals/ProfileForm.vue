<template>
  <modal ref="modal" :title="title" data-backdrop="static" data-keyboard="false" @hide="$root.onModalHide">
    <form ref="form" novalidate @submit.prevent>
      <div class="form-group">
        <label>Name: *</label>
        <input v-model="form.name" type="text" class="form-control form-control-sm" name="name" maxlength="72" placeholder="New Profile" autocomplete="off">
        <div class="invalid-feedback">Please provide a valid name</div>
      </div>
      <div class="form-group">
        <label>Hosts: *</label>
        <div v-for="(host, index) of hosts" class="input-group input-group-sm mb-1">
          <div class="input-group-prepend"><span class="input-group-text">http(s)://</span></div>
          <input v-model="host.value" type="text" class="form-control" :name="`hosts[${index}]`" maxlength="128" placeholder="new-profile.yams.org:8086" autocomplete="off">
          <div class="input-group-append">
            <button type="button" class="btn btn-outline-success" @click="onAddClick(index)">
              <i class="fas fa-plus" />
            </button>
            <button v-if="hosts.length > 1" type="button" class="btn btn-outline-danger" @click="onRemoveClick(index)">
              <i class="fas fa-times" />
            </button>
          </div>
          <div class="invalid-feedback">Please provide a valid and unused host</div>
        </div>
      </div>
      <div class="form-group">
        <label>Backend:</label>
        <input v-model="form.backend" type="url" class="form-control form-control-sm" name="backend" maxlength="128" placeholder="https://example.com" autocomplete="off">
        <div class="invalid-feedback">Please provide a valid backend URL</div>
      </div>
      <div class="form-group">
        <label>Vars lifetime (s): *</label>
        <input v-model="form.vars_lifetime" type="number" class="form-control form-control-sm" name="vars_lifetime" min="1" max="2147483647" placeholder="86400" autocomplete="off">
        <div class="invalid-feedback">Please provide a valid vars lifetime</div>
      </div>
      <div class="form-check">
        <label>
          <input v-model="form.is_debug" type="checkbox" class="form-check-input">
          debug mode
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
    name: 'ProfileForm',
    components: { Modal },
    data () {
      return {
        id: 0,
        title: 'Profile',
        hosts: [],
        form: {
          name: '',
          hosts: [],
          backend: null,
          vars_lifetime: 86400,
          is_debug: true
        }
      }
    },
    watch: {
      hosts: {
        handler () { this.$root.dirty = true },
        deep: true
      },
      form: {
        handler () { this.$root.dirty = true },
        deep: true
      }
    },
    methods: {
      showCreate () {
        this.id = 0

        this.hosts.length = 0
        this.hosts.push({value: ''})

        this.form.name = ''
        this.form.backend = null
        this.form.vars_lifetime = 86400
        this.form.is_debug = true

        this.$root.resetDirty()
        this.$root.resetFormValidity(this.$refs.form)
        this.title = 'New Profile'
        this.$refs.modal.show()
      },
      showUpdate (id) {
        this.$root.startLoading()
        this.$http.get(`/api/profiles/${id}`)
          .then(response => response.json())
          .then(profile => {
            this.id = id

            this.hosts.length = 0
            this.hosts.push(...profile.hosts.map(value => ({value})))

            this.form.name = profile.name
            this.form.backend = profile.backend
            this.form.vars_lifetime = profile.vars_lifetime
            this.form.is_debug = profile.is_debug

            this.title = 'Edit Profile'
            this.$root.resetDirty()
            this.$root.resetFormValidity(this.$refs.form)
            this.$refs.modal.show()
          })
          .catch(this.$root.httpError)
          .finally(() => {
            this.$root.stopLoading()
          })
      },
      onAddClick (index) {
        this.$root.resetFormValidity(this.$refs.form)
        this.hosts.splice(index + 1, 0, { value: '' })
      },
      onRemoveClick (index) {
        this.$root.resetFormValidity(this.$refs.form)
        this.hosts.splice(index, 1)
      },
      onSubmitClick (e) {
        const elForm = this.$refs.form
        const elSubmit = e.target

        this.form.hosts.length = 0
        this.form.hosts.push(...this.hosts.map(host => host.value))
        this.form.backend = this.form.backend || null
        this.form.vars_lifetime = +this.form.vars_lifetime

        this.$root.resetFormValidity(elForm, true)
        this.$root.lockSubmit(elSubmit)

        const xhr = this.id
          ? this.$http.put(`/api/profiles/${this.id}`, this.form)
          : this.$http.post('/api/profiles', this.form)

        xhr
          .then(() => {
            this.$root.dirty = false
            this.$parent.doLoadProfiles()
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
