import axios from 'axios';

const axiosInstance = axios.create({
  baseURL: import.meta.env.VITE_APP_API_BASE_URL!
})

const networkService = {
  delete: async function(url: string) {
    const res = await axiosInstance.delete(url);
    return res.data;
  },

  get: async function(url: string, headers?: object) {
    const res = await axiosInstance.get(url, (headers) ? { headers } : undefined);
    return res.data;
  },

  patch: async function(url: string, data: object) {
    const res = await axiosInstance.patch(url, data);
    return res.data;
  },

  post: async function(url: string, data: object) {
    const res = await axiosInstance.post(url, data);
    return res.data;
  },

  put: async function(url: string, data: object) {
    const res = await axiosInstance.put(url, data);
    return res.data;
  }
}

export default networkService;