<script>
  const COMPLETIONS = {
    lua: {
      '[yams] field': ['routeid', 'method', 'host', 'uri', 'ip', 'sessionid', 'form', 'args', 'headers', 'query', 'cookies'],
      '[yams] function': ['setstatus', 'getheader', 'setheader', 'setcookie', 'parseform', 'getparam', 'getbody', 'asset', 'sleep', 'write', 'getvar', 'setvar', 'dump', 'wbclean', 'pass', 'exit'],
      '[yams:asset] function': ['getmimetype', 'getsize', 'template'],
      '[json/base64] function': ['encode', 'decode']
    }
  }

  export default {
    name: 'Editor',
    props: {
      value: {type: String, required: true},
      mode: {type: String, default: 'text'}
    },
    data () {
      return {
        editor: null,
        langTools: null,
        backup: ''
      }
    },
    watch: {
      value (value) {
        if (this.backup !== value) {
          this.editor.setValue(value, 1)
          this.backup = value
        }
      },
      mode (mode) {
        this.editor.session.setMode(`ace/mode/${mode}`)

        if (mode in COMPLETIONS) {
          const map = COMPLETIONS[mode]
          const completions = []

          for (let meta in map) {
            completions.push(...map[meta].map(word => ({
              caption: word,
              value: word,
              meta: meta
            })))
          }

          ace.require('ace/ext/language_tools').addCompleter({
            getCompletions (editor, session, pos, prefix, callback) {
              callback(null, completions)
            }
          })

          this.editor.setOption('enableBasicAutocompletion', true)
          this.editor.setOption('enableSnippets', true)
          this.editor.setOption('enableLiveAutocompletion', true)
        }
      }
    },
    render (h) {
      return h('div')
    },
    beforeDestroy () {
      this.editor.destroy()
      this.editor.container.remove()
    },
    mounted () {
      const vm = this
      const editor = ace.edit(this.$el)
      const parentNode = this.$el.parentNode
      const baseNode = this.$el.closest('.modal') || document.body

      editor.$blockScrolling = Infinity
      editor.setPrintMarginColumn(100)
      editor.setValue(this.value, 1)
      this.backup = this.value

      editor.on('change', () => {
        const value = editor.getValue()
        this.$emit('input', value)
        this.backup = value
      })

      editor.commands.addCommand({
        name: 'save',
        bindKey: {win: 'Ctrl-S', mac: 'Cmd-S'},
        exec () {
          vm.$emit('save', editor.getValue())
        }
      })

      editor.commands.addCommand({
        name: 'fullscreen',
        bindKey: {win: 'F11', mac: 'Cmd-Ctrl-F'},
        exec () {
          if (vm.$el.parentNode === baseNode) {
            parentNode.insertBefore(vm.$el, parentNode.firstChild)
            vm.$el.classList.remove('yams-fullscreen')
          } else {
            baseNode.appendChild(vm.$el)
            vm.$el.classList.add('yams-fullscreen')
          }
          editor.resize()
          editor.focus()
        }
      })

      this.$emit('init', editor)
      this.editor = editor
    }
  }
</script>

<style lang="scss" scoped>
  .yams-fullscreen {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    z-index: 999999999;
  }
</style>
