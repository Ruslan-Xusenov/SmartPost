import { Component, createSignal, onMount } from 'solid-js';
import { api } from '../api/client';
import type { Channel } from '../api/client';
import { setChannels } from '../stores/postStore';

interface Props {
  onNavigate: (page: string) => void;
}

const Home: Component<Props> = (props) => {
  const [channelList, setChannelList] = createSignal<Channel[]>([]);
  const [loading, setLoading] = createSignal(true);

  onMount(async () => {
    try {
      const chs = await api.getChannels();
      setChannelList(chs);
      setChannels(chs);
    } catch {
      // Offline or not authenticated
    } finally {
      setLoading(false);
    }
  });

  return (
    <div class="fade-in">
      {/* Header */}
      <div class="app-header">
        <div class="app-header__logo">📡</div>
        <div>
          <div class="app-header__title">SmartPost</div>
          <div class="app-header__subtitle">Kanallarni boshqarish tizimi</div>
        </div>
      </div>

      {/* Quick actions */}
      <div style="display: flex; gap: 8px; margin-bottom: 20px;">
        <button
          class="btn btn--primary btn--full"
          onClick={() => props.onNavigate('create')}
        >
          ✨ Yangi post
        </button>
        <button
          class="btn btn--secondary btn--full"
          onClick={() => props.onNavigate('posts')}
        >
          📋 Postlar
        </button>
      </div>

      {/* Channels */}
      <div class="section-title">📢 Kanallarim</div>
      {loading() ? (
        <div>
          <div class="skeleton" style="height: 68px; margin-bottom: 8px;" />
          <div class="skeleton" style="height: 68px; margin-bottom: 8px;" />
        </div>
      ) : channelList().length === 0 ? (
        <div class="empty-state">
          <div class="empty-state__icon">📢</div>
          <div class="empty-state__title">Kanal topilmadi</div>
          <div class="empty-state__desc">
            Botni kanalingizga admin sifatida qo'shing.
          </div>
        </div>
      ) : (
        <div class="channel-list">
          {channelList().map((ch) => (
            <div class="channel-item" onClick={() => props.onNavigate('create')}>
              <div class="channel-item__avatar">
                {ch.title.charAt(0).toUpperCase()}
              </div>
              <div>
                <div class="channel-item__name">{ch.title}</div>
                {ch.username && (
                  <div class="channel-item__username">@{ch.username}</div>
                )}
              </div>
              <div style="margin-left: auto;">
                <span class="status-badge status-badge--sent">✅ Aktiv</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Stats card */}
      <div class="card card--glass" style="margin-top: 16px;">
        <div style="display: flex; justify-content: space-around; text-align: center;">
          <div>
            <div style="font-size: 24px; font-weight: 700; color: var(--accent);">
              {channelList().length}
            </div>
            <div style="font-size: 12px; color: var(--tg-hint);">Kanallar</div>
          </div>
          <div style="width: 1px; background: var(--border);" />
          <div>
            <div style="font-size: 24px; font-weight: 700; color: var(--success);">—</div>
            <div style="font-size: 12px; color: var(--tg-hint);">Yuborilgan</div>
          </div>
          <div style="width: 1px; background: var(--border);" />
          <div>
            <div style="font-size: 24px; font-weight: 700; color: var(--warning);">—</div>
            <div style="font-size: 12px; color: var(--tg-hint);">Rejada</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home;
