<template>
  <div>
    <div v-if="loading" class="text-center">
      <v-progress-circular
        :size="70"
        :width="7"
        color="blue"
        indeterminate
      ></v-progress-circular>
    </div>
    <v-alert
      border="left"
      type="warning"
      v-else-if="!settings"
    >
      Room not found!
    </v-alert>
    <template v-else>
      <div class="my-3 headline">Main settings</div>
      <v-simple-table>
        <template v-slot:default>
          <tbody>
            <tr><th style="width:50%;"> Name </th><td>{{ settings.name }}</td></tr>
            <tr><th> Max connections </th><td>{{ settings.max_connections }}</td></tr>
            <tr><th> User pass </th><td>
              <v-btn @click="showUserPass = !showUserPass" class="mr-2" icon small><v-icon small>{{ showUserPass ? 'mdi-eye' : 'mdi-eye-off' }}</v-icon></v-btn>
              <span></span>{{ showUserPass ? settings.user_pass : '****' }}
            </td></tr>
            <tr><th> Admin pass </th><td>
              <v-btn @click="showAdminPass = !showAdminPass" class="mr-2" icon small><v-icon small>{{ showAdminPass ? 'mdi-eye' : 'mdi-eye-off' }}</v-icon></v-btn>
              <span></span>{{ showAdminPass ? settings.admin_pass : '****' }}
            </td></tr>
          </tbody>
        </template>
      </v-simple-table>
        
      <div class="my-3 headline">Video settings</div>
      <v-simple-table>
        <template v-slot:default>
          <tbody>
            <tr><th style="width:50%;"> Screen </th><td>{{ settings.screen }}</td></tr>
            <tr><th> Video codec </th><td>{{ settings.video_codec }}</td></tr>
            <tr><th> Video bitrate </th><td>{{ settings.video_bitrate }}</td></tr>
            <tr><th> Video pipeline </th><td>{{ settings.video_pipeline }}</td></tr>
            <tr><th> Video max fps </th><td>{{ settings.video_max_fps }}</td></tr>
          </tbody>
        </template>
      </v-simple-table>
        
      <div class="my-3 headline">Audio settings</div>
      <v-simple-table>
        <template v-slot:default>
          <tbody>
            <tr><th style="width:50%;"> Audio codec </th><td>{{ settings.audio_codec }}</td></tr>
            <tr><th> Audio bitrate </th><td>{{ settings.audio_bitrate }}</td></tr>
            <tr><th> Audio pipeline </th><td>{{ settings.audio_pipeline }}</td></tr>
          </tbody>
        </template>
      </v-simple-table>
        
      <div class="my-3 headline">Broadcast settings</div>
      <v-simple-table>
        <template v-slot:default>
          <tbody>
            <tr><th style="width:50%;"> Broadcast pipeline </th><td>{{ settings.broadcast_pipeline }}</td></tr>
          </tbody>
        </template>
      </v-simple-table>
    </template>
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

  private showUserPass = false
  private showAdminPass = false

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
