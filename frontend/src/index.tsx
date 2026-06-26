import { render } from 'solid-js/web';
import App from './App';
import './styles/global.css';

declare global {
  interface Window {
    Telegram: {
      WebApp: {
        ready: () => void;
        expand: () => void;
        close: () => void;
        initData: string;
        initDataUnsafe: any;
        themeParams: Record<string, string>;
        MainButton: {
          text: string;
          color: string;
          textColor: string;
          isVisible: boolean;
          show: () => void;
          hide: () => void;
          onClick: (cb: () => void) => void;
          offClick: (cb: () => void) => void;
          enable: () => void;
          disable: () => void;
          showProgress: (leaveActive?: boolean) => void;
          hideProgress: () => void;
        };
        BackButton: {
          isVisible: boolean;
          show: () => void;
          hide: () => void;
          onClick: (cb: () => void) => void;
          offClick: (cb: () => void) => void;
        };
        HapticFeedback: {
          impactOccurred: (style: string) => void;
          notificationOccurred: (type: string) => void;
          selectionChanged: () => void;
        };
        showAlert: (message: string) => void;
        showConfirm: (message: string, callback: (ok: boolean) => void) => void;
      };
    };
  }
}

const tg = window.Telegram?.WebApp;
if (tg) {
  tg.ready();
  tg.expand();
}

const root = document.getElementById('root');
if (root) {
  render(() => <App />, root);
}
