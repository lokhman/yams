<template>
  <modal ref="modal" title="Request Error">
    Application cannot complete server request to <code>{{ url }}</code> due to the following error:
    <code>{{ error }}</code>.
  </modal>
</template>

<script>
  import Modal from '../bootstrap/Modal'

  export default {
    name: 'RequestError',
    components: { Modal },
    data () {
      return {
        url: '',
        error: ''
      }
    },
    methods: {
      show (url = '[invalid URL]', error = 'Unknown error') {
        this.url = url
        this.error = error
        this.$refs.modal.show()
      }
    }
  }
</script>
