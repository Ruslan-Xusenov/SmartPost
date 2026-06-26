const API_BASE = '/api';

function getInitData(): string {
  return window.Telegram?.WebApp?.initData || '';
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Telegram-Init-Data': getInitData(),
    ...(options.headers as Record<string, string> || {}),
  };

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `HTTP ${res.status}`);
  }

  if (res.status === 204) return {} as T;
  return res.json();
}

export interface Channel {
  id: number;
  chat_id: number;
  title: string;
  username: string;
  is_active: boolean;
}

export interface Button {
  id?: number;
  text: string;
  url: string;
  color_code: string;
  row_index: number;
}

export interface Post {
  id: number;
  channel_id: number;
  media_type: string;
  file_id: string;
  caption: string;
  status: string;
  scheduled_at?: string;
  sent_at?: string;
  error_message?: string;
  created_at: string;
  buttons?: Button[];
}

export const api = {
  getChannels: () => request<Channel[]>('/channels'),
  
  createPost: (data: {
    channel_id: number;
    media_type: string;
    file_id: string;
    caption: string;
    buttons: Button[];
  }) => request<{ id: number }>('/posts', {
    method: 'POST',
    body: JSON.stringify(data),
  }),

  getPosts: (params?: { status?: string; channel_id?: string }) => {
    const qs = new URLSearchParams(params as Record<string, string>).toString();
    return request<Post[]>(`/posts${qs ? '?' + qs : ''}`);
  },

  getPost: (id: number) => request<Post>(`/posts/${id}`),

  updatePost: (id: number, data: any) => request<any>(`/posts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),

  deletePost: (id: number) => request<void>(`/posts/${id}`, {
    method: 'DELETE',
  }),

  sendPost: (id: number) => request<{ status: string }>(`/posts/${id}/send`, {
    method: 'POST',
  }),

  schedulePost: (id: number, scheduledAt: string) =>
    request<{ status: string }>(`/posts/${id}/schedule`, {
      method: 'POST',
      body: JSON.stringify({ scheduled_at: scheduledAt }),
    }),
};
