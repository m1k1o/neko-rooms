<template>
  <div class="d-flex align-center">
    <v-text-field
      ref="input"
      :append-icon="showPass ? 'mdi-eye' : 'mdi-eye-off'"
      :value="showPass ? password : '*****'"
      @click:append="togglePass"
      @click="showPass && selectAll()"
      @focus="showPass && selectAll()"
      hide-details
      outlined
      readonly
      dense
    ></v-text-field>

    <v-spacer />

    <v-tooltip bottom v-if="room">
      <template v-slot:activator="{ on, attrs }">
        <v-btn v-bind="attrs" v-on="on" :disabled="!room.running" :href="url" target="_blank" small>
          <v-icon small>mdi-open-in-new</v-icon>
        </v-btn>
      </template>
      <span>{{ label }}</span>
    </v-tooltip>

    <v-spacer />

    <v-tooltip bottom v-if="room">
      <template v-slot:activator="{ on, attrs }">
        <v-btn v-bind="attrs" v-on="on" :disabled="!room.running" @click="copyToClipboard" small>
          <v-icon small v-if="copied">mdi-clipboard-check-multiple</v-icon>
          <v-icon small v-else>mdi-clipboard-multiple-outline</v-icon>
        </v-btn>
      </template>
      <span>copy link to clipboard</span>
    </v-tooltip>
  </div>
</template>

<script lang="ts">
import { Vue, Component, Prop } from 'vue-property-decorator'

import {
  RoomEntry,
} from '@/api/index'

@Component
export default class RoomLink extends Vue {
  @Prop(String) readonly roomId!: string
  @Prop(String) readonly password!: string
  @Prop(String) readonly label!: string

  private showPass = false
  private copied = false
  private copiedTimeout = 0

  get room(): RoomEntry {
    return this.$store.state.rooms.find(({ id }: RoomEntry) => id == this.roomId)
  }

  get url() {
    return this.room.url + '?pwd=' + encodeURIComponent(this.password)
  }
  
  togglePass() {
    this.showPass = !this.showPass
    if (this.showPass) {
      setTimeout(() => {
        this.selectAll()
      }, 0)
    }
  }

  selectAll() {
    this.$refs.input.$el.querySelector('input').select();
  }

  copyToClipboard() {
    if (this.copiedTimeout) {
      clearInterval(this.copiedTimeout)
    }

    navigator.clipboard.writeText(this.url)
    this.copied = true

    this.copiedTimeout = setTimeout(() => {
      this.copied = false
      this.copiedTimeout = 0
    }, 3000)
  }
}
</script>
