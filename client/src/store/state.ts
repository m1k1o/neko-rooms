import {
  RoomEntry,
} from '@/api/index.ts'

export const state = {
  rooms: [] as RoomEntry[],
}

export type State = typeof state
