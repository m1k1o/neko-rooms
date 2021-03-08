<template>
  <v-container>
    <v-row>
      <v-col>
        <v-btn @click="LoadRooms" class="mb-3" color="green" icon><v-icon>mdi-refresh</v-icon></v-btn>
      </v-col>
      <v-col class="text-right">
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
              + Add new
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
import RoomsCreate from '@/components/RoomsCreate.vue'

@Component({
  components: {
    RoomsList,
    RoomsCreate,
  }
})
export default class Home extends Vue {
  private dialog = false
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
