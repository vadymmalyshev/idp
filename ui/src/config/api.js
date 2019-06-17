import { SERVER_URL } from '../config';

// const basePath = `${SERVER_URL}/api`;
const basePath = `/api`;
const API = {
  login: `${basePath}/login`,
  register: `${basePath}/register`,
  forgot: `${basePath}/forgot`,
  account: `${basePath}/account`,
}

export default API;