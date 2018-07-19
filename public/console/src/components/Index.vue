<template>
  <main role="main" class="col-md-9 ml-sm-auto col-lg-9 px-4">
    <aside class="container-fluid row">
      <nav class="col-md-3 d-none d-md-block bg-light yams-sidebar">
        <div class="yams-sidebar-sticky px-4">
          <div>
            <h6 class="yams-sidebar-heading d-flex justify-content-between align-items-center mt-3 mb-2 text-muted">
              <router-link :to="{name: 'profiles'}"><i class="fas fa-fw fa-cogs" /> Profiles</router-link>
              <a v-if="$router.currentRoute.name === 'profiles'" class="d-flex align-items-center yams-new" data-before="new profile" @click="onPageAddClick">
                <i class="fas fa-plus-circle" />
              </a>
            </h6>
            <ul class="list-unstyled">
              <li v-for="profile of profiles" :key="profile.id" class="d-flex justify-content-between align-items-center">
                <span v-tooltip="profile.hosts.join('<br>')" data-html="true" data-placement="right" data-offset="0,10">
                  <i class="far fa-fw fa-arrow-alt-circle-right" />
                  <span v-if="$root.profile">{{ profile.name }}</span>
                  <router-link v-else :to="{name: 'routes', params: {id: profile.id}}">{{ profile.name }}</router-link>
                </span>
                <span v-if="profile.is_debug" class="badge badge-danger">Debug</span>
                <span v-else class="badge badge-success">Production</span>
              </li>
            </ul>
          </div>

          <div v-if="$root.nav.profile">
            <h6 class="yams-sidebar-heading d-flex justify-content-between align-items-center mt-3 mb-2 text-muted">
              <router-link :to="{name: 'routes', params: {id: $root.nav.profile.id}}"><i class="fas fa-fw fa-exchange-alt" /> Routes</router-link>
              <a v-if="$router.currentRoute.name === 'routes'" class="d-flex align-items-center yams-new" data-before="new route" @click="onPageAddClick">
                <i class="fas fa-plus-circle" />
              </a>
            </h6>
            <ul class="list-unstyled yams-routes">
              <li v-for="route of routes" :key="route.id">
                <a @click="onPageEditClick(route)">
                  <span :class="`badge badge-${getMethodClassName(...route.methods)}`">
                    {{ getMethodShortName(route.methods) }}
                  </span>
                </a>
                <a v-tooltip="route.hint || ''" data-placement="right" data-offset="0,10" @click="onRouteScriptClick(route)">
                  <span :class="{'text-line-through': !route.is_enabled}" class="align-middle">
                    {{ route.path }}
                  </span>
                </a>
              </li>
            </ul>
          </div>

          <div v-if="$root.nav.profile">
            <h6 class="yams-sidebar-heading d-flex justify-content-between align-items-center mt-3 mb-2 text-muted">
              <router-link :to="{name: 'assets', params: {id: $root.nav.profile.id}}"><i class="fas fa-fw fa-copy" /> Assets</router-link>
              <a v-if="$router.currentRoute.name === 'assets'" class="d-flex align-items-center yams-new" data-before="new asset" @click="onPageAddClick">
                <i class="fas fa-plus-circle" />
              </a>
            </h6>
            <ul class="list-unstyled yams-assets">
              <li v-for="asset of assets" :key="asset.path" class="d-flex justify-content-between align-items-center">
                <a @click="onAssetOpenClick(asset)">
                  <i class="far fa-file" /> {{ asset.path }}
                </a>
                <span class="badge badge-success" :title="`${formatNumber(asset.size)} bytes`">
                  {{ formatFileSize(asset.size, '') }}
                </span>
              </li>
            </ul>
          </div>

          <div v-if="$root.auth.role === 'admin'">
            <h6 class="yams-sidebar-heading d-flex justify-content-between align-items-center mt-3 mb-2 text-muted">
              <router-link :to="{name: 'users'}"><i class="fas fa-fw fa-users" /> Users</router-link>
              <a v-if="$router.currentRoute.name === 'users'" class="d-flex align-items-center yams-new" data-before="new user" @click="onPageAddClick">
                <i class="fas fa-plus-circle" />
              </a>
            </h6>
            <ul class="list-unstyled">
              <li v-for="user of users" :key="user.id" class="d-flex justify-content-between align-items-center">
                <a @click="onPageEditClick(user)"><i class="far fa-user" /> {{ user.username }}</a>
                <span :class="`badge badge-${getRoleClassName(user.role)}`">{{ user.role }}</span>
              </li>
            </ul>
          </div>
        </div>
      </nav>
    </aside>
    <section>
      <div class="pt-3 mb-3">
        <transition name="page" mode="out-in">
          <router-view ref="page" />
        </transition>
      </div>
    </section>
  </main>
</template>

<script>
  import $ from 'jquery'

  const methodMap = {
    GET: {shortName: 'GET', className: 'success'},
    HEAD: {shortName: 'HEAD', className: 'dark'},
    POST: {shortName: 'POST', className: 'surprise'},
    PUT: {shortName: 'PUT', className: 'warning'},
    DELETE: {shortName: 'DEL', className: 'danger'},
    CONNECT: {shortName: 'CON', className: 'dark'},
    OPTIONS: {shortName: 'OPT', className: 'info'},
    TRACE: {shortName: 'TRC', className: 'dark'},
    PATCH: {shortName: 'PTCH', className: 'notice'}
  }

  const roleMap = {
    admin: {className: 'danger'},
    manager: {className: 'info'},
    developer: {className: 'success'}
  }

  export default {
    name: 'Index',
    data () {
      return {
        profiles: [],
        routes: [],
        assets: [],
        users: []
      }
    },
    methods: {
      getMethodClassName (method) {
        return method in methodMap ? methodMap[method].className : 'dark'
      },
      getMethodShortName (methods) {
        if (methods.length === 0) {
          return '???'
        }
        const method = methods[0]
        let shortName = method in methodMap
          ? methodMap[method].shortName : method.slice(0, 3)
        if (methods.length > 1) {
          shortName += '+'
        }
        return shortName
      },
      getRoleClassName (role) {
        return role in roleMap ? roleMap[role].className : 'light'
      },
      onPageAddClick () {
        this.$refs.page.onAddClick()
      },
      onPageEditClick (entity) {
        this.$refs.page.onEditClick(entity)
      },
      onRouteScriptClick (route) {
        this.$refs.page.onScriptClick(route)
      },
      onAssetOpenClick (asset) {
        this.$refs.page.onOpenClick(asset)
      }
    },
    beforeRouteEnter (to, from, next) {
      next(vm => {
        if (!vm.$root.token) {
          next({name: 'login'})
        }
      })
    },
    beforeRouteUpdate (to, from, next) {
      if (!this.$root.token) {
        next({name: 'login'})
      } else if ($('[role="dialog"]').is(':visible')) {
        next(false)
      } else {
        next()
      }
    }
  }
</script>

<style lang="scss" scoped>
  .yams-sidebar {
    bottom: 0;
    box-shadow: inset -1px 0 0 rgba(0, 0, 0, .1);
    left: 0;
    padding: 56px 0 0;
    position: fixed;
    top: 0;
    z-index: 100;

    ul {
      color: #333;
      font-size: .8rem;
      font-weight: 500;

      &.yams-routes,
      &.yams-assets {
        font-size: 13px;
        font-family: monospace;
      }

      &.yams-routes {
        .badge {
          min-width: 3.6em;
          letter-spacing: -0.3px;
        }
      }

      li {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    a:not(.badge) {
      color: inherit;
      text-decoration: none;

      &.active, &:hover {
        color: #007bff;
      }
    }
  }

  .yams-sidebar-sticky {
    height: calc(100vh - 56px);
    overflow-x: hidden;
    overflow-y: auto; /* Scrollable contents if viewport is shorter than content. */
    padding-top: .5rem;
    position: relative;
    top: 0;
  }

  @supports ((position: -webkit-sticky) or (position: sticky)) {
    .yams-sidebar-sticky {
      position: -webkit-sticky;
      position: sticky;
    }
  }

  .yams-sidebar-heading {
    font-size: .8rem;
    text-transform: uppercase;
  }

  .yams-new:hover {
    &::before {
      content: attr(data-before);
      margin-right: 3px;
      font-size: 10px;
    }
  }
</style>
