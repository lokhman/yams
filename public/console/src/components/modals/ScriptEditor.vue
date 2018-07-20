<template>
  <modal ref="modal" :title="`Script Editor (${mode})`" size="large" data-backdrop="static" data-keyboard="false" @hide="$root.onModalHide">
    <div v-if="isScriptLarge()" class="alert alert-danger" role="alert">Script is too long &mdash; {{ formatNumber(size) }} bytes!</div>
    <editor v-model="script" :mode="mode" id="yams-editor" @save="onSaveClick()"></editor>
    <div class="d-flex justify-content-between">
      <code>{{ route.path || '...' }}</code>
      <small>{{ contentType }}</small>
    </div>

    <template slot="buttons">
      <button type="button" class="btn btn-primary" :disabled="!$root.dirty || isScriptLarge()" @click="onSaveClick">
        <i class="far fa-save" /> Save
      </button>
      <button type="button" class="btn btn-primary" :disabled="!$root.dirty || isScriptLarge()" @click="onSaveCloseClick">
        <i class="fas fa-save" /> Save & Close
      </button>
    </template>
  </modal>
</template>

<script>
  import Modal from '../bootstrap/Modal'
  import Editor from '../bootstrap/Editor'

  const SCRIPT_MODES = {
    'application/x-lua': 'lua'
  }

  export default {
    name: 'ScriptEditor',
    components: { Modal, Editor },
    data () {
      return {
        route: {},
        mode: 'text',
        contentType: '',
        script: '',
        size: 0
      }
    },
    watch: {
      script (val) {
        this.$root.dirty = true
        this.size = this.getStringSize(val)
      }
    },
    methods: {
      show (route) {
        this.route = route

        this.$root.startLoading()
        this.$http.get(`/api/routes/${route.id}/script`)
          .then(response => {
            this.contentType = response.headers.get('Content-Type')
            this.mode = SCRIPT_MODES[this.contentType] || 'text'
            this.script = response.data
            this.$root.resetDirty()
            this.$refs.modal.show()
          })
          .catch(this.$root.httpError)
          .finally(() => {
            this.$root.stopLoading()
          })
      },
      doSave () {
        const config = {headers: {'content-type': this.contentType}}

        return this.$http.put(`/api/routes/${this.route.id}/script`, this.script, config)
          .then(() => {
            this.$root.dirty = false
            this.route.script_size = this.size
          })
          .catch(this.$root.httpError)
      },
      isScriptLarge () {
        return this.size > 8 << 20
      },
      onSaveClick () {
        this.doSave()
      },
      onSaveCloseClick () {
        this.doSave().then(() => {
          this.$parent.doLoadRoutes()
          this.$refs.modal.hide()
        })
      }
    }
  }
</script>

<style lang="scss" scoped>
  #yams-editor {
    position: relative;
    height: 480px;
    border: 1px solid #999;
  }
</style>
