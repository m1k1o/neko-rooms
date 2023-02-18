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
              :rules="[ rules.minLen(2), rules.containerNameStart, rules.containerName ]"
              autocomplete="off"
              :hint="!data.name && '... using random name'"
              persistent-hint
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
          <v-col>
            <v-text-field
              label="User password"
              v-model="data.user_pass"
              :append-icon="showUserPass ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showUserPass ? 'text' : 'password'"
              @click:append="showUserPass = !showUserPass"
              autocomplete="off"
              :hint="!data.user_pass && '... using random password'"
              persistent-hint
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              label="Admin password"
              v-model="data.admin_pass"
              :append-icon="showAdminPass ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showAdminPass ? 'text' : 'password'"
              @click:append="showAdminPass = !showAdminPass"
              autocomplete="off"
              :hint="!data.admin_pass && '... using random password'"
              persistent-hint
            ></v-text-field>
          </v-col>
        </v-row>

        <!-- if uses mux vs. old logic -->
        <v-row align="center" v-if="usesMux">
          <v-col class="pt-0">
            <v-checkbox
              v-model="data.control_protection"
              label="Enable control protection"
              hide-details
              class="shrink ml-2 mt-0"
            ></v-checkbox>
            <div style="margin-left: 41px;"><i>Users can gain control only if at least one admin is in the room.</i></div>
          </v-col>
          <v-col class="pt-0">
            <v-checkbox
              v-model="data.implicit_control"
              label="Enable implicit control"
              hide-details
              class="shrink ml-2 mt-0"
            ></v-checkbox>
            <div style="margin-left: 41px;"><i>Users do not need to request control prior usage.</i></div>
          </v-col>
        </v-row>
        <v-row align="center" v-else>
          <v-col class="pt-0">
            <v-text-field
              label="Max connections"
              type="number"
              :rules="[ rules.required, rules.nonZero, rules.onlyPositive ]"
              v-model="data.max_connections"
            ></v-text-field>
          </v-col>
          <v-col class="pt-0">
            <v-checkbox
              v-model="data.control_protection"
              label="Enable control protection"
              hide-details
              class="shrink ml-2 mt-0"
            ></v-checkbox>
            <div style="margin-left: 41px;"><i>Users can gain control only if at least one admin is in the room.</i></div>
            <v-checkbox
              v-model="data.implicit_control"
              label="Enable implicit control"
              hide-details
              class="shrink ml-2 mt-0"
            ></v-checkbox>
            <div style="margin-left: 41px;"><i>Users do not need to request control prior usage.</i></div>
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
                :rules="[ rules.required, rules.nonZero, rules.onlyPositive ]"
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
                :rules="[ rules.required, rules.nonZero, rules.onlyPositive ]"
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
                :rules="[ rules.required, rules.nonZero, rules.onlyPositive ]"
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
          <v-row align="center" no-gutters>
            <h2> Environment variables </h2>
            <v-btn @click="addEnv" icon color="green"><v-icon>mdi-plus</v-icon></v-btn>
          </v-row>
          <v-row align="center" v-for="({ key, val }, index) in envList" :key="index">
            <v-col class="py-0">
              <v-text-field
                label="Key"
                :value="key"
                @input="setEnv(index, { key: $event, val })"
                autocomplete="off"
              ></v-text-field>
            </v-col>
            <div> = </div>
            <v-col class="py-0">
              <v-text-field
                label="Value"
                :value="val"
                @input="setEnv(index, { key, val: $event })"
                autocomplete="off"
              ></v-text-field>
            </v-col>
            <div>
              <v-btn @click="delEnv(index)" icon color="red"><v-icon>mdi-close</v-icon></v-btn>
            </div>
          </v-row>
          <v-row align="center" no-gutters class="my-3">
              <h2> Mounts </h2>
              <v-btn :disabled="!storageEnabled" @click="data.mounts = [ ...data.mounts, { type: 'private', host_path: '', container_path: '' }]" icon color="green"><v-icon>mdi-plus</v-icon></v-btn>
          </v-row>
          <v-alert
            border="left"
            type="warning"
            v-if="!storageEnabled"
          >
            <p><strong>Not available!</strong></p>
            <p class="mb-0">Mounts are not available, because storage is not enabled.</p>
          </v-alert>
          <v-row align="center" class="mb-2" v-for="({ type, host_path, container_path }, index) in data.mounts" :key="index">
            <v-col class="py-0" cols="2">
              <v-select
                label="Type"
                :items="mountTypes"
                :value="type"
                @input="$set(data.mounts, index, { type: $event, host_path, container_path })"
              ></v-select>
            </v-col>
            <v-col class="py-0 pl-0">
              <v-text-field
                label="Host path"
                :value="host_path"
                @input="$set(data.mounts, index, { type, host_path: $event, container_path })"
                :rules="[ rules.absolutePath ]"
                autocomplete="off"
              ></v-text-field>
            </v-col>
            <div> : </div>
            <v-col class="py-0">
              <v-text-field
                label="Container path"
                :value="container_path"
                @input="$set(data.mounts, index, { type, host_path, container_path: $event})"
                :rules="[ rules.absolutePath ]"
                autocomplete="off"
              ></v-text-field>
            </v-col>
            <div>
              <v-btn @click="$delete(data.mounts, index)" icon color="red"><v-icon>mdi-close</v-icon></v-btn>
            </div>
          </v-row>
          <v-row align="center" no-gutters v-if="data.mounts.length > 0">
            <p>
              <strong>Private</strong>: Host path is relative to <code class="mx-1">&lt;storage path&gt;/rooms/&lt;room name&gt;/</code>. <br />
              <strong>Template</strong>: Host path is relative to <code class="mx-1">&lt;storage path&gt;/templates/</code>, and will be readonly. <br />
              <strong>Protected</strong>: Host path must be whitelisted in config and exists on the host, will be readonly. <br />
              <strong>Public</strong>: Host path must be whitelisted in config and exists on the host.
            </p>
          </v-row>
          <v-row align="center" no-gutters>
            <h2> Resources </h2>
          </v-row>
          <v-row align="center">
            <v-col>
              <v-slider
                v-model="data.resources.memory"
                label="Memory"
                thumb-label="always"
                :thumb-size="30"
                :min="0"
                :max="8*1e9"
                :step="0.2*1e9"
                color="blue"
                hide-details
              >
                <template v-slot:thumb-label="{ value }">
                  <span v-if="value">{{ value | memory }}</span>
                  <span v-else>N/A</span>
                </template>
              </v-slider>
            </v-col>
            <v-col>
              <v-slider
                v-model="data.resources.nano_cpus"
                label="CPUs"
                thumb-label="always"
                :thumb-size="30"
                :min="0"
                :max="8*1e9"
                :step="0.2*1e9"
                color="blue"
                hide-details
              >
                <template v-slot:thumb-label="{ value }">
                  <span v-if="value">{{ value | nanocpus }}</span>
                  <span v-else>N/A</span>
                </template>
              </v-slider>
            </v-col>
          </v-row>
          <v-row align="center">
            <v-col>
              <v-slider
                v-model="data.resources.shm_size"
                label="Shared memory"
                thumb-label="always"
                :thumb-size="30"
                :min="0"
                :max="20*1e9"
                :step="0.2*1e9"
                color="blue"
              >
                <template v-slot:thumb-label="{ value }">
                  <span v-if="value">{{ value | memory }}</span>
                  <span v-else>N/A</span>
                </template>
              </v-slider>
            </v-col>
            <v-col>
              <v-checkbox
                @change="$set(data.resources, 'gpus', $event ? ['all'] : [])"
                label="Enable GPU support"
                class="shrink ml-2 mt-0"
              ></v-checkbox>
            </v-col>
          </v-row>
        </template>

        <hr />

        <v-row align="center" no-gutters class="mt-3">
          <h2 class="my-3">
            Browser policy
          </h2>
          <v-checkbox
            v-model="browserPolicyEnabled"
            :disabled="!browserPolicyConfig || !storageEnabled"
            hide-details
            class="shrink ml-2 mt-0"
          ></v-checkbox>
        </v-row>

        <v-alert
          border="left"
          type="warning"
          v-if="!storageEnabled"
        >
          <p><strong>Not available!</strong></p>
          <p class="mb-0">Browser policy is not available, because storage is not enabled.</p>
        </v-alert>
        <template v-else-if="browserPolicyConfig">
          <v-row align="center" no-gutters class="mt-0">
            <v-col>
              <v-select
                v-model="browserPolicyContent.extensions"
                label="Extensions"
                :items="browserPolicyExtensions"
                multiple
                :disabled="!browserPolicyEnabled"
              ></v-select>
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col>
              <v-checkbox
                v-model="browserPolicyContent.developer_tools"
                label="Enable developer tools"
                hide-details
                class="shrink ml-2 mt-0"
                :disabled="!browserPolicyEnabled"
              ></v-checkbox>
            </v-col>
            <v-col>
              <v-checkbox
                v-model="browserPolicyContent.persistent_data"
                label="Allow persistent data"
                hide-details
                class="shrink ml-2 mt-0"
                :disabled="!browserPolicyEnabled"
              ></v-checkbox>
            </v-col>
          </v-row>
        </template>
        <v-alert
          border="left"
          type="info"
          v-else
        >
          <p><strong>Not available!</strong></p>
          <p class="mb-0">Browser policy is not available for this image.</p>
        </v-alert>
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
import { Vue, Component, Ref, Watch } from 'vue-property-decorator'
import { randomPassword } from '@/utils/random'

import {
  RoomSettings,
  BrowserPolicyContent,
  BrowserPolicyExtension,
  BrowserPolicyTypeEnum,
  RoomMountTypeEnum,
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
  private browserPolicyEnabled = false

  private loading = false
  private data: RoomSettings = { ...this.$store.state.defaultRoomSettings }
  private browserPolicyContent: BrowserPolicyContent = { ...this.$store.state.defaultBrowserPolicyContent }
  private envList: { key: string; val: string }[] = []

  // eslint-disable-next-line
  private rules: any = {
    // eslint-disable-next-line
    required(val: any) {
      return val === null || typeof val === 'undefined' || val === "" ? 'This filed is mandatory.' : true
    },
    minLen: (min: number) =>
      (val: string) => 
        val ? (val.length >= min || 'This field must have atleast ' + min + ' characters') : true,
    onlyPositive(val: number) {
      return val < 0 ? 'Value cannot be negative.' : true
    },
    nonZero(val: string) {
      return val === "0" ? 'Value cannot be zero.' : true
    },
    containerName(val: string) {
      return val && !/^[a-zA-Z0-9_.-]+$/.test(val) ? 'Must only contain a-z A-Z 0-9 _ . -' : true
    },
    containerNameStart(val: string) {
      return val && /^[_.-]/.test(val) ? 'Cannot start with _ . -' : true
    },
    absolutePath(val: string) {
      return val[0] !== "/" ? 'Must be absolute path, starting with /.' : true
    }
  }

  get nekoImages() {
    return this.$store.state.roomsConfig.neko_images
  }

  get storageEnabled() {
    return this.$store.state.roomsConfig.storage_enabled
  }

  get usesMux() {
    return this.$store.state.roomsConfig.uses_mux
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

  get mountTypes() {
    return [
      {
        text: 'Private',
        value: 'private',
      },
      {
        text: 'Template',
        value: 'template',
      },
      {
        text: 'Protected',
        value: 'protected',
      },
      {
        text: 'Public',
        value: 'public',
      },
    ]
  }

  get browserPolicyConfig() {
    const nekoImage = this.data.neko_image
    if (!nekoImage) return undefined
  
    return this.$store.state.browserPolicyConfig.find(({ images }: { images: string[] }): boolean => images.includes(nekoImage))
  }

  get browserPolicyExtensions() {
    const config = this.browserPolicyConfig
    if (!config) return []
  
    return this.$store.state.browserPolicyExtensions.map(({ text, value: values }: {
      text: string;
      value: Record<BrowserPolicyTypeEnum, BrowserPolicyExtension>;
    }) => {
      const value = values[config.type as BrowserPolicyTypeEnum]
      if (!value) return undefined
      return { text, value }
    })
  }

  @Watch('browserPolicyContent.persistent_data')
  onPersistentDataUpdate(enabled: boolean) {
    const config = this.browserPolicyConfig
    if (!config) return

    if (enabled) {
      // eslint-disable-next-line
      this.data.mounts = [ ...(this.data.mounts || []), { type: RoomMountTypeEnum.private, host_path: '/profile', container_path: config.profile }]
    } else {
      // eslint-disable-next-line
      this.data.mounts = (this.data.mounts || []).filter(({ type, container_path }) => type != RoomMountTypeEnum.private && container_path != config.profile)
    }
  }

  addEnv() {
    Vue.set(this, 'envList', [ ...this.envList, { key: '', val: '' } ])
  }

  setEnv(index: number, data: { key: string; val: string }) {
    Vue.set(this.envList, index, data)
  }

  delEnv(index: number) {
    Vue.delete(this.envList, index)
  }

  async Create() {
    const valid = this._form.validate()
    if (!valid) return

    this.loading = true

    try {
      const envs = this.envList.reduce((obj, { key, val }) => ({ ...obj, [key]: val, }), {})

      await this.$store.dispatch('ROOMS_CREATE', {
        ...this.data,
        // eslint-disable-next-line
        user_pass: this.data.user_pass || randomPassword(),
        // eslint-disable-next-line
        admin_pass: this.data.admin_pass || randomPassword(),
        // eslint-disable-next-line
        max_connections: Number(this.data.max_connections), // ignored when uses mux
        // eslint-disable-next-line
        control_protection: Boolean(this.data.control_protection),
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
        envs,
        // eslint-disable-next-line
        browser_policy: this.browserPolicyEnabled && this.browserPolicyConfig ? {
          type: this.browserPolicyConfig.type,
          path: this.browserPolicyConfig.path,
          content: this.browserPolicyContent
        } : undefined,
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
    this.browserPolicyContent = { ...this.$store.state.defaultBrowserPolicyContent }
    this.envList = Object.entries({...this.data.envs}).map(([ key, val ]) => ({ key, val, }))
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
