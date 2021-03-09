<template>
  <div>
    {{ loading ? 'loading' : settings }}
  </div>
</template>

<script lang="ts">
import { Vue, Component, Prop, Watch } from 'vue-property-decorator'
import { RoomSettings } from '@/api/index'

@Component
export default class RoomInfo extends Vue {
  @Prop(String) readonly roomId: string | undefined

  private loading = false
  private settings: RoomSettings | null = null

  @Watch('roomId', { immediate: true })
  async SetRoomId(roomId: string) {
    this.loading = true
  
    try {
      this.settings = await this.$store.dispatch('ROOMS_GET', roomId)
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
