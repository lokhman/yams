<template>
  <div>
    <h1>Routes <small class="text-muted">{{ profile.name || '...' }}</small></h1>
    <p class="lead">
      This is the list of available routes in profile "{{ profile.name || '...' }}".
      If you want to add a new route <a tabindex="0" @click="onAddClick">click here</a>.
    </p>
    <div class="table-responsive">
      <table class="table table-sm table-hover">
        <caption class="font-weight-bold">Total: {{ routes.length }}</caption>
        <thead>
        <tr>
          <th colspan="2">Path</th>
          <th style="width: 300px">Methods</th>
          <th style="width: 200px">Script</th>
          <th colspan="2">Timeout (s)</th>
        </tr>
        </thead>
        <draggable v-model="routes" element="tbody" :options="{handle: '.yams-handle'}" @update="onPositionChange">
          <tr v-for="route of routes" :key="route.id" :data-id="route.id">
            <td style="width: 10px">
              <span class="text-black-50 yams-handle">
                <i class="fas fa-ellipsis-v" />
              </span>
            </td>
            <td class="d-flex justify-content-between">
              <div>
                <div class="text-monospace align-middle" :class="{'text-line-through': !route.is_enabled}">
                  <a v-if="route.is_enabled && route.methods.includes('GET')" tabindex="0" @click="onOpenClick(route)">{{ route.path }}</a>
                  <span v-else>{{ route.path }}</span>
                </div>
                <small>{{ route.hint }}</small>
              </div>
              <div class="yams-state">
                <button type="button" class="btn btn-xs btn-link" @click="onToggleClick(route)">
                  <span v-show="route.is_enabled" v-tooltip="'Click to disable'" data-placement="left"><i class="fas fa-ban" /></span>
                  <span v-show="!route.is_enabled" v-tooltip="'Click to enable'" data-placement="left"><i class="fas fa-check" /></span>
                </button>
              </div>
            </td>
            <td>
              <span v-for="method of route.methods" :class="`badge badge-${$parent.getMethodClassName(method)} mr-1`">
                {{ method }}
              </span>
            </td>
            <td>
              <button type="button" class="btn btn-xs btn-outline-primary" style="min-width: 130px" @click="onScriptClick(route)">
                <i class="fas fa-code" /> {{ route.adapter }}
                <code class="badge badge-primary">{{ formatFileSize(route.script_size, '') }}</code>
              </button>
            </td>
            <td style="width: 100px">{{ formatNumber(route.timeout) }}</td>
            <td class="text-right text-nowrap" style="width: 80px">
              <button type="button" class="btn btn-xs btn-outline-success" @click="onEditClick(route)">
                <i class="fas fa-fw fa-pencil-alt" />
              </button>
              <button type="button" class="btn btn-xs btn-outline-danger" @click="onDeleteClick(route)">
                <i class="far fa-fw fa-trash-alt" />
              </button>
            </td>
          </tr>
        </draggable>
      </table>
    </div>
    <route-form ref="form" />
    <script-editor ref="script" />
  </div>
</template>

<script>
  import RouteForm from './modals/RouteForm'
  import ConfirmDelete from './modals/ConfirmDelete'
  import ExternalLink from './modals/ExternalLink'
  import ScriptEditor from './modals/ScriptEditor'

  export default {
    name: 'Routes',
    metaInfo () {
      return {
        title: `Routes: ${this.profile.name || '...'}`
      }
    },
    components: { RouteForm, ScriptEditor },
    data () {
      return {
        profile: {id: this.$route.params.id},
        routes: []
      }
    },
    methods: {
      doLoadProfile () {
        this.$parent.profiles.length = 0
        this.$parent.assets.length = 0
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
      doLoadRoutes () {
        this.routes.length = 0
        this.$parent.routes.length = 0

        return this.$http.get(`/api/profiles/${this.profile.id}/routes`)
          .then(response => response.json())
          .then(routes => {
            this.routes.push(...routes)
            this.$parent.routes.push(...routes)
          })
          .catch(this.$root.httpError)
      },
      onPositionChange (e) {
        const id = e.item.dataset.id
        const position = e.newIndex

        this.$http.post(`/api/routes/${id}/position`, {position})
          .then(() => {
            this.$parent.routes.length = 0
            this.$parent.routes.push(...this.routes)
          })
          .catch(response => {
            this.$root.httpError(response)
            this.doLoadRoutes()
          })
      },
      onToggleClick (route) {
        route.is_enabled = !route.is_enabled

        this.$http.post(`/api/routes/${route.id}/state`, {is_enabled: route.is_enabled})
          .catch(response => {
            this.$root.httpError(response)
            route.is_enabled = !route.is_enabled  // rollback
          })
      },
      onAddClick () {
        this.$refs.form.showCreate()
      },
      onEditClick (route) {
        this.$refs.form.showUpdate(route.id)
      },
      onDeleteClick (route) {
        this.$root.showModal(ConfirmDelete, this, 'show', `the route to <code>${route.path}</code>`, (e, modal) => {
          const elSubmit = e.target

          this.$root.lockSubmit(elSubmit)
          this.$http.delete(`/api/routes/${route.id}`)
            .then(() => this.doLoadRoutes())
            .catch(this.$root.httpError)
            .finally(() => {
              modal.$refs.modal.hide()
              this.$root.unlockSubmit(elSubmit)
            })
        })
      },
      onOpenClick (route) {
        this.$root.showModal(ExternalLink, this, 'show', this.profile.hosts[0], route.path)
      },
      onScriptClick (route) {
        this.$refs.script.show(route)
      }
    },
    created () {
      this.$root.startLoading()
      this.$root.resetNavigation()

      this.doLoadProfile()
        .then(() => {
          return this.doLoadRoutes()
        })
        .finally(() => {
          this.$root.stopLoading()
        })
    }
  }
</script>

<style lang="scss" scoped>
  .yams-handle {
    cursor: move;
  }
  .yams-state {
    visibility: hidden;

    tr:hover & {
      visibility: visible;
    }
  }
</style>
