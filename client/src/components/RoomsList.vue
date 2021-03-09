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
      <v-btn :disabled="!item.running" :href="item.url" target="_blank" small> <v-icon small>mdi-open-in-new</v-icon></v-btn>
    </template>
    <template v-slot:[`item.max_connections`]="{ item }">
      <span v-if="item.max_connections">{{ item.max_connections }}</span>
      <i v-else>--not-specified--</i>
    </template>
    <template v-slot:[`item.status`]="{ item }">
      <v-chip :color="item.running ? 'green' : 'red'" dark small> {{ item.status }} </v-chip>
    </template>
    <template v-slot:[`item.created`]="{ item }">
      {{ item.created | timeago }}
    </template>
    <template v-slot:[`item.actions`]="{ item }">
      <RoomActionBtn action="start" :roomId="item.id" :disabled="item.running" />
      <RoomActionBtn action="stop" :roomId="item.id" :disabled="!item.running" />
      <RoomActionBtn action="restart" :roomId="item.id" :disabled="!item.running" />
    </template>
    <template v-slot:[`item.destroy`]="{ item }">
      <RoomActionBtn action="remove" :roomId="item.id" :disabled="!item.running" />
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { Vue, Component, Prop } from 'vue-property-decorator'
import RoomActionBtn from '@/components/RoomActionBtn.vue'

@Component({
  components: {
    RoomActionBtn,
  }
})
export default class RoomsList extends Vue {
  @Prop(Boolean) readonly loading: boolean = false

  private headers = [
    { text: 'Address', value: 'url' },
    { text: 'Name', value: 'name' },
    { text: 'Max connections', value: 'max_connections' },
    { text: 'Image', value: 'image' },
    { text: 'Status', value: 'status' },
    { text: 'Created', value: 'created' },
    {
      text: 'Actions',
      value: 'actions',
      sortable: false,
    },
    {
      text: 'Destroy',
      value: 'destroy',
      sortable: false,
    },
  ]

  get rooms() {
    return this.$store.state.rooms
  }
}
</script>
