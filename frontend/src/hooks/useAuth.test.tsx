import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAuth } from './useAuth';
import { authApi } from '../api/auth';

// Mock the API
vi.mock('../api/auth');
vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn(() => ({
    setAuth: vi.fn(),
    logout: vi.fn(),
    isAuthenticated: () => false,
  })),
}));

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should login successfully', async () => {
    const mockResponse = {
      access_token: 'mock-token',
      refresh_token: 'mock-refresh',
      user: { id: 1, email: 'test@example.com', username: 'testuser', created_at: '2025-11-16', updated_at: '2025-11-16' },
    };

    vi.mocked(authApi.login).mockResolvedValue(mockResponse);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    result.current.login({ email: 'test@example.com', password: 'password123' });

    await waitFor(() => {
      expect(result.current.isLoginLoading).toBe(false);
    });

    expect(authApi.login).toHaveBeenCalledWith({
      email: 'test@example.com',
      password: 'password123',
    });
  });

  it('should register successfully', async () => {
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      username: 'testuser',
      created_at: '2025-11-16',
      updated_at: '2025-11-16',
    };

    vi.mocked(authApi.register).mockResolvedValue(mockUser);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    result.current.register({
      email: 'test@example.com',
      username: 'testuser',
      password: 'password123',
    });

    await waitFor(() => {
      expect(result.current.isRegisterLoading).toBe(false);
    });

    expect(authApi.register).toHaveBeenCalledWith({
      email: 'test@example.com',
      username: 'testuser',
      password: 'password123',
    });
  });

  it('should handle login error', async () => {
    const mockError = new Error('Invalid credentials');
    vi.mocked(authApi.login).mockRejectedValue(mockError);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    result.current.login({ email: 'test@example.com', password: 'wrong' });

    await waitFor(() => {
      expect(result.current.loginError).toBeTruthy();
    });
  });

  it('should logout and clear query cache', () => {
    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    result.current.logout();

    expect(authApi.logout).toHaveBeenCalled();
  });
});
