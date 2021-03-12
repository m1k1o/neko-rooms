<template>
  <span>
    <v-btn @click="Action" :loading="loading" color="info" dark>
      + Quick room
    </v-btn>

    <v-dialog v-model="dialog" max-width="920px">
      <v-card>
        <v-card-title class="headline">
          Room information
        </v-card-title>
        <v-card-text>
          <RoomInfo :roomId="roomId" />
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="grey" text @click="dialog = false">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </span>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import RoomInfo from '@/components/RoomInfo.vue'

@Component({
  components: {
    RoomInfo,
  }
})
export default class RoomActionBtn extends Vue {
  private dialog = false
  private loading = false
  private roomId = ''

  async Action() {
    this.loading = true
  
    try {
      
      const entry = await this.$store.dispatch('ROOMS_CREATE', {
        ...this.$store.state.defaultRoomSettings,
        // eslint-disable-next-line
        user_pass: Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15),
        // eslint-disable-next-line
        admin_pass: Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15),
      })

      this.roomId = entry.id
      this.dialog = true
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
