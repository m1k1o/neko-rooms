<template>
  <v-data-table
    :headers="headers"
    :items="rooms"
    class="elevation-1"
  >
    <template v-slot:[`item.url`]="{ item }">
      <a :href="item.url"> Link </a>
    </template>
    <template v-slot:[`item.running`]="{ item }">
      <v-chip color="green" dark small v-if="item.running">yes</v-chip>
      <v-chip color="red" dark small v-else>no</v-chip>
    </template>
    <template v-slot:[`item.created`]="{ item }">
      {{ item.created | timeago }}
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'

@Component
export default class RoomsList extends Vue {
  private headers = [
    { text: 'Name', value: 'name' },
    { text: 'Link', value: 'url' },
    { text: 'Max connections', value: 'max_connections' },
    { text: 'Image', value: 'image' },
    { text: 'Running', value: 'running' },
    { text: 'Status', value: 'status' },
    { text: 'Created', value: 'created' },
  ]

  get rooms() {
    return this.$store.state.rooms
  }
}
</script>
