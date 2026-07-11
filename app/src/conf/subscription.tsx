import { Plan, Plans } from "@/api/types.tsx";

// 订阅功能已移除，保留此文件以保持兼容性

export const subscriptionType: Record<number, string> = {
  1: "basic",
  2: "standard",
  3: "pro",
};

export function getPlan(_data: Plans, _level: number): Plan {
  return { level: 0, price: 0, items: [] };
}

export function getPlanModels(_data: Plans, _level: number): string[] {
  return [];
}

export function includingModelFromPlan(
  _data: Plans,
  _level: number,
  _model: string,
): boolean {
  return false;
}

export function getPlanPrice(_data: Plans, _level: number): number {
  return 0;
}

export function getPlanName(_level: number): string {
  return "none";
}
