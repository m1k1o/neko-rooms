import {
  RoomsConfig,
  RoomEntry,
  RoomSettings,
  BrowserPolicyContent,
  BrowserPolicyTypeEnum,
  BrowserPolicyExtension,
  PullStatus,
} from '@/api/index'

export interface BrowserPolicyConfig {
  type: BrowserPolicyTypeEnum;
  path: string;
  profile: string;
  images: string[];
}

export interface BrowserPolicyExtensions {
  type: BrowserPolicyTypeEnum;
  path: string;
  profile: string;
  images: string[];
}

export const state = {
  //
  // rooms
  //

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
    // eslint-disable-next-line
    control_protection: false,
    // eslint-disable-next-line
    implicit_control: false,
  
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
    mounts: [],
    // eslint-disable-next-line
    resources: {
      // eslint-disable-next-line
      nano_cpus: 0,
      memory: 0,
      // eslint-disable-next-line
      shm_size: 2000000000,
    },
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

  //
  // browser policy
  //

  browserPolicyConfig: [
    {
      type: BrowserPolicyTypeEnum.firefox,
      path: '/usr/lib/firefox/distribution/policies.json',
      profile: '/home/neko/.mozilla/firefox/profile.default',
      images: [ 'm1k1o/neko:latest', 'm1k1o/neko:firefox' ],
    },
    {
      type: BrowserPolicyTypeEnum.firefox,
      path: '/usr/lib/firefox-esr/distribution/policies.json',
      profile: '/home/neko/.mozilla/firefox/profile.default',
      images: [ 'm1k1o/neko:arm-firefox' ],
    },
    {
      type: BrowserPolicyTypeEnum.chromium,
      path: '/etc/chromium/policies/managed/policies.json',
      profile: '/home/neko/.config/chromium',
      images: [ 'm1k1o/neko:chromium', 'm1k1o/neko:arm-chromium', 'm1k1o/neko:ungoogled-chromium' ],
    },
    {
      type: BrowserPolicyTypeEnum.chromium,
      path: '/etc/opt/chrome/policies/managed/policies.json',
      profile: '/home/neko/.config/google-chrome',
      images: [ 'm1k1o/neko:google-chrome' ],
    },
    {
      type: BrowserPolicyTypeEnum.chromium,
      path: '/etc/brave/policies/managed/policies.json',
      profile: '/home/neko/.config/brave',
      images: [ 'm1k1o/neko:brave' ],
    },
    {
      type: BrowserPolicyTypeEnum.chromium,
      path: '/etc/opt/edge/policies/managed/policies.json',
      profile: '/home/neko/.config/microsoft-edge',
      images: [ 'm1k1o/neko:microsoft-edge' ],
    },
  ] as BrowserPolicyConfig[],
  browserPolicyExtensions: [
    {
      text: 'uBlock Origin',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: 'uBlock0@raymondhill.net',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'cjpalhdlnbpafiamejdnhcphjbkeiagm',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'SponsorBlock for YouTube',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: 'sponsorBlocker@ajay.app',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/sponsorblock/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'mnjggcdmjocbbbhaepdhchncahnbgone',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'SoundFixer',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: 'soundfixer@unrelenting.technology',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/soundfixer/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'mbhbddecpoendcifccfckjkigbinefkg',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'Native MPEG-Dash + HLS Playback',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: 'native-dash-hls@cavar.net',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/native-mpeg-dash-hls-playback/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'cjfbmleiaobegagekpmlhmaadepdeedn',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'Google Input Tools',
      value: {
        // not available for firefox
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'mclkkofklkfljcocdinagocijmpgbhab',
        } as BrowserPolicyExtension,
      },
    },

    //
    // VPN
    //

    {
      text: 'Windscribe - Free Proxy and Ad Blocker',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: '@windscribeff',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/windscribe/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'hnmpcagpplmpfojmgmnngilcnanddlhb',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'Private Internet Access',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: '{3e4d2037-d300-4e95-859d-3cba866f46d3}',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/private-internet-access-ext/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'jplnlifepflhkbkgonidnobkakhmpnmh',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'SetupVPN - Lifetime Free VPN',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: '@setupvpncom',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/setupvpn/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'oofgbpoabipfcfjapgnbbjjaenockbdp',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'NordVPN',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: 'nordvpnproxy@nordvpn.com',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/nordvpn-proxy-extension/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'fjoaledfpmneenckfbpdfhkmimnjocfa',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'Surfshark VPN Extension',
      value: {
        [BrowserPolicyTypeEnum.firefox]: {
          id: '{732216ec-0dab-43bb-ac85-4b5e1977599d}',
          url: 'https://addons.mozilla.org/firefox/downloads/latest/surfshark-vpn-proxy/latest.xpi',
        } as BrowserPolicyExtension,
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'ailoabdmgclmfmhdagmlohpjlbpffblp',
        } as BrowserPolicyExtension,
      },
    },
    {
      text: 'Hola VPN - The Website Unblocker',
      value: {
        [BrowserPolicyTypeEnum.chromium]: {
          id: 'gkojfkhlekighikafcpjkiklfbnlmeio',
        } as BrowserPolicyExtension,
      },
    },
  ],
  defaultBrowserPolicyContent: {
    extensions: [],
    // eslint-disable-next-line
    developer_tools: false,
    // eslint-disable-next-line
    persistent_data: false,
  } as BrowserPolicyContent,

  //
  // default
  //
  pullStatus: {
    active: false,
  } as PullStatus
}

export type State = typeof state
