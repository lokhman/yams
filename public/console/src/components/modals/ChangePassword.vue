<template>
  <modal ref="modal" title="Change Password" data-backdrop="static" data-keyboard="false" @hide="$root.onModalHide">
    <form ref="form" novalidate @submit.prevent>
      <div class="form-group">
        <label>Old password: *</label>
        <input v-model="form.old" type="password" class="form-control" name="old" maxlength="72" placeholder="old password" autocomplete="off">
        <div class="invalid-feedback">Please provide a valid old password</div>
      </div>
      <div class="form-group">
        <label>New password: *</label>
        <input v-model="form.new" type="password" class="form-control" name="new" maxlength="72" placeholder="new password" autocomplete="off">
        <div class="invalid-feedback">Please provide a valid new password</div>
      </div>
    </form>

    <template slot="buttons">
      <button type="button" class="btn btn-primary" :disabled="!$root.dirty || form.old === form.new" @click="onSubmitClick">
        <i class="fas fa-save" /> Save & Close
      </button>
    </template>
  </modal>
</template>

<script>
  import Modal from '../bootstrap/Modal'

  export default {
    name: 'ChangePassword',
    components: { Modal },
    data () {
      return {
        form: {
          old: '',
          new: ''
        }
      }
    },
    watch: {
      form: {
        handler () { this.$root.dirty = true },
        deep: true
      }
    },
    methods: {
      show () {
        this.form.old = ''
        this.form.new = ''

        this.$root.resetDirty()
        this.$root.resetFormValidity(this.$refs.form)
        this.$refs.modal.show()
      },
      onSubmitClick (e) {
        const elForm = this.$refs.form
        const elSubmit = e.target

        this.$root.resetFormValidity(elForm, true)
        this.$root.lockSubmit(elSubmit)

        this.$http.post('/api/auth/password', this.form)
          .then(() => {
            this.$root.dirty = false
            this.$refs.modal.hide()
          })
          .catch(response => {
            this.$root.httpError(response, {
              403: () => {
                this.form.old = ''
                elForm.old.setCustomValidity('auth')
                elForm.old.focus()
              }
            }, elForm)
          })
          .finally(() => {
            this.$root.unlockSubmit(elSubmit)
          })
      }
    }
  }
</script>
