import { Model } from "@/api/types.tsx";
import { CommonResponse } from "@/api/common.ts";
import axios from "axios";
import { getErrorMessage } from "@/utils/base.ts";

export async function getMarket(): Promise<Model[]> {
  try {
    const resp = await axios.get("/admin/market");
    return (resp.data || []) as Model[];
  } catch (e) {
    console.warn(e);
    return [];
  }
}

export async function updateMarket(data: Model[]): Promise<CommonResponse> {
  try {
    const resp = await axios.post("/admin/market/update", data);
    return resp.data as CommonResponse;
  } catch (e) {
    console.warn(e);
    return { status: false, error: getErrorMessage(e) };
  }
}
