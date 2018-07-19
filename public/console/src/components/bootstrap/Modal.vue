<template>
  <div class="modal fade" tabindex="-1" role="dialog">
    <div class="modal-dialog" :class="{'modal-sm': size === 'small', 'modal-lg': size === 'large'}" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ title }}</h5>
          <button v-if="closable" type="button" class="close" data-dismiss="modal" aria-label="Close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <slot></slot>
        </div>
        <div class="modal-footer">
          <slot name="footer">
            <slot name="buttons"></slot>
            <button v-if="closable" type="button" class="btn btn-outline-secondary" data-dismiss="modal">
              <i class="fas fa-times" /> Close
            </button>
          </slot>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import $ from 'jquery'

  export default {
    name: 'Modal',
    props: {
      title: {type: String, required: true},
      closable: {type: Boolean, default: true},
      size: {type: String, validator: value => ['small', 'large'].includes(value)}
    },
    methods: {
      show () {
        $(this.$el).modal('show')
      },
      hide () {
        $(this.$el).modal('hide')
      }
    },
    mounted () {
      const vm = this

      $(this.$el).on({
        'show.bs.modal' (e) {
          vm.$emit('show', e)
        },
        'hide.bs.modal' (e) {
          vm.$emit('hide', e)
        }
      })
    }
  }
</script>
