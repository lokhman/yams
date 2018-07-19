<template>
  <modal ref="modal" title="New Asset" data-backdrop="static" data-keyboard="false" @hide="onModalHide">
    <form novalidate @submit.prevent>
      <div v-if="isAssetLarge()" class="alert alert-danger" role="alert">Asset is too large &mdash; {{ formatNumber(file.size) }} bytes!</div>
      <div class="form-group">
        <label>Path:</label>
        <input v-model="path" type="text" class="form-control" maxlength="72" :placeholder="initPath" :disabled="progress !== -1" autocomplete="off">
      </div>
      <div class="form-group">
        <label>Asset (max 64MB): *</label>
        <div class="custom-file">
          <input type="file" class="custom-file-input" id="asset-form-file" :disabled="progress !== -1" autocomplete="off" @change="onFileChange">
          <label class="custom-file-label" for="asset-form-file">{{ file.name || 'Choose file...' }}</label>
        </div>
      </div>
      <div v-if="progress !== -1" class="progress">
        <div class="progress-bar" role="progressbar" :style="{width: `${progress}%`}">{{ progress }}%</div>
      </div>
    </form>

    <template slot="buttons">
      <button type="button" class="btn btn-primary" :disabled="!file.name || isAssetLarge()" @click="onSubmitClick">
        <i class="fas fa-upload" /> Upload & Close
      </button>
    </template>
  </modal>
</template>

<script>
  import Modal from '../bootstrap/Modal'

  export default {
    name: 'AssetForm',
    components: { Modal },
    data () {
      return {
        path: '',
        file: {},
        initPath: '',
        progress: -1,
        request: null
      }
    },
    methods: {
      show () {
        this.path = ''
        this.file = {}

        this.initPath = `${this.getRandomString()}.dat`
        this.progress = -1
        this.request = null

        this.$root.resetDirty()
        this.$refs.modal.show()
      },
      isAssetLarge () {
        return this.file.size > 64 << 20
      },
      onFileChange (e) {
        const files = e.target.files
        if (files.length === 0) {
          return
        }
        this.file = files[0]
        this.$root.dirty = true
      },
      onSubmitClick (e) {
        if (!(this.file instanceof File)) {
          return
        }

        const vm = this
        const elSubmit = e.target
        const path = this.path || this.initPath
        const config = {
          headers: {'Content-Type': this.file.type.split(';')[0] || 'application/octet-stream'},
          uploadProgress (e) { vm.progress = +(e.loaded * 100 / e.total).toFixed(1) },
          before (request) { vm.request = request }
        }

        let exists = false
        this.$parent.assets.forEach(asset => {
          if (asset.path === path) {
            exists = true
          }
        })

        if (exists && !confirm(`Asset with path "${path}" already exists in the current profile. Do you want to overwrite it?`)) {
          return  // protection against accidental asset overwriting
        }

        this.$root.lockSubmit(elSubmit)

        this.$http.put(`/api/profiles/${this.$parent.profile.id}/assets/${path}`, this.file, config)
          .then(() => {
            this.$root.dirty = false
            this.$parent.doLoadAssets()
            this.$refs.modal.hide()
          })
          .catch(this.$root.httpError)
          .finally(() => {
            this.progress = -1
            this.request = null
            this.$root.unlockSubmit(elSubmit)
          })
      },
      onModalHide (e) {
        if (this.$root.onModalHide(e) && this.request) {
          this.request.abort()
        }
      }
    }
  }
</script>
