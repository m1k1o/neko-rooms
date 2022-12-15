<template>
  <v-tooltip bottom v-if="tmpl" open-delay="300">
    <template v-slot:activator="{ on, attrs }">
      <v-btn v-bind="attrs" v-on="on" @click="Action" :color="tmpl.color" :disabled="disabled" :loading="loading" icon><v-icon>{{ tmpl.icon }}</v-icon></v-btn>
    </template>
    <span>{{ tmpl.tooltip }}</span>
  </v-tooltip>
</template>

<script lang="ts">
import { Vue, Component, Prop } from 'vue-property-decorator'

@Component
export default class RoomActionBtn extends Vue {
  @Prop(String) readonly action!: string
  @Prop(String) readonly roomId!: string
  @Prop(Boolean) readonly disabled!: boolean

  private loading = false

  // eslint-disable-next-line
  get tmpl(): any | undefined {
    switch (this.action) {
      case 'start':
        return {
          dispatch: 'ROOMS_START',
          msg: 'Room started!',
          tooltip: 'Start',
          color: 'green',
          icon: 'mdi-play-circle-outline',
        }
      case 'stop':
        return {
          dispatch: 'ROOMS_STOP',
          msg: 'Room stopped!',
          tooltip: 'Stop',
          color: 'warning',
          icon: 'mdi-stop-circle-outline',
        }
      case 'restart':
        return {
          dispatch: 'ROOMS_RESTART',
          msg: 'Room restarted!',
          tooltip: 'Restart',
          color: 'blue',
          icon: 'mdi-refresh',
        }
      case 'recreate':
        return {
          dispatch: 'ROOMS_RECREATE',
          msg: 'Room recreated!',
          tooltip: 'Recreate',
          color: 'blue',
          icon: 'mdi-cloud-refresh',
        }
      case 'remove':
        return {
          dispatch: 'ROOMS_REMOVE',
          msg: 'Room removed!',
          tooltip: 'Remove',
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
    } else if (this.action === 'recreate') {
      const { value } = await this.$swal({
        title: "Recreate room",
        text: "Do you really want to recreate this room? It will delete all your non-persistent data.",
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
