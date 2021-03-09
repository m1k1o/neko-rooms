<template>
  <v-btn v-if="tmpl" @click="Action" :color="tmpl.color" :disabled="disabled" :loading="loading" icon><v-icon>{{ tmpl.icon }}</v-icon></v-btn>
</template>

<script lang="ts">
import { Vue, Component, Prop } from 'vue-property-decorator'

@Component
export default class RoomActionBtn extends Vue {
  @Prop(String) readonly action: string | undefined
  @Prop(String) readonly roomId: string | undefined
  @Prop(Boolean) readonly disabled: boolean | undefined

  private loading = false

  // eslint-disable-next-line
  get tmpl(): any | undefined {
    switch (this.action) {
      case 'start':
        return {
          dispatch: 'ROOMS_START',
          msg: 'Room started!',
          color: 'green',
          icon: 'mdi-play-circle-outline',
        }
      case 'stop':
        return {
          dispatch: 'ROOMS_STOP',
          msg: 'Room stopped!',
          color: 'warning',
          icon: 'mdi-stop-circle-outline',
        }
      case 'restart':
        return {
          dispatch: 'ROOMS_RESTART',
          msg: 'Room restarted!',
          color: 'blue',
          icon: 'mdi-refresh',
        }
      case 'remove':
        return {
          dispatch: 'ROOMS_REMOVE',
          msg: 'Room removed!',
          color: 'red',
          icon: 'mdi-trash-can-outline',
        }
    }

    return undefined
  }

  async Action() {
    if (!this.tmpl) return

    if (this.action === 'remove') {
      const { value } = await this.$swal({
        title: "Remove room",
        text: "Do you really want to remove this room?",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: "Yes",
        cancelButtonText: "No",
      })

      if (!value) return
    }

    this.loading = true
  
    try {
      await this.$store.dispatch(this.tmpl.dispatch, this.roomId)
      this.$swal({
        title: this.tmpl.msg,
        icon: 'success',
      })
    } catch(e) {
      if (e.response) {
        this.$swal({
          title: 'Server error',
          text: e.response.data,
          icon: 'error',
        })
      } else {
        this.$swal({
          title: 'Network error',
          text: String(e),
          icon: 'error',
        })
      }
    } finally {
      this.loading = false
    }
  }
}
</script>
