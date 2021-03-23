<template>
  <v-card>
    <v-card-title>
      <span class="headline"> Create new Room </span>
    </v-card-title>
    <v-card-text>
      <v-form
        ref="form"
        v-model="valid"
        lazy-validation
      >
        <v-row align="center">
          <v-col class="pb-0">
            <v-text-field
              label="Name"
              v-model="data.name"
              :rules="[ rules.slug ]"
              autocomplete="off"
            ></v-text-field>
          </v-col>
          <v-col class="pb-0">
            <v-select
              label="Neko image"
              :items="nekoImages"
              v-model="data.neko_image"
            ></v-select>
          </v-col>
        </v-row>

        <v-row align="center">
          <v-col class="py-0">
            <v-text-field
              label="User password"
              :rules="[ rules.required ]"
              v-model="data.user_pass"
              :append-icon="showUserPass ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showUserPass ? 'text' : 'password'"
              @click:append="showUserPass = !showUserPass"
              autocomplete="off"
            ></v-text-field>
          </v-col>
          <v-col class="py-0">
            <v-text-field
              label="Admin password"
              :rules="[ rules.required ]"
              v-model="data.admin_pass"
              :append-icon="showAdminPass ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showAdminPass ? 'text' : 'password'"
              @click:append="showAdminPass = !showAdminPass"
              autocomplete="off"
            ></v-text-field>
          </v-col>
        </v-row>

        <v-row align="center">
          <v-col class="pt-0">
            <v-text-field
              label="Max connections"
              type="number"
              :rules="[ rules.required, rules.nonZero, rules.onlyPositive ]"
              v-model="data.max_connections"
            ></v-text-field>
          </v-col>
          <v-col class="pt-0">
            
          </v-col>
        </v-row>

        <div class="my-3">
          <a @click="screen = !screen"> {{ screen ? 'Hide' : 'View' }} screen settings </a>
        </div>

        <v-row align="center" v-if="screen">
          <v-col>
            <v-select
              label="Initial screen configuration"
              :items="availableScreens"
              v-model="data.screen"
            ></v-select>
          </v-col>
          <v-col>
            <v-row align="center" no-gutters>
              <v-text-field
                :disabled="!maxFpsEnabled"
                label="Max frames per second"
                v-model="data.video_max_fps"
              ></v-text-field>
              <v-checkbox
                v-model="maxFpsEnabled"
                hide-details
                class="shrink ml-2 mt-0"
              ></v-checkbox>
            </v-row>
          </v-col>
        </v-row>

        <div class="my-3">
          <a @click="extended = !extended"> {{ extended ? 'Hide' : 'View' }} extended settings </a>
        </div>

        <template v-if="extended">
          <v-row align="center">
            <v-col>
              <v-select
                label="Video codec"
                :items="videoCodecs"
                v-model="data.video_codec"
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field
                label="Video bitrate"
                type="number"
                :rules="[ rules.nonZero, rules.onlyPositive ]"
                v-model="data.video_bitrate"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row align="center">
            <v-col>
              <v-select
                label="Audio codec"
                :items="audioCodecs"
                v-model="data.audio_codec"
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field
                label="Audio bitrate"
                type="number"
                :rules="[ rules.nonZero, rules.onlyPositive ]"
                v-model="data.audio_bitrate"
              ></v-text-field>
            </v-col>
          </v-row>
        </template>

        <div class="my-3">
          <a @click="expert = !expert"> {{ expert ? 'Hide' : 'View' }} expert settings </a>
        </div>

        <template v-if="expert">
          <v-row align="center" no-gutters>
            <v-checkbox
              v-model="videoPipelineEnabled"
              hide-details
              class="shrink mr-2 mt-0"
            ></v-checkbox>
            <v-textarea
              :disabled="!videoPipelineEnabled"
              label="Video pipeline"
              :rows="1"
              auto-grow
              v-model="data.video_pipeline"
            ></v-textarea>
          </v-row>
          <v-row align="center" no-gutters>
            <v-checkbox
              v-model="audioPipelineEnabled"
              hide-details
              class="shrink mr-2 mt-0"
            ></v-checkbox>
            <v-textarea
              :disabled="!audioPipelineEnabled"
              label="Audio pipeline"
              :rows="1"
              auto-grow
              v-model="data.audio_pipeline"
            ></v-textarea>
          </v-row>
          <v-row align="center" no-gutters>
            <v-checkbox
              v-model="broadcastPipelineEnabled"
              hide-details
              class="shrink mr-2 mt-0"
            ></v-checkbox>
            <v-textarea
              :disabled="!broadcastPipelineEnabled"
              label="Broadcast pipeline"
              :rows="1"
              auto-grow
              v-model="data.broadcast_pipeline"
            ></v-textarea>
          </v-row>
        </template>
      </v-form>
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
        :loading="loading"
      >
        Create
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts">
import { Vue, Component, Ref } from 'vue-property-decorator'

import {
  RoomSettings,
} from '@/api/index'

export type VForm = Vue & {
  validate: () => boolean;
  resetValidation: () => boolean;
  reset: () => void;
}

@Component
export default class RoomsCreate extends Vue {
  @Ref('form') readonly _form!: VForm

  private valid = true

  private screen = true
  private extended = false
  private expert = false

  private showUserPass = false
  private showAdminPass = false

  private maxFpsEnabled = true
  private videoPipelineEnabled = false
  private audioPipelineEnabled = false
  private broadcastPipelineEnabled = false

  private loading = false
  private data: RoomSettings = { ...this.$store.state.defaultRoomSettings }

  // eslint-disable-next-line
  private rules: any = {
    // eslint-disable-next-line
    required(val: any) {
      return val === null || typeof val === 'undefined' || val === "" ? 'This filed is mandatory.' : true
    },
    onlyPositive(val: number) {
      return val < 0 ? 'Value cannot be negative.' : true
    },
    nonZero(val: string) {
      return val === "0" ? 'Value cannot be zero.' : true
    },
    slug(val: string) {
      return val && !/^[A-Za-z0-9-_.]+$/.test(val) ? 'Should only contain A-Z a-z 0-9 - _ .' : true
    },
  }

  get nekoImages() {
    return this.$store.state.roomsConfig.neko_images
  }

  get videoCodecs() {
    return this.$store.state.videoCodecs
  }

  get audioCodecs() {
    return this.$store.state.audioCodecs
  }

  get availableScreens() {
    return this.$store.state.availableScreens
  }

  async Create() {
    const valid = this._form.validate()
    if (!valid) return

    this.loading = true

    try {
      await this.$store.dispatch('ROOMS_CREATE', {
        ...this.data,
        // eslint-disable-next-line
        max_connections: Number(this.data.max_connections),
        // eslint-disable-next-line
        video_bitrate: Number(this.data.video_bitrate),
        // eslint-disable-next-line
        video_pipeline: this.videoPipelineEnabled ? this.data.video_pipeline : '',
        // eslint-disable-next-line
        video_max_fps: this.maxFpsEnabled ? Number(this.data.video_max_fps) : 0,
        // eslint-disable-next-line
        audio_bitrate: Number(this.data.audio_bitrate),
        // eslint-disable-next-line
        audio_pipeline: this.audioPipelineEnabled ? this.data.audio_pipeline : '',
        // eslint-disable-next-line
        broadcast_pipeline: this.broadcastPipelineEnabled ? this.data.broadcast_pipeline : '',
      })
      this.Clear()
      this.$emit('finished', true)
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

  Clear() {
    this._form.resetValidation()
    this.data = {
      ...this.$store.state.defaultRoomSettings,
      // eslint-disable-next-line
      neko_image: this.nekoImages[0],
    }
  }

  Close() {
    this.Clear()
    this.$emit('finished', true)
  }

  mounted() {
    this.Clear()
  }
}
</script>
