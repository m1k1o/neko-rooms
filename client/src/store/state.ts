import {
  RoomsConfig,
  RoomEntry,
  RoomSettings,
} from '@/api/index'

export const state = {
  roomsConfig: {} as RoomsConfig,
  rooms: [] as RoomEntry[],
  defaultRoomSettings: {
    name: '',
    // eslint-disable-next-line
    neko_image: '',
    // eslint-disable-next-line
    max_connections: 10,
    // eslint-disable-next-line
    user_pass: '',
    // eslint-disable-next-line
    admin_pass: '',
  
    screen: '1280x720@30',
    // eslint-disable-next-line
    video_codec: 'VP8',
    // eslint-disable-next-line
    video_bitrate: 3072,
    // eslint-disable-next-line
    video_pipeline: '',
    // eslint-disable-next-line
    video_max_fps: 25,
  
    // eslint-disable-next-line
    audio_codec: 'OPUS',
    // eslint-disable-next-line
    audio_bitrate: 128,
    // eslint-disable-next-line
    audio_pipeline: '',
  
    // eslint-disable-next-line
    broadcast_pipeline: '',
  
    // eslint-disable-next-line
    envs: {},
    // eslint-disable-next-line
    mounts: [],
  } as RoomSettings,
  videoCodecs: [
    "VP8",
    "VP9",
    "H264",
  ] as string[],
  audioCodecs: [
    "OPUS",
    "G722",
    "PCMU",
    "PCMA",
  ] as string[],
  availableScreens: [
    "1920x1080@60",
    "1920x1080@30",
    "1680x1050@60",
    "1600x900@60",
    "1440x900@60",
    "1440x810@60",
    "1400x1050@60",
    "1400x900@60",
    "1368x768@60",
    "1360x768@60",
    "1280x1024@60",
    "1280x960@60",
    "1280x800@60",
    "1280x720@60",
    "1280x720@30",
    "1152x864@60",
    "1152x648@60",
    "1024x768@60",
    "1024x576@60",
    "960x720@60",
    "960x720@30",
    "960x600@60",
    "960x540@60",
    "928x696@60",
    "896x672@60",
    "864x486@60",
    "840x525@60",
    "800x600@60",
    "800x450@60",
    "720x450@60",
    "720x405@60",
    "700x525@60",
    "700x450@60",
    "684x384@60",
    "680x384@60",
    "640x512@60",
    "640x480@60",
    "640x400@60",
    "640x360@60"
  ] as string[],
}

export type State = typeof state
