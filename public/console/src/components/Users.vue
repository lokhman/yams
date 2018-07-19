<template>
  <div>
    <h1>Users</h1>
    <p class="lead">
      This is the list of your currently available users.
      If you want to add a new user <a tabindex="0" @click="onAddClick">click here</a>.
    </p>
    <div class="table-responsive">
      <table class="table table-sm table-hover">
        <caption class="font-weight-bold">Total: {{ users.length }}</caption>
        <thead>
        <tr>
          <th style="width: 250px">Username</th>
          <th style="width: 180px">Role</th>
          <th>ACL</th>
          <th style="width: 180px">Last Auth At</th>
          <th colspan="2">Created At</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="user of users" :key="user.id">
          <td>{{ user.username }}</td>
          <td><strong :class="`text-${$parent.getRoleClassName(user.role)}`">{{ user.role }}</strong></td>
          <td>
            <span v-if="!user.acl">n/a</span>
            <template v-else v-for="(profile, index) of profiles.filter(profile => user.acl.includes(profile.id))">
              <template v-if="index">, </template> <router-link :to="{name: 'routes', params: {id: profile.id}}">{{ profile.name }}</router-link>
            </template>
          </td>
          <td>{{ user.last_auth_at ? formatDate(user.last_auth_at) : 'n/a' }}</td>
          <td style="width: 160px">{{ formatDate(user.created_at) }}</td>
          <td class="text-right text-nowrap" style="width: 80px">
            <button type="button" class="btn btn-xs btn-outline-success" @click="onEditClick(user)">
              <i class="fas fa-fw fa-pencil-alt" />
            </button>
            <button type="button" class="btn btn-xs btn-outline-danger" :disabled="user.id === $root.auth.id" @click="onDeleteClick(user)">
              <i class="far fa-fw fa-trash-alt" />
            </button>
          </td>
        </tr>
        </tbody>
      </table>
    </div>
    <user-form ref="form" />
  </div>
</template>

<script>
  import UserForm from './modals/UserForm'
  import ConfirmDelete from './modals/ConfirmDelete'

  export default {
    name: 'Users',
    metaInfo: {
      title: 'Users'
    },
    components: { UserForm },
    data () {
      return {
        users: [],
        profiles: []
      }
    },
    methods: {
      doLoadUsers () {
        this.users.length = 0
        this.$parent.profiles.length = 0
        this.$parent.users.length = 0

        return this.$http.get('/api/users')
          .then(response => response.json())
          .then(users => {
            this.users.push(...users)
            this.$parent.users.push(...users)
          })
          .catch(this.$root.httpError)
      },
      doLoadProfiles() {
        this.profiles.length = 0

        return this.$http.get('/api/profiles?preview=true')
          .then(response => response.json())
          .then(profiles => {
            this.profiles.push(...profiles)
          })
          .catch(this.$root.httpError)
      },
      onAddClick () {
        this.$refs.form.showCreate()
      },
      onEditClick (user) {
        this.$refs.form.showUpdate(user.id)
      },
      onDeleteClick (user) {
        this.$root.showModal(ConfirmDelete, this, 'show', `user "${user.username}"`, (e, modal) => {
          const elSubmit = e.target

          this.$root.lockSubmit(elSubmit)
          this.$http.delete(`/api/users/${user.id}`)
            .then(() => this.doLoadUsers())
            .catch(this.$root.httpError)
            .finally(() => {
              modal.$refs.modal.hide()
              this.$root.unlockSubmit(elSubmit)
            })
        }, user.username)
      }
    },
    created () {
      this.$root.startLoading()
      this.$root.resetNavigation()

      this.doLoadUsers()
        .then(() => {
          return this.doLoadProfiles()
        })
        .finally(() => {
          this.$root.stopLoading()
        })
    }
  }
</script>
