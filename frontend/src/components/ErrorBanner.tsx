import type { ReactNode } from 'react'

export type ErrorBannerProps = {
    message: string
    action?: ReactNode
}

export function ErrorBanner({ message, action }: ErrorBannerProps) {
    return (
        <div className="errorBanner" role="alert">
            <p className="error" style={{ margin: 0 }}>
                {message}
            </p>
            {action && <div className="errorBannerAction">{action}</div>}
        </div>
    )
}