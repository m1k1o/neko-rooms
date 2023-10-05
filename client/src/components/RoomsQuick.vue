<template>
  <span>
    <v-menu bottom left close-on-click>
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          v-bind="attrs"
          v-on="on"
          :loading="loading"
          color="info"
          dark
        >
          + Quick room
        </v-btn>
      </template>

      <v-list>
        <v-list-item
          v-for="(neko_image, index) in nekoImages"
          :key="index"
          @click="Action(neko_image)"
          link
        >
          <v-list-item-title>{{ neko_image }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>

    <v-dialog v-model="dialog" max-width="920px">
      <v-card>
        <v-card-title class="headline">
          Room information
        </v-card-title>
        <v-card-text>
          <RoomInfo v-if="dialog" :roomId="roomId" />
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
import { AxiosError } from 'axios'
import { randomPassword } from '@/utils/random'

import RoomInfo from '@/components/RoomInfo.vue'

@Component({
  components: {
    RoomInfo,
  }
})
export default class RoomActionBtn extends Vue {
  public dialog = false
  public loading = false
  public roomId = ''

  get nekoImages() {
    return this.$store.state.roomsConfig.neko_images
  }

  // eslint-disable-next-line
  async Action(neko_image: string) {
    this.loading = true
  
    try {
      const entry = await this.$store.dispatch('ROOMS_CREATE_AND_START', {
        ...this.$store.state.defaultRoomSettings,
        // eslint-disable-next-line
        neko_image,
        // eslint-disable-next-line
        user_pass: randomPassword(),
        // eslint-disable-next-line
        admin_pass: randomPassword(),
      })

      this.roomId = entry.id
      this.dialog = true
    } catch(e) {
      const response = (e as AxiosError).response
      if (response) {
        this.$swal({
          title: 'Server error',
          text: String(response.data),
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
