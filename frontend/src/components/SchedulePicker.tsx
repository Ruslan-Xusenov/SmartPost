import { Component } from 'solid-js';
import { scheduledAt, setScheduledAt } from '../stores/postStore';

const SchedulePicker: Component = () => {
  // Get the current datetime in local format for min attribute
  const now = () => {
    const d = new Date();
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
    return d.toISOString().slice(0, 16);
  };

  return (
    <div>
      <div class="section-title">⏰ Yuborish vaqti</div>
      <div class="schedule-picker">
        <input
          type="datetime-local"
          class="input"
          value={scheduledAt()}
          min={now()}
          onInput={(e) => setScheduledAt(e.currentTarget.value)}
        />
        <button
          class="btn btn--secondary btn--sm"
          onClick={() => {
            setScheduledAt('');
            window.Telegram?.WebApp?.HapticFeedback?.selectionChanged();
          }}
          title="Hozir yuborish"
        >
          🚀
        </button>
      </div>
      <div style="margin-top: 6px; font-size: 12px; color: var(--tg-hint);">
        {scheduledAt()
          ? `📅 Rejalashtirilgan: ${new Date(scheduledAt()).toLocaleString('uz-UZ')}`
          : '💡 Bo\'sh qoldiring = Darhol yuboriladi'
        }
      </div>
    </div>
  );
};

export default SchedulePicker;
