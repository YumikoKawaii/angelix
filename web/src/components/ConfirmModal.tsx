import { Modal } from './Modal'

interface ConfirmModalProps {
  message: string
  onConfirm: () => void
  onClose: () => void
}

export function ConfirmModal({ message, onConfirm, onClose }: ConfirmModalProps) {
  return (
    <Modal title="CONFIRM ACTION" onClose={onClose}>
      <p className="text-sm mb-6" style={{ color: 'var(--c-text)' }}>{message}</p>
      <div className="flex justify-end gap-3">
        <button onClick={onClose} className="hud-btn">
          [CANCEL]
        </button>
        <button
          onClick={() => { onConfirm(); onClose() }}
          className="hud-btn-danger"
        >
          [CONFIRM]
        </button>
      </div>
    </Modal>
  )
}
