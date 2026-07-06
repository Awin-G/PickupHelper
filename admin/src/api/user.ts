import { http } from "@/utils/http";
import type {
  UserItem,
  RunnerApplicationItem,
  RunnerAuditRequest,
  PaginatedList
} from "./types/parcel";

/** 用户列表 */
export const getUserList = (params?: {
  keyword?: string;
  user_type?: number;
  is_blacklisted?: number;
  page?: number;
  page_size?: number;
}) => {
  return http.request<PaginatedList<UserItem>>("get", "/users", { params });
};

/** 跑腿员申请列表 */
export const getRunnerApplications = (params?: {
  status?: number;
  keyword?: string;
  page?: number;
  page_size?: number;
}) => {
  return http.request<PaginatedList<RunnerApplicationItem>>(
    "get",
    "/user/runner/applications",
    { params }
  );
};

/** 审核跑腿员申请 */
export const auditRunnerApplication = (
  id: number,
  data: RunnerAuditRequest
) => {
  const url = `/user/runner/applications/${id}/audit`;
  return http.request<RunnerApplicationItem>("put", url, { data });
};
