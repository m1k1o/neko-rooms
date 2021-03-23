<template>
  <div>
    <div v-if="settingsLoading" class="text-center">
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
      <p><strong>Room not loaded!</strong></p>
      <p class="mb-0">Check your internet connectivity. Try to recreate room.</p>
    </v-alert>
    <template v-else>
      <div class="my-3 headline">Room members</div>
      <v-row v-if="stats">
        <v-col class="text-center">
          <div class="mb-3">
            <v-progress-circular
              :rotate="270"
              :size="100"
              :width="15"
              :value="(stats.connections / settings.max_connections) * 100"
              color="blue"
            >
              {{ stats.connections }} / {{ settings.max_connections }}
            </v-progress-circular>
          </div>
        </v-col>
        <v-col>
          <v-simple-table>
            <template v-slot:default>
              <thead>
                <tr>
                  <th> Dispaly name </th>
                  <th class="text-center"> Is Hosting </th>
                  <th class="text-center"> Is Admin </th>
                  <th class="text-center"> Is Muted </th>
                </tr>
              </thead>
              <tbody v-if="stats.members.length > 0">
                <tr v-for="member in stats.members" :key="member.id">
                  <td>{{ member.displayname }}</td>
                  <td class="text-center">
                    <v-icon v-if="stats.host == member.id" small color="green">mdi-keyboard</v-icon>
                    <v-icon v-else small color="red">mdi-keyboard</v-icon>
                  </td>
                  <td class="text-center">
                    <v-icon v-if="member.admin" small color="green">mdi-shield-check</v-icon>
                    <v-icon v-else small color="red">mdi-shield-off</v-icon>
                  </td>
                  <td class="text-center">
                    <v-icon v-if="member.muted" small color="red">mdi-volume-mute</v-icon>
                    <v-icon v-else small color="green">mdi-volume-high</v-icon>
                  </td>
                </tr>
              </tbody>
              <tbody v-else>
                <tr>
                  <td style="pointer-events: none;" colspan="4" class="text-center">no members</td>
                </tr>
              </tbody>
            </template>
          </v-simple-table>
        </v-col>
      </v-row>
      <div class="text-center mt-3">
        <v-btn @click="LoadStats" :loading="statsLoading">Reload</v-btn>
      </div>

      <div class="my-3 headline">Main settings</div>
      <v-simple-table>
        <template v-slot:default>
          <tbody>
            <tr><th style="width:50%;"> Name </th><td>{{ settings.name }}</td></tr>
            <tr><th> Neko image </th><td>{{ settings.neko_image }}</td></tr>
            <tr><th> User password </th><td>
              <v-btn @click="showUserPass = !showUserPass" icon small><v-icon small>{{ showUserPass ? 'mdi-eye' : 'mdi-eye-off' }}</v-icon></v-btn>
              <span class="mx-2">{{ showUserPass ? settings.user_pass : '****' }}</span>
              <v-tooltip top v-if="room">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn v-bind="attrs" v-on="on" :disabled="!room.running" :href="room.url + '?pwd=' + encodeURIComponent(settings.user_pass)" target="_blank" small> <v-icon small>mdi-open-in-new</v-icon></v-btn>
                </template>
                <span>Invite link for users</span>
              </v-tooltip>
            </td></tr>
            <tr><th> Admin password </th><td>
              <v-btn @click="showAdminPass = !showAdminPass" icon small><v-icon small>{{ showAdminPass ? 'mdi-eye' : 'mdi-eye-off' }}</v-icon></v-btn>
              <span class="mx-2">{{ showAdminPass ? settings.admin_pass : '****' }}</span>
              <v-tooltip bottom v-if="room">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn  v-bind="attrs" v-on="on" :disabled="!room.running" :href="room.url + '?pwd=' + encodeURIComponent(settings.admin_pass)" target="_blank" small> <v-icon small>mdi-open-in-new</v-icon></v-btn>
                </template>
                <span>Invite link for admins</span>
              </v-tooltip>
            </td></tr>
            <tr><th> Max connections </th><td>{{ settings.max_connections }}</td></tr>
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
import { RoomStats, RoomMember, RoomSettings, RoomEntry } from '@/api/index'

@Component
export default class RoomInfo extends Vue {
  @Prop(String) readonly roomId: string | undefined

  private statsLoading = false
  private stats: RoomStats | null = null

  private settingsLoading = false
  private settings: RoomSettings | null = null

  private showUserPass = false
  private showAdminPass = false

  get room(): RoomEntry {
    return this.$store.state.rooms.find(({ id }: RoomEntry) => id == this.roomId)
  }

  @Watch('roomId', { immediate: true })
  async SetRoomId(roomId: string) {
    this.stats = null
    this.settings = null
    this.settingsLoading = true
  
    try {
      this.settings = await this.$store.dispatch('ROOMS_SETTINGS', roomId)
      this.LoadStats()
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
      this.settingsLoading = false
    }
  }

  async LoadStats() {
    this.statsLoading = true
  
    try {
      const stats = await this.$store.dispatch('ROOMS_STATS', this.roomId)
      stats.members.sort(function(a: RoomMember, b: RoomMember) {
        const nameA = a.displayname?.toUpperCase() || "";
        const nameB = b.displayname?.toUpperCase() || "";

        if (nameA < nameB) {
          return -1
        }

        if (nameA > nameB) {
          return 1
        }

        return 0
      })
      this.stats = stats
    } finally {
      this.statsLoading = false
    }
  }
}
</script>
