import { request } from './request';
import type { Station } from './types';

export const stationApi = {
  list: () =>
    request<Station[]>({
      url: '/stations',
      method: 'GET',
    }),
};
