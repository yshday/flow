/**
 * Flow Auth Provider
 * 호스트 앱에서 주입받은 인증 정보를 관리하는 컨텍스트
 */

import { createContext, useContext, useMemo, ReactNode } from 'react'

export interface FlowUser {
  id: number
  email: string
  username: string
  name?: string
  avatar_url?: string
}

export interface FlowCompany {
  id: number
  name: string
  code?: string
}

interface FlowAuthContextValue {
  user: FlowUser | null
  company: FlowCompany | null
  accessToken: string | null
  isAuthenticated: boolean
  getAuthHeader: () => Record<string, string>
}

const FlowAuthContext = createContext<FlowAuthContextValue | null>(null)

export interface FlowAuthProviderProps {
  children: ReactNode
  user: FlowUser | null
  company?: FlowCompany | null
  accessToken: string | null
}

export function FlowAuthProvider({
  children,
  user,
  company = null,
  accessToken,
}: FlowAuthProviderProps) {
  const value = useMemo<FlowAuthContextValue>(
    () => ({
      user,
      company,
      accessToken,
      isAuthenticated: !!accessToken && !!user,
      getAuthHeader: (): Record<string, string> => {
        if (!accessToken) return {}
        return { Authorization: `Bearer ${accessToken}` }
      },
    }),
    [user, company, accessToken]
  )

  return (
    <FlowAuthContext.Provider value={value}>
      {children}
    </FlowAuthContext.Provider>
  )
}

export function useFlowAuth(): FlowAuthContextValue {
  const context = useContext(FlowAuthContext)
  if (!context) {
    throw new Error('useFlowAuth must be used within a FlowAuthProvider')
  }
  return context
}
