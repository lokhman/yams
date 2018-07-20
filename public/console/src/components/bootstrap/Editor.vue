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
        exec (editor) {
          vm.$emit('save', editor.getValue())
        }
      })

      this.$emit('init', editor)
      this.editor = editor
    }
  }
</script>
