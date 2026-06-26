import { Component, For } from 'solid-js';
import { buttons, addButton, updateButton, removeButton } from '../stores/postStore';

const colorOptions = [
  { value: 'default', emoji: '⬜', bg: '#666' },
  { value: 'green', emoji: '🟢', bg: '#00c853' },
  { value: 'red', emoji: '🔴', bg: '#ff1744' },
  { value: 'blue', emoji: '🔵', bg: '#2979ff' },
  { value: 'yellow', emoji: '🟡', bg: '#ffd600' },
  { value: 'orange', emoji: '🟠', bg: '#ff9100' },
  { value: 'purple', emoji: '🟣', bg: '#aa00ff' },
];

const ButtonEditor: Component = () => {
  return (
    <div>
      <div class="section-title">🔘 Inline tugmalar</div>

      <For each={buttons()}>
        {(btn, index) => (
          <div class="btn-editor">
            <div class="btn-editor__row">
              <input
                class="btn-editor__input"
                placeholder="Tugma matni"
                value={btn.text}
                onInput={(e) => updateButton(index(), 'text', e.currentTarget.value)}
              />
              <input
                class="btn-editor__input"
                placeholder="https://..."
                value={btn.url}
                onInput={(e) => updateButton(index(), 'url', e.currentTarget.value)}
              />
              <div
                class="color-dot color-dot--remove"
                onClick={() => {
                  removeButton(index());
                  window.Telegram?.WebApp?.HapticFeedback?.impactOccurred('light');
                }}
              >
                ✕
              </div>
            </div>
            <div class="btn-editor__color">
              <For each={colorOptions}>
                {(color) => (
                  <div
                    class={`color-dot ${btn.color_code === color.value ? 'color-dot--active' : ''}`}
                    style={{ background: color.bg }}
                    onClick={() => {
                      updateButton(index(), 'color_code', color.value);
                      window.Telegram?.WebApp?.HapticFeedback?.selectionChanged();
                    }}
                    title={color.value}
                  />
                )}
              </For>
            </div>
          </div>
        )}
      </For>

      <button
        class="btn btn--secondary btn--full btn--sm"
        onClick={() => {
          addButton();
          window.Telegram?.WebApp?.HapticFeedback?.impactOccurred('light');
        }}
      >
        ➕ Tugma qo'shish
      </button>
    </div>
  );
};

export default ButtonEditor;
