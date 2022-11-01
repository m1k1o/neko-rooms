<template>
  <span>
    <v-btn
      @click="dialog = true"
      :loading="loading"
      color="info"
      dark
    >
      <v-icon class="mr-2" color="white">{{ status.active ? 'mdi-cloud-sync' : 'mdi-cloud-download-outline' }}</v-icon> Pull neko images
    </v-btn>

    <v-dialog v-model="dialog" max-width="780px">
      <v-card>
        <v-card-title class="headline">
          Pull neko images
        </v-card-title>
        <v-card-text>
          <template v-if="status.active && status.layers && status.layers.length > 0">
            <pre v-for="layer in status.layers" :key="layer.id">{{ layer.id }} {{ layer.status }}{{ layer.progress && ' ' + layer.progress }}</pre>
            <br />
          </template>
          <template v-if="status.status && status.status.length > 0">
            <pre v-for="text in status.status" :key="text">{{ text }}</pre>
          </template>
          <pre v-else-if="status.active">Preparing docker image pull</pre>
        </v-card-text>
        <v-card-actions>
          <template v-if="!status.active">
            <v-select
              v-model="nekoImage"
              :items="nekoImages"
              dense
              outlined
              hide-details
              label="Neko image"
            ></v-select>
            <v-btn color="green" class="ml-2" :loading="loading" @click="Start">
              Start
            </v-btn>
          </template>
          <v-spacer v-else></v-spacer>
          <v-btn v-if="status.active" color="red" text :loading="loading" @click="Stop">
            Stop
          </v-btn>
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

@Component
export default class Pull extends Vue {
  private dialog = false
  private loading = false
  private nekoImage = ''

  get status() {
    return this.$store.state.pullStatus
  }

  get nekoImages() {
    return this.$store.state.roomsConfig.neko_images
  }

  async Start() {
    this.loading = true
  
    try {
      await this.$store.dispatch('PULL_START', this.nekoImage)
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

  async Stop() {
    this.loading = true
  
    try {
      await this.$store.dispatch('PULL_STOP')
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
      this.dialog = false
    }
  }

  private interval!: number

  async mounted() {
    await this.$store.dispatch('PULL_STATUS')

    this.interval = window.setInterval(async () => {
      if (!this.status.active) return

      try {
        await this.$store.dispatch('PULL_STATUS')
      } catch(e) {
        console.error(e)
      }
    }, 1000)
  }

  beforeDestroy() {
    window.clearInterval(this.interval)
  }
}
</script>
