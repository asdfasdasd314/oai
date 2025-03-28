import './App.css';
import Home from './components/Home';
import GithubSync from './components/GithubSync';
import { Route, Routes, BrowserRouter } from 'react-router-dom';

function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/github-sync" element={<GithubSync />} />
            </Routes>
        </BrowserRouter>
    )
}

export default App
