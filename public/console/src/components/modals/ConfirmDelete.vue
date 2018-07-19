<template>
  <modal ref="modal" title="Are you sure?">
    <div v-if="confirmText">
      <p>This will <em>permanently</em> delete <span v-html="what" /> from the system. <span class="text-danger">This action <b>cannot</b> be undone!</span></p>
      <p>Please type in <strong>{{ confirmText }}</strong> to confirm:</p>
      <input v-model="confirmInput" type="text" class="form-control form-control-sm" title="">
    </div>
    <div v-else>Are you sure that you want to delete <span v-html="what" />?</div>

    <template slot="buttons">
      <button type="button" class="btn btn-danger" :disabled="confirmText !== confirmInput" @click="onDeleteClick">
        <i class="fas fa-trash-alt" /> Delete
      </button>
    </template>
  </modal>
</template>

<script>
  import Modal from '../bootstrap/Modal'

  export default {
    name: 'ConfirmDelete',
    components: { Modal },
    data () {
      return {
        what: '',
        confirmText: '',
        confirmInput: ''
      }
    },
    methods: {
      show (what, callback, confirmText = '') {
        this.what = what
        this.callback = callback
        this.confirmText = confirmText
        this.$refs.modal.show()
      },
      onDeleteClick (e) {
        this.callback(e, this)
      }
    }
  }
</script>
