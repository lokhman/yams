<template>
  <main role="main" class="px-4">
    <h1 class="display-4 mt-4">403 Forbidden</h1>
    <p class="lead">You are not allowed to access the requested URL.</p>
  </main>
</template>

<script>
  export default {
    name: '403',
    metaInfo: {
      title: '403 Forbidden'
    },
    created () {
      this.$root.stopLoading()
    }
  }
</script>
