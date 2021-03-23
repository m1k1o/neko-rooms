<template>
  <v-container>

    <!-- show off-site -->
    <div v-if="configConnections" class="text-center" style="position: absolute; left: 50%; transform: translate(-50%, -100%)">
      <p>Ports used:</p>
      <v-progress-circular
        :rotate="270"
        :size="100"
        :width="15"
        :value="(usedConnections / configConnections) * 100"
        :color="usedConnections == configConnections ? 'red' : 'blue'"
      >
        {{ usedConnections }} / {{ configConnections }}
      </v-progress-circular>
    </div>

    <v-row>
      <v-col>
        <v-btn @click="LoadRooms" class="mb-3" color="green" icon><v-icon>mdi-refresh</v-icon></v-btn>
      </v-col>
      <v-col class="text-right">
        <RoomsQuick class="mr-3" />
        <v-dialog
          v-model="dialog"
          persistent
          max-width="600px"
        >
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              v-bind="attrs"
              v-on="on"
              color="success"
              dark
            >
              + Add room
            </v-btn>
          </template>

          <RoomsCreate @finished="dialog = false" />
        </v-dialog>
      </v-col>
    </v-row>

    <RoomsList :loading="loading" />
  </v-container>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import RoomsList from '@/components/RoomsList.vue'
import RoomsQuick from '@/components/RoomsQuick.vue'
import RoomsCreate from '@/components/RoomsCreate.vue'
import { RoomEntry } from '@/api/index'

@Component({
  components: {
    RoomsList,
    RoomsQuick,
    RoomsCreate,
  }
})
export default class Home extends Vue {
  private loading = false
  private dialog = false

  get configConnections() {
    return this.$store.state.roomsConfig.connections
  }

  get usedConnections() {
    // eslint-disable-next-line
    return this.$store.state.rooms.reduce((sum: number, { max_connections }: RoomEntry) => sum + (max_connections || 0), 0)
  }

  async LoadRooms() {
    this.loading = true

    try {
      await this.$store.dispatch('ROOMS_LOAD')
    } finally {
      this.loading = false
    }
  }

  mounted() {
    this.$store.dispatch('ROOMS_CONFIG')
    this.LoadRooms()
  }
}
</script>
