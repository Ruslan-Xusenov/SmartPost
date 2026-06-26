import { Component, createSignal } from 'solid-js';
import Home from './pages/Home';
import CreatePost from './pages/CreatePost';
import PostList from './pages/PostList';

const App: Component = () => {
  const [page, setPage] = createSignal('home');

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