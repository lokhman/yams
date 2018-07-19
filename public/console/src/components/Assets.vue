<template>
  <div>
    <h1>Assets <small class="text-muted">{{ profile.name || '...' }}</small></h1>
    <p class="lead">
      This is the list of available assets in profile "{{ profile.name || '...' }}".
      If you want to upload a new asset <a tabindex="0" @click="onAddClick">click here</a>.
    </p>
    <div class="table-responsive">
      <table class="table table-sm table-hover">
        <caption class="font-weight-bold">Total: {{ assets.length }}</caption>
        <thead>
        <tr>
          <th>Path</th>
          <th style="width: 350px">MIME Type</th>
          <th style="width: 150px">Size</th>
          <th colspan="2">Created At</th>
        </tr>
        </thead>
        <tbody>
          <tr v-for="asset of assets" :key="asset.path">
            <td><a tabindex="0" @click="onOpenClick(asset)">{{ asset.path }}</a></td>
            <td>{{ asset.mime_type }}</td>
            <td><abbr :title="`${formatNumber(asset.size)} bytes`">{{ formatFileSize(asset.size) }}</abbr></td>
            <td style="width: 160px">{{ formatDate(asset.created_at) }}</td>
            <td class="text-right text-nowrap" style="width: 60px">
              <button type="button" class="btn btn-xs btn-outline-danger" @click="onDeleteClick(asset)">
                <i class="far fa-fw fa-trash-alt" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <asset-form ref="form" />
  </div>
</template>

<script>
  import ConfirmDelete from './modals/ConfirmDelete'
  import AssetForm from './modals/AssetForm'

  export default {
    name: 'Assets',
    metaInfo () {
      return {
        title: `Assets: ${this.profile.name || '...'}`
      }
    },
    components: { AssetForm },
    data () {
      return {
        profile: {id: this.$route.params.id},
        assets: []
      }
    },
    methods: {
      doLoadProfile () {
        this.$parent.profiles.length = 0
        this.$parent.routes.length = 0
        this.$parent.users.length = 0

        return this.$http.get(`/api/profiles/${this.profile.id}`)
          .then(response => response.json())
          .then(profile => {
            this.$root.nav.profile = profile
            this.$parent.profiles.push(profile)
            this.profile = Object.assign({}, this.profile, profile)
          })
          .catch(this.$root.httpError)
      },
      doLoadAssets () {
        this.assets.length = 0
        this.$parent.assets.length = 0

        return this.$http.get(`/api/profiles/${this.profile.id}/assets`)
          .then(response => response.json())
          .then(assets => {
            this.assets.push(...assets)
            this.$parent.assets.push(...assets)
          })
          .catch(this.$root.httpError)
      },
      onAddClick () {
        this.$refs.form.show()
      },
      onDeleteClick (asset) {
        this.$root.showModal(ConfirmDelete, this, 'show', `the asset with path <code>${asset.path}</code>`, (e, modal) => {
          const elSubmit = e.target

          this.$root.lockSubmit(elSubmit)
          this.$http.delete(`/api/profiles/${this.profile.id}/assets/${asset.path}`)
            .then(() => this.doLoadAssets())
            .catch(this.$root.httpError)
            .finally(() => {
              modal.$refs.modal.hide()
              this.$root.unlockSubmit(elSubmit)
            })
        })
      },
      onOpenClick (asset) {
        if (asset.size <= 1024 * 1024 * 5 || confirm('Size of the requested asset is more than 5 MB. Do you still want open it?')) {
          open(`/api/profiles/${this.profile.id}/assets/${asset.path}`, '_blank').focus()
        }
      }
    },
    created () {
      this.$root.startLoading()
      this.$root.resetNavigation()

      this.doLoadProfile()
        .then(() => {
          return this.doLoadAssets()
        })
        .finally(() => {
          this.$root.stopLoading()
        })
    }
  }
</script>

<style scoped>

</style>
