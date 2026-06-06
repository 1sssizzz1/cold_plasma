import { AnimatePresence } from 'framer-motion'
import { Route, Routes, useLocation } from 'react-router-dom'
import SiteShell from './components/SiteShell'
import HomePage from './pages/HomePage'
import ProceduresPage from './pages/ProceduresPage'
import BookingPage from './pages/BookingPage'
import BeforeAfterPage from './pages/BeforeAfterPage'
import AccountPage from './pages/AccountPage'
import AdminPage from './pages/AdminPage'
import HelpPage from './pages/HelpPage'
import NotFoundPage from './pages/NotFoundPage'
import RequireAdmin from './components/auth/RequireAdmin'

export default function App() {
  const location = useLocation()
  return (
    <SiteShell>
      <AnimatePresence mode="wait">
        <Routes location={location} key={location.pathname}>
          <Route path="/" element={<HomePage />} />
          <Route path="/procedures" element={<ProceduresPage />} />
          <Route path="/booking" element={<BookingPage />} />
          <Route path="/before-after" element={<BeforeAfterPage />} />
          <Route path="/account" element={<AccountPage />} />
          <Route
            path="/admin"
            element={
              <RequireAdmin>
                <AdminPage />
              </RequireAdmin>
            }
          />
          <Route path="/help" element={<HelpPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </AnimatePresence>
    </SiteShell>
  )
}
