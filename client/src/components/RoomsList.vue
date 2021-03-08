<template>
  <v-data-table
    :headers="headers"
    :items="rooms"
    class="elevation-1"
    :loading="loading"
    loading-text="Loading... Please wait"
    hide-default-footer
  >
    <template v-slot:[`item.url`]="{ item }">
      <a :href="item.url"> Link </a>
    </template>
    <template v-slot:[`item.max_connections`]="{ item }">
      <span v-if="item.max_connections">{{ item.max_connections }}</span>
      <i v-else>--not-specified--</i>
    </template>
    <template v-slot:[`item.running`]="{ item }">
      <v-chip color="green" dark small v-if="item.running">yes</v-chip>
      <v-chip color="red" dark small v-else>no</v-chip>
    </template>
    <template v-slot:[`item.created`]="{ item }">
      {{ item.created | timeago }}
    </template>
    <template v-slot:[`item.actions`]="{ item }">
      <v-btn @click="() => RoomStart(item.id)" class="mb-3" color="green" icon :disabled="item.running"><v-icon>mdi-play-circle-outline</v-icon></v-btn>
      <v-btn @click="() => RoomStop(item.id)" class="mb-3" color="warning" icon :disabled="!item.running"><v-icon>mdi-stop-circle-outline</v-icon></v-btn>
      <v-btn @click="() => RoomRestart(item.id)" class="mb-3" color="blue" icon :disabled="!item.running"><v-icon>mdi-refresh</v-icon></v-btn>
    </template>
    <template v-slot:[`item.destroy`]="{ item }">
      <v-btn @click="() => RoomRemove(item.id)" class="mb-3" color="red" icon :loading="deleteLoading"><v-icon>mdi-trash-can-outline</v-icon></v-btn>
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { Vue, Component, Prop } from 'vue-property-decorator'

@Component
export default class RoomsList extends Vue {
  @Prop(Boolean) readonly loading: boolean = false

  private deleteLoading = false
  private headers = [
    { text: 'Name', value: 'name' },
    { text: 'Link', value: 'url' },
    { text: 'Max connections', value: 'max_connections' },
    { text: 'Image', value: 'image' },
    { text: 'Running', value: 'running' },
    { text: 'Status', value: 'status' },
    { text: 'Created', value: 'created' },
    {
      text: 'Actions',
      value: 'actions',
      align: 'end',
      sortable: false,
    },
    {
      text: 'Destroy',
      value: 'destroy',
      align: 'end',
      sortable: false,
    },
  ]

  get rooms() {
    return this.$store.state.rooms
  }

  async RoomStart(roomId: string) {
    try {
      await this.$store.dispatch('ROOMS_START', roomId)
      await this.$swal({
        title: 'Room started!',
        icon: 'success',
      })
    } catch(e) {
      this.handeErrors(e)
    }
  }

  async RoomStop(roomId: string) {
    try {
      await this.$store.dispatch('ROOMS_STOP', roomId)
      await this.$swal({
        title: 'Room stopped!',
        icon: 'success',
      })
    } catch(e) {
      this.handeErrors(e)
    }
  }

  async RoomRestart(roomId: string) {
    try {
      await this.$store.dispatch('ROOMS_RESTART', roomId)
      await this.$swal({
        title: 'Room restarted!',
        icon: 'success',
      })
    } catch(e) {
      this.handeErrors(e)
    }
  }

  async RoomRemove(roomId: string) {
    const { value } = await this.$swal({
      title: "Remove room",
      text: "Do you really want to remove this room?",
      icon: 'warning',
      showCancelButton: true,
      confirmButtonText: "Yes",
      cancelButtonText: "No",
    })

    if (!value) {
      return
    }

    this.deleteLoading = true

    try {
      await this.$store.dispatch('ROOMS_REMOVE', roomId)
      await this.$swal({
        title: 'Room removed!',
        icon: 'success',
      })
    } catch(e) {
      this.handeErrors(e)
    } finally {
      this.deleteLoading = false
    }
  }

  // eslint-disable-next-line
  handeErrors(e: any) {
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
  }
}
</script>
