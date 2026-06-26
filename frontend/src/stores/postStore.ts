import { createSignal } from 'solid-js';
import type { Channel, Button } from '../api/client';

export const [channels, setChannels] = createSignal<Channel[]>([]);
export const [selectedChannel, setSelectedChannel] = createSignal<Channel | null>(null);
export const [mediaType, setMediaType] = createSignal<string>('text');
export const [fileId, setFileId] = createSignal<string>('');
export const [mediaPreview, setMediaPreview] = createSignal<string>('');
export const [caption, setCaption] = createSignal<string>('');
export const [buttons, setButtons] = createSignal<Button[]>([]);
export const [scheduledAt, setScheduledAt] = createSignal<string>('');
export const [isLoading, setIsLoading] = createSignal(false);

export function addButton() {
  setButtons([...buttons(), {
    text: '',
    url: '',
    color_code: 'default',
    row_index: buttons().length,
  }]);
}

export function updateButton(index: number, field: keyof Button, value: string | number) {
  const updated = [...buttons()];
  (updated[index] as any)[field] = value;
  setButtons(updated);
}

export function removeButton(index: number) {
  setButtons(buttons().filter((_, i) => i !== index));
}

export function resetStore() {
  setSelectedChannel(null);
  setMediaType('text');
  setFileId('');
  setMediaPreview('');
  setCaption('');
  setButtons([]);
  setScheduledAt('');
  setIsLoading(false);
}
