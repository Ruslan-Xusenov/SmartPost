import { Component, createSignal, Show } from 'solid-js';
import { mediaType, setMediaType, setMediaPreview, setFileId } from '../stores/postStore';
import { api } from '../api/client';

const MediaUploader: Component = () => {
  let fileInput: HTMLInputElement | undefined;

  const [isUploading, setIsUploading] = createSignal(false);

  const handleFile = async (e: Event) => {
    const target = e.target as HTMLInputElement;
    const file = target.files?.[0];
    if (!file) return;

    // Determine media type from file if not already explicitly set to a specific video type
    if (file.type.startsWith('image/')) {
      setMediaType('photo');
    } else if (file.type.startsWith('video/') && mediaType() !== 'video_note') {
      setMediaType('video');
    }

    setIsUploading(true);
    try {
      // Upload to server and get file_id by sending a preview to the user
      const res = await api.uploadMedia(file, mediaType());
      setFileId(res.file_id);
      
      // Create preview URL
      const url = URL.createObjectURL(file);
      setMediaPreview(url);
      
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success');
    } catch (err: any) {
      window.Telegram?.WebApp?.showAlert(`Fayl yuklashda xatolik: ${err.message}`);
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('error');
    } finally {
      setIsUploading(false);
    }
  };

  const mediaTypeOptions = [
    { value: 'text', icon: '📝', label: 'Matn' },
    { value: 'photo', icon: '📷', label: 'Rasm' },
    { value: 'video', icon: '🎥', label: 'Video' },
    { value: 'video_note', icon: '🔵', label: 'Dumaloq' },
  ];

  return (
    <div>
      <div class="section-title">📎 Media turi</div>

      <div class="nav-tabs" style="margin-bottom: 12px;">
        {mediaTypeOptions.map(opt => (
          <button
            class={`nav-tab ${mediaType() === opt.value ? 'nav-tab--active' : ''}`}
            onClick={() => {
              setMediaType(opt.value);
              window.Telegram?.WebApp?.HapticFeedback?.selectionChanged();
            }}
          >
            {opt.icon} {opt.label}
          </button>
        ))}
      </div>

      {mediaType() !== 'text' && (
        <div
          class="upload-zone fade-in"
          onClick={() => fileInput?.click()}
        >
          <input
            ref={fileInput}
            type="file"
            accept={
              mediaType() === 'photo' ? 'image/*' :
              mediaType() === 'video' || mediaType() === 'video_note' ? 'video/*' : '*'
            }
            style="display: none"
            onChange={handleFile}
          />
          <div class="upload-zone__icon">
            {isUploading() ? '⏳' : mediaType() === 'photo' ? '📷' : mediaType() === 'video_note' ? '🔵' : '🎥'}
          </div>
          <div class="upload-zone__text">
            {isUploading() ? 'Fayl yuklanmoqda... Kuting' : 'Fayl yuklash uchun bosing'}
          </div>
        </div>
      )}
    </div>
  );
};

export default MediaUploader;
