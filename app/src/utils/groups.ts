import { useSelector } from "react-redux";
import { selectAdmin, selectAuthenticated } from "@/store/auth.ts";
import { useMemo } from "react";

export const AnonymousType = "anonymous";
export const NormalType = "normal";
export const BasicType = "basic";
export const StandardType = "standard";
export const ProType = "pro";
export const AdminType = "admin";

export const allGroups: string[] = [
  AnonymousType,
  NormalType,
  BasicType,
  StandardType,
  ProType,
  AdminType,
];

export function useGroup(countAdminLevel?: boolean): string {
  const auth = useSelector(selectAuthenticated);
  const admin = useSelector(selectAdmin);

  return useMemo(() => {
    if (!auth) return AnonymousType;
    if (countAdminLevel && admin) return AdminType;
    // 订阅功能已移除，所有用户都是 normal
    return NormalType;
  }, [auth, admin]);
}

export function hitGroup(group: string[]): boolean {
  const current = useGroup();
  const admin = useSelector(selectAdmin);

  return useMemo(() => {
    if (group.includes(AdminType) && admin) return true;
    return group.includes(current);
  }, [group, current, admin]);
}
