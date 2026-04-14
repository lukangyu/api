import { Navigate, Route, Routes } from 'react-router-dom'
import { isLoggedIn } from './utils/auth'
import LoginPage from './pages/Login'
import DashboardPage from './pages/Dashboard'
import UpstreamsPage from './pages/Upstreams'
import UsersPage from './pages/Users'
import ApiKeysPage from './pages/ApiKeys'
import LogsPage from './pages/Logs'
import AppLayout from './components/Layout'

function Protected({ children }: { children: JSX.Element }) {
  if (!isLoggedIn()) {
    return <Navigate to="/login" replace />
  }
  return children
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/*"
        element={
          <Protected>
            <AppLayout>
              <Routes>
                <Route path="/" element={<DashboardPage />} />
                <Route path="/upstreams" element={<UpstreamsPage />} />
                <Route path="/users" element={<UsersPage />} />
                <Route path="/api-keys" element={<ApiKeysPage />} />
                <Route path="/logs" element={<LogsPage />} />
              </Routes>
            </AppLayout>
          </Protected>
        }
      />
    </Routes>
  )
}
