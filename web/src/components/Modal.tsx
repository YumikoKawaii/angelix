interface ModalProps {
  title: string
  onClose: () => void
  children: React.ReactNode
}

export function Modal({ title, onClose, children }: ModalProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0"
        style={{ background: 'rgba(0,0,0,0.75)' }}
        onClick={onClose}
      />
      <div
        className="relative w-full max-w-md mx-4"
        style={{
          background: 'var(--c-surface)',
          border: '1px solid var(--c-border)',
          boxShadow: '0 0 32px rgba(0,255,225,0.12)',
          padding: '1.5rem',
        }}
      >
        {/* Corner brackets */}
        <span className="absolute top-0 left-0 text-xs leading-none select-none" style={{ color: 'var(--c-cyan)', padding: '3px' }}>┌</span>
        <span className="absolute top-0 right-0 text-xs leading-none select-none" style={{ color: 'var(--c-cyan)', padding: '3px' }}>┐</span>
        <span className="absolute bottom-0 left-0 text-xs leading-none select-none" style={{ color: 'var(--c-cyan)', padding: '3px' }}>└</span>
        <span className="absolute bottom-0 right-0 text-xs leading-none select-none" style={{ color: 'var(--c-cyan)', padding: '3px' }}>┘</span>

        {/* Header */}
        <div
          className="flex items-center justify-between mb-5 pb-3"
          style={{ borderBottom: '1px solid var(--c-border)' }}
        >
          <h2 className="text-xs tracking-widest uppercase" style={{ color: 'var(--c-cyan)' }}>
            // {title}
          </h2>
          <button
            onClick={onClose}
            className="text-xs tracking-widest transition-colors"
            style={{ color: 'var(--c-muted)', background: 'none', border: 'none', cursor: 'pointer' }}
            onMouseEnter={e => (e.currentTarget.style.color = 'var(--c-magenta)')}
            onMouseLeave={e => (e.currentTarget.style.color = 'var(--c-muted)')}
          >
            [X]
          </button>
        </div>

        {children}
      </div>
    </div>
  )
}
