import {
  RoomEntry,
} from '@/api/index'

export const state = {
  rooms: [] as RoomEntry[],
}

export type State = typeof state
