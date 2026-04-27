// Minimal unauthenticated JSON fetch helper (no cookies, no auth headers). Prefer `axios` for
// anything that may need a JWT or the refresh cookie; this is for simple or pre-login use.
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface ApiResponse<T> {
  data: T | null;
  error: string | null;
}

interface ApiErrorResponse {
  message?: string;
}

export function extractApiErrorMessage(fallbackMessage: string, payload: string): string {
  try {
    const parsed = JSON.parse(payload) as ApiErrorResponse;
    if (parsed.message && typeof parsed.message === "string") {
      return parsed.message;
    }
    return fallbackMessage;
  } catch {
    return payload || fallbackMessage;
  }
}

async function request<T>(
  path: string,
  options?: RequestInit
): Promise<ApiResponse<T>> {
  try {
    const res = await fetch(`${API_URL}${path}`, {
      headers: { "Content-Type": "application/json" },
      ...options,
    });
    if (!res.ok) {
      const text = await res.text();
      return {
        data: null,
        error: extractApiErrorMessage(res.statusText, text),
      };
    }
    const data = await res.json();
    return { data, error: null };
  } catch (err) {
    return { data: null, error: (err as Error).message };
  }
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "POST", body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body) }),
  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PATCH", body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};
