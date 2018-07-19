<template>
  <div>
    <h1>Profiles</h1>
    <p class="lead">
      This is the list of your currently available profiles.
      <span v-if="['admin', 'manager'].includes($root.auth.role)">
        If you want to add a new profile <a tabindex="0" @click="onAddClick">click here</a>.
      </span>
    </p>
    <div class="table-responsive">
      <table class="table table-sm table-hover">
        <caption class="font-weight-bold">Total: {{ profiles.length }}</caption>
        <thead>
        <tr>
          <th>Name</th>
          <th style="width: 280px">Hosts</th>
          <th style="width: 300px">Backend</th>
          <th style="width: 150px">Vars Lifetime (s)</th>
          <th colspan="2">Created At</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="profile of profiles" :key="profile.id">
          <td class="d-flex justify-content-between align-items-center">
            <router-link :to="{name: 'routes', params: {id: profile.id}}">{{ profile.name }}</router-link>
            <span v-if="profile.is_debug" class="badge badge-danger">Debug</span>
            <span v-else class="badge badge-success">Production</span>
          </td>
          <td><span v-for="host of profile.hosts">{{ host }}<br></span></td>
          <td>
            <a v-if="profile.backend" :href="profile.backend" target="_blank">{{ profile.backend }}</a>
            <span v-else>n/a</span>
          </td>
          <td>{{ formatNumber(profile.vars_lifetime) }}</td>
          <td style="width: 160px">{{ formatDate(profile.created_at) }}</td>
          <td class="text-right text-nowrap" style="width: 80px">
            <button type="button" class="btn btn-xs btn-outline-success" @click="onEditClick(profile)">
              <i class="fas fa-fw fa-pencil-alt" />
            </button>
            <button v-if="['admin', 'manager'].includes($root.auth.role)" type="button" class="btn btn-xs btn-outline-danger" @click="onDeleteClick(profile)">
              <i class="far fa-fw fa-trash-alt" />
            </button>
          </td>
        </tr>
        </tbody>
      </table>
    </div>
    <profile-form ref="form" />
  </div>
</template>

<script>
  import ProfileForm from './modals/ProfileForm'
  import ConfirmDelete from './modals/ConfirmDelete'

  export default {
    name: 'Profiles',
    metaInfo: {
      title: 'Profiles'
    },
    components: { ProfileForm },
    data () {
      return {
        profiles: []
      }
    },
    methods: {
      doLoadProfiles () {
        this.profiles.length = 0

        this.$parent.profiles.length = 0
        this.$parent.routes.length = 0
        this.$parent.assets.length = 0
        this.$parent.users.length = 0

        return this.$http.get('/api/profiles')
          .then(response => response.json())
          .then(profiles => {
            this.profiles.push(...profiles)
            this.$parent.profiles.push(...profiles)
          })
          .catch(this.$root.httpError)
      },
      onAddClick () {
        this.$refs.form.showCreate()
      },
      onEditClick (profile) {
        this.$refs.form.showUpdate(profile.id)
      },
      onDeleteClick (profile) {
        this.$root.showModal(ConfirmDelete, this, 'show', `profile "${profile.name}"`, (e, modal) => {
          const elSubmit = e.target

          this.$root.lockSubmit(elSubmit)
          this.$http.delete(`/api/profiles/${profile.id}`)
            .then(() => this.doLoadProfiles())
            .catch(this.$root.httpError)
            .finally(() => {
              modal.$refs.modal.hide()
              this.$root.unlockSubmit(elSubmit)
            })
        }, profile.name)
      }
    },
    created () {
      this.$root.startLoading()
      this.$root.resetNavigation()

      this.doLoadProfiles()
        .finally(() => {
          this.$root.stopLoading()
        })
    }
  }
</script>
