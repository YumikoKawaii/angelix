import { Navigate, Route, Routes } from 'react-router-dom'
import { auth } from './api/client'
import { Layout } from './components/Layout'
import { Login } from './pages/Login'
import { Credentials } from './pages/Credentials'
import { Members } from './pages/Members'
import { Metrics } from './pages/Metrics'

function Guard({ children }: { children: React.ReactNode }) {
  return auth.isLoggedIn() ? <>{children}</> : <Navigate to="/login" replace />
}

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        element={
          <Guard>
            <Layout />
          </Guard>
        }
      >
        <Route index element={<Navigate to="/members" replace />} />
        <Route path="/members" element={<Members />} />
        <Route path="/credentials" element={<Credentials />} />
        <Route path="/metrics" element={<Metrics />} />
      </Route>
    </Routes>
  )
}
