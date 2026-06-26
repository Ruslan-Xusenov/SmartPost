import { Component, Show } from 'solid-js';
import { api } from '../api/client';
import ChannelSelector from '../components/ChannelSelector';
import MediaUploader from '../components/MediaUploader';
import PostEditor from '../components/PostEditor';
import ButtonEditor from '../components/ButtonEditor';
import PostPreview from '../components/PostPreview';
import SchedulePicker from '../components/SchedulePicker';
import {
  channels, selectedChannel, mediaType, fileId,
  caption, buttons, scheduledAt, isLoading,
  setIsLoading, resetStore,
} from '../stores/postStore';

interface Props {
  onNavigate: (page: string) => void;
}

const CreatePost: Component<Props> = (props) => {
  const tg = () => window.Telegram?.WebApp;

  const handleSubmit = async () => {
    const ch = selectedChannel();
    if (!ch) {
      tg()?.showAlert('Kanalni tanlang!');
      return;
    }
    if (mediaType() === 'text' && !caption()) {
      tg()?.showAlert('Matn kiriting!');
      return;
    }

    setIsLoading(true);
    try {
      const result = await api.createPost({
        channel_id: ch.id,
        media_type: mediaType(),
        file_id: fileId(),
        caption: caption(),
        buttons: buttons(),
      });

      const postId = result.id;

      if (scheduledAt()) {
        await api.schedulePost(postId, new Date(scheduledAt()).toISOString());
        tg()?.showAlert('✅ Post rejalashtirildi!');
      } else {
        await api.sendPost(postId);
        tg()?.showAlert('✅ Post yuborildi!');
      }

      tg()?.HapticFeedback?.notificationOccurred('success');
      resetStore();
      props.onNavigate('home');
    } catch (err: any) {
      tg()?.showAlert(`❌ Xatolik: ${err.message}`);
      tg()?.HapticFeedback?.notificationOccurred('error');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div class="fade-in">
      <div class="app-header">
        <button
          class="btn btn--secondary btn--sm"
          onClick={() => { resetStore(); props.onNavigate('home'); }}
        >
          ← Orqaga
        </button>
        <div>
          <div class="app-header__title" style="font-size: 18px;">Yangi post</div>
        </div>
      </div>

      {/* Step 1: Channel */}
      <ChannelSelector channels={channels()} />

      {/* Step 2: Media */}
      <Show when={selectedChannel()}>
        <div class="slide-up" style="margin-top: 16px;">
          <MediaUploader />
        </div>
      </Show>

      {/* Step 3: Caption */}
      <Show when={selectedChannel()}>
        <div class="slide-up" style="margin-top: 16px;">
          <PostEditor />
        </div>
      </Show>

      {/* Step 4: Buttons */}
      <Show when={selectedChannel()}>
        <div class="slide-up" style="margin-top: 16px;">
          <ButtonEditor />
        </div>
      </Show>

      {/* Step 5: Schedule */}
      <Show when={selectedChannel()}>
        <div class="slide-up" style="margin-top: 16px;">
          <SchedulePicker />
        </div>
      </Show>

      {/* Preview */}
      <PostPreview />

      {/* Submit */}
      <Show when={selectedChannel()}>
        <div style="margin-top: 20px; padding-bottom: 32px;">
          <button
            class="btn btn--primary btn--full"
            onClick={handleSubmit}
            disabled={isLoading()}
          >
            {isLoading() ? '⏳ Yuborilmoqda...' :
             scheduledAt() ? '📅 Rejalashtirish' : '🚀 Hozir yuborish'}
          </button>
        </div>
      </Show>
    </div>
  );
};

export default CreatePost;
