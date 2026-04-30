import { Component } from 'react';
import type { ErrorInfo, ReactNode } from 'react';
import { Link } from 'react-router-dom';

interface Props { children: ReactNode; }
interface State { hasError: boolean; error: Error | null; }

export class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error('ErrorBoundary caught:', error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            return (
                <div className="min-h-screen flex items-center justify-center px-4">
                    <div className="card max-w-md text-center">
                        <div className="text-5xl mb-4">💥</div>
                        <h2 className="text-xl font-bold mb-2">Something went wrong</h2>
                        <p className="text-text-muted text-sm mb-4">{this.state.error?.message || 'An unexpected error occurred'}</p>
                        <div className="flex gap-2 justify-center">
                            <button onClick={() => this.setState({ hasError: false, error: null })} className="btn-secondary text-sm">Try Again</button>
                            <Link to="/" className="btn-primary text-sm" onClick={() => this.setState({ hasError: false, error: null })}>Go Home</Link>
                        </div>
                    </div>
                </div>
            );
        }
        return this.props.children;
    }
}

export default function NotFoundPage() {
    return (
        <div className="min-h-[60vh] flex items-center justify-center">
            <div className="text-center">
                <div className="text-8xl font-bold bg-gradient-to-r from-accent to-emerald-300 bg-clip-text text-transparent mb-4">404</div>
                <h2 className="text-xl font-bold mb-2">Page Not Found</h2>
                <p className="text-text-muted mb-6">The page you're looking for doesn't exist</p>
                <Link to="/" className="btn-primary">Back to Events</Link>
            </div>
        </div>
    );
}

export function LoadingSpinner({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) {
    const sizes = { sm: 'w-5 h-5', md: 'w-8 h-8', lg: 'w-12 h-12' };
    return (
        <div className="flex items-center justify-center p-8">
            <div className={`${sizes[size]} border-2 border-border-light border-t-accent rounded-full animate-spin`} />
        </div>
    );
}

export function EmptyState({ icon = '📭', title, message }: { icon?: string; title: string; message: string }) {
    return (
        <div className="card text-center py-12">
            <div className="text-4xl mb-3">{icon}</div>
            <p className="text-text-primary text-lg font-medium">{title}</p>
            <p className="text-text-muted text-sm mt-1">{message}</p>
        </div>
    );
}
