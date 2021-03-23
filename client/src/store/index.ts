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
} from '@/api/index'

import { state, State } from './state'

Vue.use(Vuex)

const configuration = new Configuration({
  basePath: (location.protocol + '//' + location.host + location.pathname).replace(/\/+$/, ''),
})

const configApi = new ConfigApi(configuration)
const roomsApi = new RoomsApi(configuration)

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
      Vue.set(state, 'rooms', [roomEntry, ...state.rooms])
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
    async ROOMS_CREATE({ commit }: ActionContext<State, State>, roomSettings: RoomSettings): Promise<RoomEntry>  {
      const res = await roomsApi.roomCreate(roomSettings)
      commit('ROOMS_ADD', res.data);
      return res.data
    },
    async ROOMS_GET({ commit }: ActionContext<State, State>, roomId: string)  {
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
        status: 'Up',
      });
    },
    async ROOMS_STOP({ commit }: ActionContext<State, State>, roomId: string) {
      await roomsApi.roomStop(roomId)
      commit('ROOMS_PUT', {
        id: roomId,
        running: false,
        status: 'Exited',
      });
    },
    async ROOMS_RESTART(_: ActionContext<State, State>, roomId: string) {
      await roomsApi.roomRestart(roomId)
    },
  },
  modules: {
  }
})
