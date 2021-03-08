<template>
  <div class="home">
    <RoomsList />
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import RoomsList from '@/components/RoomsList.vue'

@Component({
  components: {
    RoomsList
  }
})
export default class Home extends Vue {}
</script>
