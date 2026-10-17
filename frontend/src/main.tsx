import '@fontsource-variable/inter';
import './styles/tokens.css';
import './styles/global.css';
import './styles/print.css';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

const root = document.getElementById('root');
if (!root) throw new Error('#root not found');

createRoot(root).render(
  <StrictMode>
    <p>my-site</p>
  </StrictMode>,
);
