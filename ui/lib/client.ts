import GLOBAL from "@/global";
import { ServerError } from "@/types/server-error";

async function refreshToken() {
    return new Promise((resolve, reject) => {
      fetch(GLOBAL.API_URL + "/auth/refresh-token", {
        method: "GET",
      })
      .then((res) => {
        if (res.ok) {
          return res.text();
        } else {
            throw res;
        }
      })
      .then((data) => {
        resolve(data);
      })
      .catch((err) => {
        reject(null);
      })
    });
  }

const handleResponse = async <T>(
    response: Response,
    originalUrl: string,
    method: string,
    headers: any
): Promise<T> => {
    if (!response.ok) {
        if (response.status === 401) {
            fetch(GLOBAL.API_URL + "/auth/refresh-token", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                credentials: "include",
            }).then((res) => {
                if (!res.ok) {
                    if (window != undefined) {
                        window.location.href = "/login";
                    }

                    return res.json();
                } else {
                    fetch(originalUrl, {
                        method: method,
                        headers: headers,
                        credentials: "include",
                    }).then((res) => {
                        return res.json();
                    });
                }
            });
        }

        throw errorBody as ServerError;
    }

    return response.json();
};

const request = {
    get: async <T>(path: string): Promise<T> => {
        const originalUrl = GLOBAL.API_URL + path;
        const response = await fetch(originalUrl, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
            credentials: "include",
        });
        return handleResponse<T>(response, originalUrl);
    },
    post: async <T>(path: string, body: {}): Promise<T> => {
        const response = await fetch(GLOBAL.API_URL + path, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            credentials: "include",
            body: JSON.stringify(body),
        });
        return handleResponse<T>(response);
    },
};

export default request;
