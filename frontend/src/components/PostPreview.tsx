import { Component, Show, For } from 'solid-js';
import { mediaType, caption, buttons, mediaPreview } from '../stores/postStore';

const colorEmojiMap: Record<string, string> = {
  green: '🟢', red: '🔴', blue: '🔵', yellow: '🟡',
  orange: '🟠', purple: '🟣', white: '⚪', black: '⚫',
};

const PostPreview: Component = () => {
  const displayText = (text: string, color: string) => {
    const emoji = colorEmojiMap[color];
    return emoji ? `${emoji} ${text}` : text;
  };

  // Group buttons by row_index
  const groupedButtons = () => {
    const groups: Record<number, typeof buttons.prototype[]> = {};
    for (const btn of buttons()) {
      if (!groups[btn.row_index]) groups[btn.row_index] = [];
      groups[btn.row_index].push(btn);
    }
    return Object.entries(groups).sort(([a], [b]) => Number(a) - Number(b));
  };

  const hasContent = () => caption() || mediaPreview() || buttons().length > 0;

  return (
    <Show when={hasContent()}>
      <div class="preview-container slide-up">
        <div class="preview__label">👁 Ko'rib chiqish</div>
        <div class="preview-msg">
          {/* Media preview */}
          <Show when={mediaPreview()}>
            <Show when={mediaType() === 'video_note'}>
              <video
                class="preview-msg__video-note"
                src={mediaPreview()}
                muted
                autoplay
                loop
                playsinline
              />
            </Show>
            <Show when={mediaType() === 'photo'}>
              <img class="preview-msg__media" src={mediaPreview()} alt="Preview" />
            </Show>
            <Show when={mediaType() === 'video'}>
              <video
                class="preview-msg__media"
                src={mediaPreview()}
                controls
                playsinline
              />
            </Show>
          </Show>

          {/* Caption text */}
          <Show when={caption()}>
            <div class="preview-msg__text">{caption()}</div>
          </Show>

          {/* Inline buttons */}
          <Show when={buttons().length > 0}>
            <div class="preview-msg__buttons">
              <For each={groupedButtons()}>
                {([_, rowButtons]) => (
                  <div class="preview-msg__btn-row">
                    <For each={rowButtons as any[]}>
                      {(btn: any) => (
                        <button class="preview-msg__btn">
                          {displayText(btn.text || 'Tugma', btn.color_code)}
                        </button>
                      )}
                    </For>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </div>
      </div>
    </Show>
  );
};

export default PostPreview;
