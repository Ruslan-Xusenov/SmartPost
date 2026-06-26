import { Component, For } from 'solid-js';
import type { Channel } from '../api/client';
import { selectedChannel, setSelectedChannel } from '../stores/postStore';

interface Props {
  channels: Channel[];
}

const ChannelSelector: Component<Props> = (props) => {
  const haptic = () => window.Telegram?.WebApp?.HapticFeedback;

  return (
    <div>
      <div class="section-title">📢 Kanalni tanlang</div>
      <div class="channel-list">
        <For each={props.channels}>
          {(ch) => (
            <div
              class={`channel-item ${selectedChannel()?.id === ch.id ? 'channel-item--selected' : ''}`}
              onClick={() => {
                setSelectedChannel(ch);
                haptic()?.selectionChanged();
              }}
            >
              <div class="channel-item__avatar">
                {ch.title.charAt(0).toUpperCase()}
              </div>
              <div>
                <div class="channel-item__name">{ch.title}</div>
                {ch.username && (
                  <div class="channel-item__username">@{ch.username}</div>
                )}
              </div>
            </div>
          )}
        </For>
      </div>
    </div>
  );
};

export default ChannelSelector;
