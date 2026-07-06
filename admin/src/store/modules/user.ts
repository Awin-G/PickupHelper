import { defineStore } from "pinia";
import {
  type userType,
  store,
  router,
  resetRouter,
  routerArrays,
  storageLocal
} from "../utils";
import { type LoginResult, type RefreshTokenResult, adminLogin, refreshTokenApi } from "@/api/auth";
import { useMultiTagsStoreHook } from "./multiTags";
import { type DataInfo, setToken, removeToken, userKey } from "@/utils/auth";

export const useUserStore = defineStore("pure-user", {
  state: (): userType => ({
    avatar: storageLocal().getItem<DataInfo<number>>(userKey)?.avatar ?? "",
    username: storageLocal().getItem<DataInfo<number>>(userKey)?.username ?? "",
    nickname: storageLocal().getItem<DataInfo<number>>(userKey)?.nickname ?? "",
    roles: storageLocal().getItem<DataInfo<number>>(userKey)?.roles ?? [],
    permissions:
      storageLocal().getItem<DataInfo<number>>(userKey)?.permissions ?? [],
    verifyCode: "",
    currentPage: 0,
    isRemembered: false,
    loginDay: 7
  }),
  actions: {
    SET_AVATAR(avatar: string) {
      this.avatar = avatar;
    },
    SET_USERNAME(username: string) {
      this.username = username;
    },
    SET_NICKNAME(nickname: string) {
      this.nickname = nickname;
    },
    SET_ROLES(roles: Array<string>) {
      this.roles = roles;
    },
    SET_PERMS(permissions: Array<string>) {
      this.permissions = permissions;
    },
    SET_VERIFYCODE(verifyCode: string) {
      this.verifyCode = verifyCode;
    },
    SET_CURRENTPAGE(value: number) {
      this.currentPage = value;
    },
    SET_ISREMEMBERED(bool: boolean) {
      this.isRemembered = bool;
    },
    SET_LOGINDAY(value: number) {
      this.loginDay = Number(value);
    },
    async loginByUsername(data: { username: string; password: string }) {
      return new Promise<LoginResult>((resolve, reject) => {
        adminLogin(data)
          .then(data => {
            if (data.code === 0) {
              const userData = data.data;
              setToken({
                accessToken: userData.access_token,
                refreshToken: userData.refresh_token,
                expires: new Date(Date.now() + userData.expires_in * 1000),
                avatar: userData.user.avatar || "",
                username: userData.user.nickname || userData.user.phone,
                nickname: userData.user.nickname || "",
                roles: userData.user.roles || [],
                permissions: userData.user.permissions || []
              });
              resolve(data);
            } else {
              reject(data.message);
            }
          })
          .catch(error => {
            reject(error);
          });
      });
    },
    logOut() {
      this.username = "";
      this.roles = [];
      this.permissions = [];
      removeToken();
      useMultiTagsStoreHook().handleTags("equal", [...routerArrays]);
      resetRouter();
      router.push("/login");
    },
    async handRefreshToken(data: { refresh_token: string }) {
      return new Promise<RefreshTokenResult>((resolve, reject) => {
        refreshTokenApi(data)
          .then(data => {
            if (data.code === 0) {
              setToken({
                accessToken: data.data.access_token,
                refreshToken: "",
                expires: new Date(Date.now() + data.data.expires_in * 1000)
              });
              resolve(data);
            } else {
              reject(data.message);
            }
          })
          .catch(error => {
            reject(error);
          });
      });
    }
  }
});

export function useUserStoreHook() {
  return useUserStore(store);
}
