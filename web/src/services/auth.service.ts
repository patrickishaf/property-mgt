import networkService from "./network.service";

const authService = {
  loginWithPassword: async (identifier: string, password: string) => {
    const acc = await networkService.post('/auth/login', {
      identifier,
      password,
    });
    return acc;
  },

  sendOtp: async (identifier: string) => {
    const { code } = await networkService.post('/auth/otp', { identifier });
    return code;
  },

  verifyOtp: async (identifier: string, code: string) => {
    const acc = await networkService.post('/auth/otp-verify', { identifier, code });
    return acc;
  },

  logout: async (identifier: string) => {
    const res = await networkService.post('/auth/logout', { identifier });
    return res;
  },
}

export default authService;