import { Component } from 'solid-js';
import { caption, setCaption } from '../stores/postStore';

const PostEditor: Component = () => {
  return (
    <div>
      <div class="section-title">📝 Matn</div>
      <div class="input-group">
        <textarea
          class="input textarea"
          placeholder="Post matni yoki caption yozing..."
          value={caption()}
          onInput={(e) => setCaption(e.currentTarget.value)}
          rows={4}
        />
        <div style="display: flex; justify-content: space-between; margin-top: 6px;">
          <span style="font-size: 12px; color: var(--tg-hint);">
            HTML formatlar: &lt;b&gt;, &lt;i&gt;, &lt;a&gt;
          </span>
          <span style="font-size: 12px; color: var(--tg-hint);">
            {caption().length} belgi
          </span>
        </div>
      </div>
    </div>
  );
};

export default PostEditor;
