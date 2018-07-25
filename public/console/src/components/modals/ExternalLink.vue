<template>
  <modal ref="modal" title="External URL">
    <div ref="url" class="text-monospace" style="overflow-x: auto">
      <span><span
        v-if="host.includes(':')"
      >{{ scheme }}</span><a
        v-else
        v-tooltip="'Click to switch scheme'"
        data-boundary="viewport"
        tabindex="0"
        @click="onSchemeClick"
      >{{ scheme }}</a></span>://{{ host }}<template
        v-for="segment of segments"
      ><a
        v-if="segment[0] === '{' && segment[segment.length-1] === '}'"
        v-tooltip="`Click to set ${segment}`"
        data-boundary="viewport"
        contenteditable="true"
        tabindex="0"
        @click="onSegmentClick"
        @input="onSegmentInput"
      >{{ segment }}</a><span
        v-else
      >{{ segment }}</span></template>
    </div>

    <template slot="buttons">
      <button type="button" class="btn btn-success" :disabled="/{\w+}/.test(url)" @click="onOpenClick">
        <i class="fas fa-external-link-alt" /> Open
      </button>
    </template>
  </modal>
</template>

<script>
  import Modal from '../bootstrap/Modal'

  export default {
    name: 'ExternalLink',
    components: { Modal },
    data () {
      return {
        scheme: 'http',
        host: '',
        segments: [],
        url: ''
      }
    },
    methods: {
      show (host, path) {
        this.scheme = 'http'
        this.host = host
        this.path = path

        this.segments.length = 0
        this.segments.push(...path.split(/({\w+})/))

        this.$nextTick(() => this.url = this.$refs.url.textContent)

        this.$refs.modal.show()
      },
      onSchemeClick () {
        this.scheme = this.scheme === 'http' ? 'https' : 'http'
        this.url = this.$refs.url.textContent
      },
      onSegmentClick () {
        document.execCommand('selectAll', false, null)
      },
      onSegmentInput () {
        this.url = this.$refs.url.textContent
      },
      onOpenClick () {
        const url = this.$refs.url.textContent
        open(url, '_blank').focus()
        this.$refs.modal.hide()
      }
    }
  }
</script>
