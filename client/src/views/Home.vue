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
      <v-col class="mb-3">
        <v-select
          v-model="autoRefresh"
          :items="autoRefreshItems"
          dense
          outlined
          hide-details
          label="Auto refresh"
          class="mr-2"
          style="width:100px;display:inline-block"
        ></v-select>
        <v-tooltip right open-delay="300">
          <template v-slot:activator="{ on, attrs }">
            <v-btn v-bind="attrs" v-on="on" @click="LoadRooms" color="green" icon><v-icon>mdi-reload</v-icon></v-btn>
          </template>
          <span>Manual refresh</span>
        </v-tooltip>
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

    <div class="mt-5 text-center">
      <Pull />
    </div>
  </v-container>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import RoomsList from '@/components/RoomsList.vue'
import RoomsQuick from '@/components/RoomsQuick.vue'
import RoomsCreate from '@/components/RoomsCreate.vue'
import Pull from '@/components/Pull.vue'
import { RoomEntry } from '@/api/index'

@Component({
  components: {
    RoomsList,
    RoomsQuick,
    RoomsCreate,
    Pull,
  }
})
export default class Home extends Vue {
  private loading = false
  private dialog = false

  private interval!: number
  private autoRefresh = 10
  private autoRefreshItems = [
    { text: 'Off', value: 0 },
    { text: '5s', value: 5 },
    { text: '10s', value: 10 },
    { text: '30s', value: 30 },
    { text: '60s', value: 60 },
  ]

  get configConnections() {
    return this.$store.state.roomsConfig.connections
  }

  get usedConnections() {
    // eslint-disable-next-line
    return this.$store.state.rooms.reduce((sum: number, { max_connections }: RoomEntry) => sum + (max_connections || 1/* 0 when using mux */), 0)
  }

  async LoadRooms() {
    // do not load config if document is hidden
    if (document.hidden) return

    this.loading = true

    try {
      await this.$store.dispatch('ROOMS_LOAD')
    } finally {
      this.loading = false
    }
  }


  @Watch('autoRefresh', { immediate: true })
  onAutoRefresh() {
    if (this.interval) {
      clearInterval(this.interval)
    }
  
    if (this.autoRefresh) {
      this.interval = setInterval(this.LoadRooms, this.autoRefresh * 1000)
    }
  }

  mounted() {
    this.$store.dispatch('ROOMS_CONFIG')
    this.LoadRooms()
  }

  beforeDestroy() {
    this.autoRefresh = 0
  }
}
</script>
