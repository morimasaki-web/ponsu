export type ThemeMode = 'system' | 'light' | 'dark'

const STORAGE_KEY = 'ponsu.theme'

function applyToDom(theme: Exclude<ThemeMode, 'system'> | 'system') {
  // system は data-theme を外して、OS設定に追従させる
  if (theme === 'system') {
    document.documentElement.removeAttribute('data-theme')
    return
  }
  document.documentElement.setAttribute('data-theme', theme)
}

export function getStoredTheme(): ThemeMode {
  const v = localStorage.getItem(STORAGE_KEY)
  if (v === 'light' || v === 'dark' || v === 'system') return v
  return 'system'
}

export function setStoredTheme(theme: ThemeMode) {
  localStorage.setItem(STORAGE_KEY, theme)
}

export function initTheme(): ThemeMode {
  const theme = getStoredTheme()
  applyToDom(theme)

  // system の場合は OS の変更に追従する（data-theme が外れているだけなので特に何もしないが、
  // 将来の拡張用に listener を持っておく）
  const mq = window.matchMedia?.('(prefers-color-scheme: dark)')
  const handler = () => {
    if (getStoredTheme() === 'system') applyToDom('system')
  }

  // 現行ブラウザ前提（ポートフォリオ用デモ）として addEventListener のみ対応
  mq?.addEventListener?.('change', handler)

  return theme
}

export function applyTheme(theme: ThemeMode) {
  setStoredTheme(theme)
  applyToDom(theme)
}
