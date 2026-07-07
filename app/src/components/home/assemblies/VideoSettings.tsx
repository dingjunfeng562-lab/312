import { useDispatch, useSelector } from "react-redux";
import { useTranslation } from "react-i18next";
import {
  selectModel,
  selectVideoAspectRatio,
  selectVideoDuration,
  selectVideoResolution,
  setVideoAspectRatio,
  setVideoDuration,
  setVideoResolution,
} from "@/store/chat.ts";
import {
  getVideoCapabilities,
  normalizeVideoSettings,
  videoDurationOptions,
  videoResolutionOptions,
} from "@/conf/video.ts";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover.tsx";
import { Video, Settings2 } from "lucide-react";
import Icon from "@/components/utils/Icon.tsx";
import { cn } from "@/components/ui/lib/utils.ts";
import { ChatAction } from "./ChatAction.tsx";
import { useEffect, useMemo } from "react";

type OptionButtonProps = {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
};

function OptionButton({ active, onClick, children }: OptionButtonProps) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "px-3 py-1.5 rounded-md text-sm font-medium transition-all duration-200 border",
        active
          ? "bg-primary text-primary-foreground border-primary shadow-sm"
          : "bg-muted/50 text-muted-foreground border-transparent hover:bg-muted hover:text-foreground",
      )}
    >
      {children}
    </button>
  );
}

export function VideoSettings() {
  const { t } = useTranslation();
  const dispatch = useDispatch();
  const model = useSelector(selectModel);
  const aspectRatio = useSelector(selectVideoAspectRatio);
  const duration = useSelector(selectVideoDuration);
  const resolution = useSelector(selectVideoResolution);
  const capabilities = useMemo(() => getVideoCapabilities(model), [model]);

  useEffect(() => {
    const normalized = normalizeVideoSettings(model, {
      aspectRatio,
      duration,
      resolution,
    });

    if (normalized.aspectRatio !== aspectRatio) {
      dispatch(setVideoAspectRatio(normalized.aspectRatio));
    }
    if (normalized.duration !== duration) {
      dispatch(setVideoDuration(normalized.duration));
    }
    if (normalized.resolution !== resolution) {
      dispatch(setVideoResolution(normalized.resolution));
    }
  }, [aspectRatio, dispatch, duration, model, resolution]);

  const aspectRatios = useMemo(
    () =>
      capabilities.aspectRatios.map((value) => ({
        value,
        label: value,
      })),
    [capabilities],
  );
  const durations = useMemo(
    () =>
      capabilities.durations.map((value) => ({
        value,
        label: videoDurationOptions[value],
      })),
    [capabilities],
  );
  const resolutions = useMemo(
    () =>
      capabilities.resolutions.map((value) => ({
        value,
        label: videoResolutionOptions[value],
      })),
    [capabilities],
  );

  return (
    <Popover>
      <PopoverTrigger asChild>
        <div>
          <ChatAction
            active
            text={t("chat.video-settings")}
          >
            <Icon
              icon={<Video className="h-4 w-4 text-violet-500" />}
            />
          </ChatAction>
        </div>
      </PopoverTrigger>
      <PopoverContent
        className="w-80 p-4"
        side="top"
        align="start"
      >
        <div className="space-y-4">
          <div className="flex items-center gap-2 mb-1">
            <Icon icon={<Settings2 className="h-4 w-4 text-violet-500" />} />
            <div className="flex min-w-0 flex-col">
              <span className="text-sm font-medium">{t("chat.video-settings")}</span>
              <span className="text-xs text-muted-foreground">
                {videoResolutionOptions[resolution]} · {aspectRatio} · {duration}s
              </span>
            </div>
          </div>

          {/* Aspect Ratio */}
          <div className="space-y-2">
            <label className="text-xs text-muted-foreground font-medium">
              {t("chat.video-aspect-ratio")}
            </label>
            <div className="flex flex-wrap gap-1.5">
              {aspectRatios.map((item) => (
                <OptionButton
                  key={item.value}
                  active={aspectRatio === item.value}
                  onClick={() => dispatch(setVideoAspectRatio(item.value))}
                >
                  {item.label}
                </OptionButton>
              ))}
            </div>
          </div>

          {/* Duration */}
          <div className="space-y-2">
            <label className="text-xs text-muted-foreground font-medium">
              {t("chat.video-duration")}
            </label>
            <div className="flex flex-wrap gap-1.5">
              {durations.map((item) => (
                <OptionButton
                  key={item.value}
                  active={duration === item.value}
                  onClick={() => dispatch(setVideoDuration(item.value))}
                >
                  {item.label}
                </OptionButton>
              ))}
            </div>
          </div>

          {/* Resolution */}
          <div className="space-y-2">
            <label className="text-xs text-muted-foreground font-medium">
              {t("chat.video-resolution")}
            </label>
            <div className="flex flex-wrap gap-1.5">
              {resolutions.map((item) => (
                <OptionButton
                  key={item.value}
                  active={resolution === item.value}
                  onClick={() => dispatch(setVideoResolution(item.value))}
                >
                  {item.label}
                </OptionButton>
              ))}
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
