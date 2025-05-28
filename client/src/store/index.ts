import Vue from 'vue'
import Vuex, { ActionContext } from 'vuex'

import {
  Configuration,
  RoomsConfig,
  ConfigApi,
  RoomEntry,
  RoomSettings,
  RoomStats,
  RoomsApi,
  DefaultApi,
  PullStatus,
} from '@/api/index'

import { state, State } from './state'

Vue.use(Vuex)

const configuration = new Configuration({
  basePath: (location.protocol + '//' + location.host + location.pathname).replace(/\/+$/, ''),
})

const configApi = new ConfigApi(configuration)
const roomsApi = new RoomsApi(configuration)
const defaultApi = new DefaultApi(configuration)

export default new Vuex.Store({
  state,
  mutations: {
    ROOMS_CONFIG_SET(state: State, roomsConfig: RoomsConfig) {
      Vue.set(state, 'roomsConfig', roomsConfig)
    },
    ROOMS_SET(state: State, roomEntries: RoomEntry[]) {
      Vue.set(state, 'rooms', roomEntries)
    },
    ROOMS_ADD(state: State, roomEntry: RoomEntry) {
      // check if room already exists
      if (state.rooms.some(({ id }) => id == roomEntry.id)) {
        // replace room
        Vue.set(state, 'rooms', state.rooms.map((room) => {
          if (room.id == roomEntry.id) {
            return roomEntry
          } else {
            return room
          }
        }))
      } else {
        // add room
        Vue.set(state, 'rooms', [roomEntry, ...state.rooms])
      }
    },
    ROOMS_PUT(state: State, roomEntry: RoomEntry) {
      let exists = false
      const roomEntries = state.rooms.map((room) => {
        if (room.id == roomEntry.id) {
          exists = true
          return { ...room, ...roomEntry }
        } else {
          return room
        }
      })

      if (exists) {
        Vue.set(state, 'rooms', roomEntries)
      } else {
        Vue.set(state, 'rooms', [roomEntry, ...roomEntries])
      }
    },
    ROOMS_DEL(state: State, roomId: string) {
      const roomEntries = state.rooms.filter(({ id }) => id != roomId)
      Vue.set(state, 'rooms', roomEntries)
    },
    PULL_STATUS(state: State, pullStatus: PullStatus) {
      Vue.set(state, 'pullStatus', pullStatus)
    },
  },
  actions: {
    async ROOMS_CONFIG({ commit }: ActionContext<State, State>) {
      const res = await configApi.roomsConfig()
      commit('ROOMS_CONFIG_SET', res.data);
    },
    async ROOMS_LOAD({ commit }: ActionContext<State, State>) {
      const res = await roomsApi.roomsList()
      commit('ROOMS_SET', res.data);
    },
    async ROOMS_CREATE({ commit }: ActionContext<State, State>, roomSettings: RoomSettings): Promise<RoomEntry> {
      const res = await roomsApi.roomCreate(roomSettings, false)
      commit('ROOMS_ADD', res.data);
      return res.data
    },
    async ROOMS_CREATE_AND_START({ commit }: ActionContext<State, State>, roomSettings: RoomSettings): Promise<RoomEntry> {
      const res = await roomsApi.roomCreate(roomSettings, true)
      commit('ROOMS_ADD', res.data);
      return res.data
    },
    async ROOMS_GET({ commit }: ActionContext<State, State>, roomId: string) {
      const res = await roomsApi.roomGet(roomId)
      commit('ROOMS_PUT', res.data);
      return res.data
    },
    async ROOMS_REMOVE({ commit }: ActionContext<State, State>, roomId: string) {
      await roomsApi.roomRemove(roomId)
      commit('ROOMS_DEL', roomId);
    },
    async ROOMS_SETTINGS(_: ActionContext<State, State>, roomId: string): Promise<RoomSettings> {
      const res = await roomsApi.roomSettings(roomId)
      return res.data
    },
    async ROOMS_STATS(_: ActionContext<State, State>, roomId: string): Promise<RoomStats> {
      const res = await roomsApi.roomStats(roomId)
      return res.data
    },
    async ROOMS_START({ commit }: ActionContext<State, State>, roomId: string) {
      await roomsApi.roomStart(roomId)
      commit('ROOMS_PUT', {
        id: roomId,
        running: true,
        paused: false,
        status: 'Up',
      });
    },
    async ROOMS_STOP({ commit }: ActionContext<State, State>, roomId: string) {
      await roomsApi.roomStop(roomId)
      commit('ROOMS_PUT', {
        id: roomId,
        running: false,
        paused: false,
        status: 'Exited',
      });
    },
    async ROOMS_PAUSE({ commit }: ActionContext<State, State>, roomId: string) {
      await roomsApi.roomPause(roomId)
      commit('ROOMS_PUT', {
        id: roomId,
        running: false,
        paused: true,
        status: 'Paused',
      });
    },
    async ROOMS_RESTART(_: ActionContext<State, State>, roomId: string) {
      await roomsApi.roomRestart(roomId)
    },
    async ROOMS_RECREATE({ commit }: ActionContext<State, State>, roomId: string) {
      const res = await roomsApi.roomRecreate(roomId, {} as RoomSettings)
      commit('ROOMS_DEL', roomId)
      commit('ROOMS_PUT', res.data)
      return res.data
    },

    async PULL_START({ commit }: ActionContext<State, State>, nekoImage: string) {
      const res = await defaultApi.pullStart({
        // eslint-disable-next-line
        neko_image: nekoImage,
      })
      commit('PULL_STATUS', res.data)
      return res.data
    },
    async PULL_STATUS({ commit }: ActionContext<State, State>) {
      const res = await defaultApi.pullStatus()
      commit('PULL_STATUS', res.data)
    },
    async PULL_STOP() {
      const res = await defaultApi.pullStop()
      return res.data
    },

    async EVENTS_SSE(): Promise<EventSource> {
      return new EventSource(configuration.basePath + '/api/events?sse', {
        withCredentials: true,
      })
    },
  },
  modules: {
  }
})
