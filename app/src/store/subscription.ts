import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { getSubscription } from "@/api/addition.ts";

// 订阅功能已移除，保留此文件以保持兼容性

export const subscriptionSlice = createSlice({
  name: "subscription",
  initialState: {
    is_subscribed: false,
    level: 0,
    enterprise: false,
    expired: 0,
    expired_at: "",
    refresh: 0,
    refresh_at: "",
    usage: {},
  },
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(refreshSubscription.fulfilled, (_state, _action) => {
      // 订阅功能已移除，不再更新状态
    });
  },
});

export default subscriptionSlice.reducer;

export const isSubscribedSelector = (_state: any): boolean => false;
export const levelSelector = (_state: any): number => 0;
export const expiredSelector = (_state: any): number => 0;
export const expiredAtSelector = (_state: any): string => "";
export const refreshSelector = (_state: any): number => 0;
export const refreshAtSelector = (_state: any): string => "";
export const usageSelector = (_state: any): any => ({});

export const refreshSubscription = createAsyncThunk(
  "subscription/refreshSubscription",
  async () => {
    return await getSubscription();
  },
);
