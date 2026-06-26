import { Component, createSignal, onMount, For, Show } from 'solid-js';
import { api } from '../api/client';
import type { Post } from '../api/client';

interface Props {
  onNavigate: (page: string) => void;
}

const mediaIcons: Record<string, string> = {
  photo: '📷', video: '🎥', video_note: '🔵', text: '📝',
};

const PostList: Component<Props> = (props) => {
  const [posts, setPosts] = createSignal<Post[]>([]);
  const [loading, setLoading] = createSignal(true);
  const [filter, setFilter] = createSignal('');

  onMount(async () => {
    try {
      const data = await api.getPosts(filter() ? { status: filter() } : undefined);
      setPosts(data);
    } catch {
      // handle error
    } finally {
      setLoading(false);
    }
  });

  const loadPosts = async (status: string) => {
    setFilter(status);
    setLoading(true);
    try {
      const data = await api.getPosts(status ? { status } : undefined);
      setPosts(data);
    } catch {
      // handle error
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('uz-UZ', {
      day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit',
    });
  };

  return (
    <div class="fade-in">
      <div class="app-header">
        <button
          class="btn btn--secondary btn--sm"
          onClick={() => props.onNavigate('home')}
        >
          ← Orqaga
        </button>
        <div>
          <div class="app-header__title" style="font-size: 18px;">Postlar</div>
        </div>
      </div>

      {/* Filters */}
      <div class="nav-tabs">
        {[
          { value: '', label: 'Hammasi' },
          { value: 'draft', label: 'Qoralama' },
          { value: 'scheduled', label: 'Rejada' },
          { value: 'sent', label: 'Yuborilgan' },
          { value: 'failed', label: 'Xato' },
        ].map(f => (
          <button
            class={`nav-tab ${filter() === f.value ? 'nav-tab--active' : ''}`}
            onClick={() => loadPosts(f.value)}
          >
            {f.label}
          </button>
        ))}
      </div>

      {/* Post list */}
      <Show when={!loading()} fallback={
        <div>
          <div class="skeleton" style="height: 64px; margin-bottom: 8px;" />
          <div class="skeleton" style="height: 64px; margin-bottom: 8px;" />
          <div class="skeleton" style="height: 64px; margin-bottom: 8px;" />
        </div>
      }>
        <Show when={posts().length === 0}>
          <div class="empty-state">
            <div class="empty-state__icon">📭</div>
            <div class="empty-state__title">Post topilmadi</div>
            <div class="empty-state__desc">
              Yangi post yaratish uchun tugmani bosing.
            </div>
            <button
              class="btn btn--primary"
              style="margin-top: 16px;"
              onClick={() => props.onNavigate('create')}
            >
              ✨ Yangi post
            </button>
          </div>
        </Show>

        <For each={posts()}>
          {(post) => (
            <div class="post-item">
              <div class={`post-item__icon post-item__icon--${post.media_type}`}>
                {mediaIcons[post.media_type] || '📄'}
              </div>
              <div class="post-item__info">
                <div class="post-item__title">
                  {post.caption?.slice(0, 50) || `${post.media_type} post`}
                  {post.caption && post.caption.length > 50 ? '...' : ''}
                </div>
                <div class="post-item__meta">
                  {formatDate(post.created_at)}
                  {post.scheduled_at && ` · 📅 ${formatDate(post.scheduled_at)}`}
                </div>
              </div>
              <span class={`status-badge status-badge--${post.status}`}>
                {post.status === 'draft' ? '📝' :
                 post.status === 'scheduled' ? '⏰' :
                 post.status === 'sent' ? '✅' : '❌'}
                {' '}{post.status}
              </span>
            </div>
          )}
        </For>
      </Show>
    </div>
  );
};

export default PostList;
