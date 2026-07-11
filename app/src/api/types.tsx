import { ChargeBaseProps } from "@/admin/charge.ts";
import { useMemo } from "react";
import { BotIcon, ServerIcon, UserIcon } from "lucide-react";

export const UserRole = "user";
export const AssistantRole = "assistant";
export const SystemRole = "system";
export const VirtualRolePrefix = "virtualRole::";
export const VirtualWebSearchRole = "virtualRole::websearch";
export type Role = typeof UserRole | typeof AssistantRole | typeof SystemRole;
export const Roles = [UserRole, AssistantRole, SystemRole];

export const getRoleIcon = (role: string) => {
  return useMemo(() => {
    switch (role) {
      case UserRole:
        return <UserIcon />;
      case AssistantRole:
        return <BotIcon />;
      case SystemRole:
        return <ServerIcon />;
      default:
        return <UserIcon />;
    }
  }, [role]);
};

export type Message = {
  role: string;
  content: string;
  keyword?: string;
  quota?: number;
  end?: boolean;
  plan?: boolean;
  search_query?: {
    type: string;
    search_queries: string[];
  };
  search_result?: {
    type: string;
    search_results: Array<{
      url: string;
      title: string;
      snippet: string;
      published_at?: number;
      site_name?: string;
      site_icon?: string;
    }>;
  };
  search_index?: {
    type: string;
    search_indexes: Array<{
      url: string;
      cite_index: number;
    }>;
  };
  tool_calls?: Array<{
    index: number;
    type: string;
    id: string;
    function: {
      name: string;
      arguments: string;
    };
    status?: "start" | "executing" | "success" | "error";
    result?: string;
    error?: string;
  }>;
  tool_call_id?: string;
  name?: string;
  response_type?: string;
};

export type Model = {
  id: string;
  channel_id?: number;
  channel_name?: string;
  name: string;
  description?: string;
  enabled?: boolean;
  free: boolean;
  auth: boolean;
  default: boolean;
  high_context: boolean;
  response_speed?: string;
  model_type?: string;
  function_calling?: boolean;
  vision_model?: boolean;
  ocr_model?: boolean;
  reverse_model?: boolean;
  thinking_model?: boolean;
  avatar: string;
  tag?: string[];
  channels?: string[];

  price?: ChargeBaseProps;
};

export function getModelEntryKey(
  model: Pick<Model, "id" | "channel_id">,
): string {
  return model.channel_id === undefined
    ? model.id
    : `${model.channel_id}:${model.id}`;
}

export type Id = number;

export type ConversationInstance = {
  id: number;
  name: string;
  message: Message[];
  model?: string;
  shared?: boolean;
};

export type PlanItem = {
  id: string;
  name: string;
  value: number;
  icon: string;
  models: string[];
};

export type Plan = {
  level: number;
  price: number;
  items: PlanItem[];
  discounts?: Record<string, number>;
};

export type Plans = Plan[];

export function newModel(id: string, name?: string, avatar?: string): Model {
  return {
    id,
    name: name ?? id,
    avatar: avatar ?? "",
    enabled: true,
    free: false,
    auth: false,
    default: false,
    high_context: false,
    response_speed: "standard",
    model_type: "chat",
  };
}
