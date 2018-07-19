<template>
  <main role="main" class="px-4">
    <h1 class="display-4 mt-4">404 Not Found</h1>
    <p class="lead">The requested URL was not found on this server.</p>
  </main>
</template>

<script>
  export default {
    name: '404',
    metaInfo: {
      title: '404 Not Found'
    },
    created () {
      this.$root.stopLoading()
    }
  }
</script>
