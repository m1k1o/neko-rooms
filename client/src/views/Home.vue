<template>
  <v-container>
    <v-btn @click="LoadRooms" class="mb-3" color="green" icon><v-icon>mdi-refresh</v-icon></v-btn>

    <RoomsList :loading="loading" />
  </v-container>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import RoomsList from '@/components/RoomsList.vue'

@Component({
  components: {
    RoomsList
  }
})
export default class Home extends Vue {
  private loading = false

  async LoadRooms() {
    this.loading = true

    try {
      await this.$store.dispatch('ROOMS_LOAD')
    } finally {
      this.loading = false
    }
  }
}
</script>
