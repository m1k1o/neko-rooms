<template>
  <v-card>
    <v-card-title>
      <span class="headline"> Room Create </span>
    </v-card-title>
    <v-card-text>
      <v-text-field
        label="Name"
        v-model="data.name"
      ></v-text-field>
      <v-text-field
        label="max_connections"
        v-model="data.max_connections"
      ></v-text-field>
      <v-text-field
        label="user_pass"
        v-model="data.user_pass"
      ></v-text-field>
      <v-text-field
        label="admin_pass"
        v-model="data.admin_pass"
      ></v-text-field>
      <v-text-field
        label="broadcast_pipeline"
        v-model="data.broadcast_pipeline"
      ></v-text-field>
      <v-text-field
        label="screen"
        v-model="data.screen"
      ></v-text-field>
      <v-text-field
        label="video_codec"
        v-model="data.video_codec"
      ></v-text-field>
      <v-text-field
        label="video_bitrate"
        v-model="data.video_bitrate"
      ></v-text-field>
      <v-text-field
        label="video_pipeline"
        v-model="data.video_pipeline"
      ></v-text-field>
      <v-text-field
        label="video_max_fps"
        v-model="data.video_max_fps"
      ></v-text-field>
      <v-text-field
        label="audio_codec"
        v-model="data.audio_codec"
      ></v-text-field>
      <v-text-field
        label="audio_bitrate"
        v-model="data.audio_bitrate"
      ></v-text-field>
      <v-text-field
        label="audio_pipeline"
        v-model="data.audio_pipeline"
      ></v-text-field>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        color="gray darken-1"
        text
        @click="Close"
      >
        Close
      </v-btn>
      <v-btn
        color="green"
        dark
        @click="Create"
      >
        Create
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'

import {
  RoomSettings,
} from '@/api/index'

@Component
export default class RoomsCreate extends Vue {
  private loading = false
  private data: RoomSettings = {}

  async Create() {
    this.loading = true

    try {
      await this.$store.dispatch('ROOMS_CREATE', this.data)
    } finally {
      this.loading = false
      this.$emit('finished', true)
    }
  }

  Close() {
    this.data = {}
    this.$emit('finished', true)
  }
}
</script>
