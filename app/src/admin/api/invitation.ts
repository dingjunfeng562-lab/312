import axios from "axios";

export type InvitationCode = {
  code: string;
  quota: number;
  type: string;
  used: boolean;
  created_at: string;
  updated_at: string;
  username: string;
  expires_at?: string;
  is_expired: boolean;
  creator_name: string;
  used_at?: string;
  used_ip?: string;
  notes: string;
  status?: string;
};

export type InvitationList = {
  status: boolean;
  total: number;
  data: InvitationCode[];
  message?: string;
};

export type GenerateInvitationRequest = {
  type: string;
  quota: number;
  number: number;
};

export type GenerateInvitationAdvancedRequest = {
  type: string;
  quota: number;
  number: number;
  expires_days: number;
  notes: string;
};

export type GenerateInvitationResponse = {
  status: boolean;
  message?: string;
  data?: string[];
};

export type InvitationUsageDetail = {
  code: string;
  quota: number;
  type: string;
  used: boolean;
  created_at: string;
  expires_at?: string;
  is_expired: boolean;
  creator_name: string;
  used_by_user: string;
  used_at?: string;
  used_ip: string;
  notes: string;
};

export type InvitationUsageResponse = {
  status: boolean;
  data?: InvitationUsageDetail;
  message?: string;
};

export async function getInvitationList(
  page: number,
): Promise<InvitationList> {
  try {
    const response = await axios.get(`/admin/invitation/list`, {
      params: { page },
    });
    return response.data as InvitationList;
  } catch (e) {
    console.error(e);
    return { status: false, total: 0, data: [], message: "network error" };
  }
}

export async function generateInvitation(
  data: GenerateInvitationRequest,
): Promise<GenerateInvitationResponse> {
  try {
    const response = await axios.post(`/admin/invitation/generate`, data);
    return response.data as GenerateInvitationResponse;
  } catch (e) {
    console.error(e);
    return { status: false, message: "network error" };
  }
}

export async function generateInvitationAdvanced(
  data: GenerateInvitationAdvancedRequest,
): Promise<GenerateInvitationResponse> {
  try {
    const response = await axios.post(
      `/admin/invitation/generate-advanced`,
      data,
    );
    return response.data as GenerateInvitationResponse;
  } catch (e) {
    console.error(e);
    return { status: false, message: "network error" };
  }
}

export async function deleteInvitation(
  code: string,
): Promise<{ status: boolean; error?: string }> {
  try {
    const response = await axios.post(`/admin/invitation/delete`, { code });
    return response.data;
  } catch (e) {
    console.error(e);
    return { status: false, error: "network error" };
  }
}

export async function disableInvitation(
  code: string,
): Promise<{ status: boolean; error?: string }> {
  try {
    const response = await axios.post(`/admin/invitation/disable`, { code });
    return response.data;
  } catch (e) {
    console.error(e);
    return { status: false, error: "network error" };
  }
}

export async function enableInvitation(
  code: string,
): Promise<{ status: boolean; error?: string }> {
  try {
    const response = await axios.post(`/admin/invitation/enable`, { code });
    return response.data;
  } catch (e) {
    console.error(e);
    return { status: false, error: "network error" };
  }
}

export async function getInvitationUsage(
  code: string,
): Promise<InvitationUsageResponse> {
  try {
    const response = await axios.get(
      `/admin/invitation/usage/${encodeURIComponent(code)}`,
    );
    return response.data as InvitationUsageResponse;
  } catch (e) {
    console.error(e);
    return { status: false, message: "network error" };
  }
}
