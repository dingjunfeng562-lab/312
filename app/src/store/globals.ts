import { createSlice } from "@reduxjs/toolkit";
import { Plans } from "@/api/types.tsx";
import { AppDispatch, RootState } from "@/store/index.ts";
import { getTheme, Theme } from "@/components/ThemeProvider.tsx";

type GlobalState = {
  theme: Theme;
};

export const globalSlice = createSlice({
  name: "global",
  initialState: {
    theme: getTheme(),
  } as GlobalState,
  reducers: {
    setTheme: (state, action) => {
      state.theme = action.payload;
    },
  },
});

export const { setTheme } = globalSlice.actions;

export default globalSlice.reducer;

export const themeSelector = (state: RootState): Theme => state.global.theme;

// 订阅功能已移除，返回空数组以保持兼容性
export const subscriptionDataSelector = (_state: RootState): Plans => [];
export const dispatchSubscriptionData = (
  _dispatch: AppDispatch,
  _subscription: Plans,
) => {
  // 订阅功能已移除，空函数保持兼容性
};
