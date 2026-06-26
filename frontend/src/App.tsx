import { Component, createSignal } from 'solid-js';
import Home from './pages/Home';
import CreatePost from './pages/CreatePost';
import PostList from './pages/PostList';

const App: Component = () => {
  const [page, setPage] = createSignal('home');

  const tg = window.Telegram?.WebApp;

  if (!tg?.initData) {
    return (
      <div style={{ padding: "20px", "text-align": "center", color: "var(--tg-theme-text-color, #000)" }}>
        <h2>Xatolik</h2>
        <p>Kechirasiz, ushbu ilovaga faqatgina Telegram orqali kirish mumkin.</p>
      </div>
    );
  }

  const navigate = (p: string) => {
    setPage(p);
    const tg = window.Telegram?.WebApp;
    if (p === 'home') {
      tg?.BackButton?.hide();
    } else {
      tg?.BackButton?.show();
      tg?.BackButton?.onClick(() => setPage('home'));
    }
  };

  return (
    <div class="app-container">
      {page() === 'home' && <Home onNavigate={navigate} />}
      {page() === 'create' && <CreatePost onNavigate={navigate} />}
      {page() === 'posts' && <PostList onNavigate={navigate} />}
    </div>
  );
};

export default App;