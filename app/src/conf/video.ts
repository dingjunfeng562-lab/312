export type VideoAspectRatio = "16:9" | "9:16" | "1:1" | "4:3" | "3:4";
export type VideoDuration = "4" | "5" | "8" | "10" | "12" | "15" | "16" | "20" | "30" | "60";
export type VideoResolution = "480p" | "720p" | "1024p" | "1080p" | "4k";

export type VideoSettingsValue = {
  aspectRatio: VideoAspectRatio;
  duration: VideoDuration;
  resolution: VideoResolution;
};

export type VideoCapabilities = {
  aspectRatios: VideoAspectRatio[];
  durations: VideoDuration[];
  resolutions: VideoResolution[];
  defaults: VideoSettingsValue;
};

export const defaultVideoSettings: VideoSettingsValue = {
  aspectRatio: "16:9",
  duration: "5",
  resolution: "720p",
};

const genericVideoCapabilities: VideoCapabilities = {
  aspectRatios: ["16:9", "9:16", "1:1", "4:3", "3:4"],
  durations: ["5", "10", "15", "30", "60"],
  resolutions: ["480p", "720p", "1080p", "4k"],
  defaults: defaultVideoSettings,
};

const soraVideoCapabilities: VideoCapabilities = {
  aspectRatios: ["16:9", "9:16"],
  durations: ["4", "8", "12", "16", "20"],
  resolutions: ["720p", "1024p"],
  defaults: {
    aspectRatio: "16:9",
    duration: "4",
    resolution: "720p",
  },
};

const soraProVideoCapabilities: VideoCapabilities = {
  ...soraVideoCapabilities,
  resolutions: ["720p", "1024p", "1080p"],
};

const videoSizes: Record<VideoResolution, Record<VideoAspectRatio, string>> = {
  "480p": {
    "16:9": "854x480",
    "9:16": "480x854",
    "1:1": "480x480",
    "4:3": "640x480",
    "3:4": "480x640",
  },
  "720p": {
    "16:9": "1280x720",
    "9:16": "720x1280",
    "1:1": "720x720",
    "4:3": "960x720",
    "3:4": "720x960",
  },
  "1024p": {
    "16:9": "1792x1024",
    "9:16": "1024x1792",
    "1:1": "1024x1024",
    "4:3": "1365x1024",
    "3:4": "1024x1365",
  },
  "1080p": {
    "16:9": "1920x1080",
    "9:16": "1080x1920",
    "1:1": "1080x1080",
    "4:3": "1440x1080",
    "3:4": "1080x1440",
  },
  "4k": {
    "16:9": "3840x2160",
    "9:16": "2160x3840",
    "1:1": "2160x2160",
    "4:3": "2880x2160",
    "3:4": "2160x2880",
  },
};

export const videoDurationOptions: Record<VideoDuration, string> = {
  "4": "4s",
  "5": "5s",
  "8": "8s",
  "10": "10s",
  "12": "12s",
  "15": "15s",
  "16": "16s",
  "20": "20s",
  "30": "30s",
  "60": "60s",
};

export const videoResolutionOptions: Record<VideoResolution, string> = {
  "480p": "480p",
  "720p": "720p",
  "1024p": "1024p",
  "1080p": "1080p",
  "4k": "4K",
};

export function getVideoCapabilities(model: string): VideoCapabilities {
  const normalized = model.toLowerCase().trim();

  if (/sora-2-pro/.test(normalized)) return soraProVideoCapabilities;
  if (/sora-2|sora/.test(normalized)) return soraVideoCapabilities;

  return genericVideoCapabilities;
}

export function normalizeVideoSettings(
  model: string,
  settings: VideoSettingsValue,
): VideoSettingsValue {
  const capabilities = getVideoCapabilities(model);

  return {
    aspectRatio: capabilities.aspectRatios.includes(settings.aspectRatio)
      ? settings.aspectRatio
      : capabilities.defaults.aspectRatio,
    duration: capabilities.durations.includes(settings.duration)
      ? settings.duration
      : capabilities.defaults.duration,
    resolution: capabilities.resolutions.includes(settings.resolution)
      ? settings.resolution
      : capabilities.defaults.resolution,
  };
}

export function getVideoSize(
  resolution: VideoResolution,
  aspectRatio: VideoAspectRatio,
): string {
  return videoSizes[resolution][aspectRatio];
}

export function getVideoRequestSettings(
  model: string,
  settings: VideoSettingsValue,
): { seconds: string; size: string } {
  const normalized = normalizeVideoSettings(model, settings);

  return {
    seconds: normalized.duration,
    size: getVideoSize(normalized.resolution, normalized.aspectRatio),
  };
}
