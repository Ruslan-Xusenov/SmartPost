import { Component } from 'solid-js';
import { mediaType, setMediaType, setMediaPreview, setFileId } from '../stores/postStore';

const MediaUploader: Component = () => {
  let fileInput: HTMLInputElement | undefined;

  const handleFile = (e: Event) => {
    const target = e.target as HTMLInputElement;
    const file = target.files?.[0];
    if (!file) return;

    // Determine media type from file
    if (file.type.startsWith('image/')) {
      setMediaType('photo');
    } else if (file.type.startsWith('video/')) {
      setMediaType('video');
    }

    // Create preview URL
    const url = URL.createObjectURL(file);
    setMediaPreview(url);

    // In production, we would upload to server and get file_id
    // For now, store the filename as a placeholder
    setFileId(file.name);

    window.Telegram?.WebApp?.HapticFeedback?.impactOccurred('light');
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
            {mediaType() === 'photo' ? '📷' : mediaType() === 'video_note' ? '🔵' : '🎥'}
          </div>
          <div class="upload-zone__text">
            Fayl yuklash uchun bosing
          </div>
        </div>
      )}
    </div>
  );
};

export default MediaUploader;
