/**
 * Flow Provider
 * Flow Issue Tracker의 설정과 상태를 관리하는 메인 컨텍스트
 */

import { createContext, useContext, useMemo, ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { FlowAuthProvider, FlowUser, FlowCompany } from './FlowAuthProvider'
import type { FlowEventCallbacks } from '../types'

export interface FlowConfig {
  /** Flow API 서버 URL */
  apiBaseUrl: string
  /** 디버그 모드 활성화 */
  debug?: boolean
  /** 커스텀 테마 색상 */
  theme?: {
    primary?: string
    secondary?: string
  }
}

interface FlowContextValue {
  config: FlowConfig
  callbacks: FlowEventCallbacks
}

const FlowContext = createContext<FlowContextValue | null>(null)

// 패키지 내부에서 사용할 QueryClient 인스턴스
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1분
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

export interface FlowProviderProps {
  children: ReactNode
  /** Flow API 설정 */
  config: FlowConfig
  /** 호스트 앱에서 주입받은 사용자 정보 */
  user: FlowUser | null
  /** 호스트 앱에서 주입받은 회사 정보 */
  company?: FlowCompany | null
  /** 호스트 앱에서 주입받은 액세스 토큰 */
  accessToken: string | null
  /** 이벤트 콜백 */
  callbacks?: FlowEventCallbacks
}

export function FlowProvider({
  children,
  config,
  user,
  company,
  accessToken,
  callbacks = {},
}: FlowProviderProps) {
  const contextValue = useMemo<FlowContextValue>(
    () => ({
      config,
      callbacks,
    }),
    [config, callbacks]
  )

  return (
    <QueryClientProvider client={queryClient}>
      <FlowAuthProvider user={user} company={company} accessToken={accessToken}>
        <FlowContext.Provider value={contextValue}>
          {children}
        </FlowContext.Provider>
      </FlowAuthProvider>
    </QueryClientProvider>
  )
}

export function useFlowConfig(): FlowConfig {
  const context = useContext(FlowContext)
  if (!context) {
    throw new Error('useFlowConfig must be used within a FlowProvider')
  }
  return context.config
}

export function useFlowCallbacks(): FlowEventCallbacks {
  const context = useContext(FlowContext)
  if (!context) {
    throw new Error('useFlowCallbacks must be used within a FlowProvider')
  }
  return context.callbacks
}
